package manual

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Check is one deterministic environment diagnosis.
type Check struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DoctorReport contains static checks and optional bounded probe output.
type DoctorReport struct {
	State      string  `json:"state"`
	Probe      bool    `json:"probe"`
	Platform   string  `json:"platform"`
	ConfigPath string  `json:"configuration_path"`
	Checks     []Check `json:"checks"`
}

// Doctor inspects the environment without installation or cloud mutation.
func Doctor(ctx context.Context, root, configPath string, probe bool) (DoctorReport, error) {
	loaded, err := LoadConfig(root, configPath)
	if err != nil {
		return DoctorReport{}, err
	}
	report := DoctorReport{
		State: "ready", Probe: probe, Platform: runtime.GOOS + "/" + runtime.GOARCH,
		ConfigPath: configPath, Checks: []Check{},
	}
	addExecutableCheck(&report, "node", "node", "Install Node.js 22 or newer.")
	addExecutableCheck(&report, "npm", "npm", "Install npm and run npm ci in tools/manual-capture.")
	addPathCheck(&report, root, "runner-lockfile", "tools/manual-capture/package-lock.json", "Run soku docs manual init before installing runner dependencies.")
	addPathCheck(&report, root, "runner-build", "tools/manual-capture/dist/cli.js", "Run npm ci && npm run build in tools/manual-capture.")
	if loaded.Config.PDF != nil {
		addExecutableCheck(&report, "python3", "python3", "Install Python 3 and pypdf for PDF page-count validation.")
		addExecutableCheck(&report, "pdftoppm", "pdftoppm", "Install Poppler for explicit PDF raster review.")
	}
	for _, font := range loaded.Config.Browser.Fonts {
		report.Checks = append(report.Checks, Check{ID: "font:" + font, Status: "review", Message: "The runner will verify representative glyphs for " + font + "; doctor does not install fonts."})
	}
	dirty, dirtyErr := worktreeDirty(ctx, root)
	if dirtyErr != nil {
		report.Checks = append(report.Checks, Check{ID: "worktree", Status: "warn", Message: "Could not inspect Git worktree: " + redact(dirtyErr.Error())})
	} else if dirty {
		report.Checks = append(report.Checks, Check{ID: "worktree", Status: "warn", Message: "Worktree is dirty; capture requires --allow-dirty and records this state."})
	} else {
		report.Checks = append(report.Checks, Check{ID: "worktree", Status: "pass", Message: "Worktree is clean."})
	}
	if loaded.Config.Map.APIKeyEnv != "" {
		status := "pass"
		message := loaded.Config.Map.APIKeyEnv + " is present; its value was not read into the report."
		if os.Getenv(loaded.Config.Map.APIKeyEnv) == "" {
			status = "fail"
			message = loaded.Config.Map.APIKeyEnv + " is not present."
		}
		report.Checks = append(report.Checks, Check{ID: "environment:" + loaded.Config.Map.APIKeyEnv, Status: status, Message: message})
	}
	addOutputChecks(&report, root, loaded.Config)
	if probe {
		if loaded.Config.Execution.Mode != "local-manual" {
			return DoctorReport{}, failure(4, "manual.probe.refused", "probe requires local-manual execution")
		}
		node, lookErr := exec.LookPath("node")
		if lookErr != nil {
			return DoctorReport{}, failure(4, "manual.probe.refused", "probe requires Node.js 22 or newer")
		}
		runner := filepath.Join(root, "tools", "manual-capture", "dist", "cli.js")
		if _, statErr := os.Stat(runner); statErr != nil {
			return DoctorReport{}, failure(4, "manual.probe.refused", "probe runner is not built; run npm ci and npm run build in tools/manual-capture")
		}
		command := exec.CommandContext(ctx, node, runner, "probe", "--config", configPath)
		command.Dir = root
		output, commandErr := command.CombinedOutput()
		if commandErr != nil {
			return DoctorReport{}, failure(4, "manual.probe.failed", "bounded runner probe failed: %s", redact(strings.TrimSpace(string(output))))
		}
		report.Checks = append(report.Checks, Check{ID: "probe", Status: "pass", Message: redact(strings.TrimSpace(string(output)))})
	}
	sort.Slice(report.Checks, func(i, j int) bool { return report.Checks[i].ID < report.Checks[j].ID })
	for _, check := range report.Checks {
		if check.Status == "fail" {
			report.State = "not-ready"
			break
		}
		if check.Status == "warn" || check.Status == "review" {
			report.State = "review-required"
		}
	}
	return report, nil
}

type previousCaptureReport struct {
	GeneratedFiles []struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"generated_files"`
}

func addOutputChecks(report *DoctorReport, root string, config Config) {
	planned := []string{config.Output.GeneratedIndex}
	for _, scenario := range config.Scenarios {
		for _, step := range scenario.Steps {
			if step.Action == "capture" {
				planned = append(planned, strings.TrimSuffix(config.Output.Directory, "/")+"/"+step.ID+".png")
			}
		}
	}
	reportPath := filepath.Join(root, filepath.FromSlash(config.Output.Report))
	reportData, reportErr := os.ReadFile(reportPath)
	if errors.Is(reportErr, os.ErrNotExist) {
		for _, output := range planned {
			if _, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(output))); statErr == nil {
				report.Checks = append(report.Checks, Check{
					ID: "output:" + output, Status: "fail",
					Message: "Output exists without a prior capture report and is not owned by the runner.",
				})
			}
		}
		return
	}
	if reportErr != nil {
		report.Checks = append(report.Checks, Check{
			ID: "output:" + config.Output.Report, Status: "fail",
			Message: "The prior capture report cannot be read.",
		})
		return
	}
	var previous previousCaptureReport
	if json.Unmarshal(reportData, &previous) != nil {
		report.Checks = append(report.Checks, Check{
			ID: "output:" + config.Output.Report, Status: "fail",
			Message: "The prior capture report is not valid JSON.",
		})
		return
	}
	owned := map[string]bool{}
	for _, item := range previous.GeneratedFiles {
		if manifestPathErr := validateProjectPath("generated_files.path", item.Path); manifestPathErr != nil || owned[item.Path] {
			report.Checks = append(report.Checks, Check{
				ID: "output:" + config.Output.Report, Status: "fail",
				Message: "The prior capture report has invalid or duplicate generated paths.",
			})
			return
		}
		owned[item.Path] = true
		data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Path)))
		sum, decodeErr := hex.DecodeString(item.SHA256)
		actual := sha256.Sum256(data)
		if readErr != nil || decodeErr != nil || len(sum) != sha256.Size || !equalBytes(actual[:], sum) {
			report.Checks = append(report.Checks, Check{
				ID: "output:" + item.Path, Status: "fail",
				Message: "A previously generated output is missing or modified.",
			})
		}
	}
	for _, output := range planned {
		if owned[output] {
			continue
		}
		if _, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(output))); statErr == nil {
			report.Checks = append(report.Checks, Check{
				ID: "output:" + output, Status: "fail",
				Message: "A planned output exists but is not owned by the prior report.",
			})
		}
	}
	report.Checks = append(report.Checks, Check{
		ID: "output:" + config.Output.Report, Status: "review",
		Message: "The runner will validate the prior report schema and integrity hash before replacement.",
	})
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var difference byte
	for index := range left {
		difference |= left[index] ^ right[index]
	}
	return difference == 0
}

func addExecutableCheck(report *DoctorReport, id, name, guidance string) {
	value, err := exec.LookPath(name)
	if err != nil {
		report.Checks = append(report.Checks, Check{ID: id, Status: "fail", Message: guidance})
		return
	}
	message := name + " is available."
	if name == "node" {
		command := exec.Command(value, "--version")
		output, runErr := command.Output()
		if runErr != nil {
			report.Checks = append(report.Checks, Check{ID: id, Status: "fail", Message: "Could not determine Node.js version."})
			return
		}
		majorText := strings.TrimPrefix(strings.SplitN(strings.TrimSpace(string(output)), ".", 2)[0], "v")
		major, parseErr := strconv.Atoi(majorText)
		if parseErr != nil || major < 22 {
			report.Checks = append(report.Checks, Check{ID: id, Status: "fail", Message: "Node.js 22 or newer is required."})
			return
		}
		message = "Node.js " + strings.TrimSpace(string(output)) + " is available."
	}
	report.Checks = append(report.Checks, Check{ID: id, Status: "pass", Message: message})
}

func addPathCheck(report *DoctorReport, root, id, value, guidance string) {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(value)))
	if err != nil || !info.Mode().IsRegular() {
		report.Checks = append(report.Checks, Check{ID: id, Status: "fail", Message: guidance})
		return
	}
	report.Checks = append(report.Checks, Check{ID: id, Status: "pass", Message: value + " is present."})
}

func worktreeDirty(ctx context.Context, root string) (bool, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return false, err
	}
	command := exec.CommandContext(ctx, git, "status", "--porcelain=v1", "--untracked-files=normal")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return false, err
	}
	return len(strings.TrimSpace(string(output))) != 0, nil
}

// HumanDoctor renders concise review guidance.
func HumanDoctor(report DoctorReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku docs manual doctor: %s\nPlatform: %s\n", report.State, report.Platform)
	for _, check := range report.Checks {
		fmt.Fprintf(&builder, "  [%s] %s: %s\n", check.Status, check.ID, check.Message)
	}
	return builder.String()
}
