package cicd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const initDecision = `schema_version: 1
mode: ci-only
source_host: github.com
workload: library
artifact: package
required_os:
  - linux
network_scope: public
cloud_authority: none
operations_owner: absent
capabilities: []
verification:
  pr: ci-quick
  full: full
delivery:
  enabled: false
`

const initSpecializedDecision = `schema_version: 1
mode: ci-only
source_host: github.com
workload: desktop-app
artifact: installer
required_os:
  - darwin
network_scope: private-vpc
cloud_authority: none
operations_owner: declared
capabilities:
  - hardware
  - native
verification:
  pr: ci-quick
  full: full
delivery:
  enabled: false
`

func TestInitDryRunApplyRerunAndManifestVersionCompatibility(t *testing.T) {
	for _, version := range []int{manifest.SchemaVersion, manifest.SchemaVersionV2, manifest.SchemaVersionV3} {
		t.Run(fmt.Sprintf("manifest-v%d", version), func(t *testing.T) {
			fixture := newInitFixture(t, version, initDecision)
			beforeManifest := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath))
			if _, err := Init(InitOptions{Root: fixture.root, ConfigPath: fixture.config, DryRun: true}); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(fixture.root, ".github", "workflows", "soku-ci.yml")); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("dry-run output exists: %v", err)
			}
			afterManifest := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath))
			if !bytes.Equal(beforeManifest, afterManifest) {
				t.Fatal("dry-run changed the manifest")
			}
			if _, err := os.Stat(filepath.Join(fixture.root, ".soku", "transactions")); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("dry-run created transaction state: %v", err)
			}

			report, err := Init(InitOptions{Root: fixture.root, ConfigPath: fixture.config, Yes: true, SokuVersion: "v1.2.3"})
			if err != nil || report.State != "applied" {
				t.Fatalf("apply report=%#v err=%v", report, err)
			}
			document, err := manifest.NewStore(fixture.root).Load()
			if err != nil {
				t.Fatal(err)
			}
			wantVersion := version
			if wantVersion == manifest.SchemaVersion {
				wantVersion = manifest.SchemaVersionV2
			}
			if document.SchemaVersion != wantVersion || len(document.Components) != 1 {
				t.Fatalf("document schema/components = %d/%#v", document.SchemaVersion, document.Components)
			}
			if document.Components[0].ID != CICDComponentID || document.Components[0].ConfigurationPath != ".soku/ci-cd-decision.yml" {
				t.Fatalf("component=%#v", document.Components[0])
			}
			file, ok := initManifestFile(document, ".github/workflows/soku-ci.yml")
			if !ok || file.Owner != "core" || file.Class != "core-managed" || file.ContentMode != "text" || file.LifecycleState != "current" {
				t.Fatalf("caller manifest file=%#v exists=%t", file, ok)
			}
			caller := mustReadInitFile(t, filepath.Join(fixture.root, ".github", "workflows", "soku-ci.yml"))
			callerHash, _ := manifest.HashContent(caller, "text")
			if file.BaselineSHA256 != callerHash || report.OutputSHA256 != sha256Hex(caller) {
				t.Fatalf("caller hashes manifest=%s report=%s normalized=%s", file.BaselineSHA256, report.OutputSHA256, callerHash)
			}

			rerun, err := Init(InitOptions{Root: fixture.root, ConfigPath: fixture.config, Yes: true})
			if err != nil || rerun.State != "no-op" {
				t.Fatalf("rerun report=%#v err=%v", rerun, err)
			}
		})
	}
}

func TestInitInstallsTheValidationOnlySelfHostedCaller(t *testing.T) {
	fixture := newInitFixture(t, manifest.SchemaVersionV2, initSpecializedDecision)
	report, err := Init(InitOptions{Root: fixture.root, ConfigPath: fixture.config, Yes: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.Platform != platformGitHubSelfHosted || report.OutputPath != ".soku/ci-cd/github-self-hosted-validation.yml" {
		t.Fatalf("report=%#v", report)
	}
	document, err := manifest.NewStore(fixture.root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := initManifestFile(document, report.OutputPath); !ok {
		t.Fatalf("self-hosted caller is not recorded: %#v", document.Files)
	}
	if _, err := os.Stat(filepath.Join(fixture.root, filepath.FromSlash(report.OutputPath))); err != nil {
		t.Fatal(err)
	}
}

func TestInitRejectsUnmanagedCollisionAndDoesNotMutateManifest(t *testing.T) {
	fixture := newInitFixture(t, manifest.SchemaVersionV2, initDecision)
	output := filepath.Join(fixture.root, ".github", "workflows", "soku-ci.yml")
	writeInitFile(t, output, "project-owned\n", 0o600)
	before := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath))
	_, err := Init(InitOptions{Root: fixture.root, ConfigPath: fixture.config, DryRun: true})
	var initError *InitError
	if !errors.As(err, &initError) || initError.Code != "ci-cd.init.collision" || initError.ExitCode != 4 {
		t.Fatalf("error=%T %v", err, err)
	}
	if after := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath)); !bytes.Equal(before, after) {
		t.Fatal("collision changed the manifest")
	}
	if got := string(mustReadInitFile(t, output)); got != "project-owned\n" {
		t.Fatalf("collision changed existing output: %q", got)
	}
}

func TestInitRollbackRestoresFilesAndManifest(t *testing.T) {
	fixture := newInitFixture(t, manifest.SchemaVersionV2, initDecision)
	before := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath))
	_, err := Init(InitOptions{
		Root: fixture.root, ConfigPath: fixture.config, Yes: true,
		ApplyHook: func(stage, _ string) error {
			if stage == "before-manifest" {
				return errors.New("injected manifest failure")
			}
			return nil
		},
	})
	var initError *InitError
	if !errors.As(err, &initError) || initError.Code != "apply.rolled_back" || initError.ExitCode != 7 {
		t.Fatalf("error=%T %v", err, err)
	}
	if after := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath)); !bytes.Equal(before, after) {
		t.Fatal("rollback changed the manifest")
	}
	if _, err := os.Stat(filepath.Join(fixture.root, ".github", "workflows", "soku-ci.yml")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("rollback left caller: %v", err)
	}
	if _, err := os.Stat(filepath.Join(fixture.root, ".soku", "transactions")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("rollback left transaction state: %v", err)
	}
}

func TestInitRejectsWriteBoundaryDriftAndRollsBack(t *testing.T) {
	fixture := newInitFixture(t, manifest.SchemaVersionV2, initDecision)
	before := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath))
	_, err := Init(InitOptions{
		Root: fixture.root, ConfigPath: fixture.config, Yes: true,
		ApplyHook: func(stage, _ string) error {
			if stage == "before-write" {
				writeInitFile(t, fixture.config, strings.Replace(initDecision, "artifact: package", "artifact: archive", 1), 0o600)
			}
			return nil
		},
	})
	var initError *InitError
	if !errors.As(err, &initError) || initError.Code != "apply.rolled_back" || initError.ExitCode != 7 {
		t.Fatalf("drift error=%T %v", err, err)
	}
	if after := mustReadInitFile(t, filepath.Join(fixture.root, manifest.ManifestPath)); !bytes.Equal(before, after) {
		t.Fatal("write-boundary drift changed the manifest")
	}
	if _, err := os.Stat(filepath.Join(fixture.root, ".github", "workflows", "soku-ci.yml")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("write-boundary drift left caller: %v", err)
	}
}

func TestInitRejectsUnsupportedPlanAndExactFlagContract(t *testing.T) {
	fixture := newInitFixture(t, manifest.SchemaVersionV2, initDecision)
	if _, err := Init(InitOptions{Root: fixture.root, ConfigPath: fixture.config}); !hasInitError(err, "ci-cd.init.flags", 2) {
		t.Fatalf("missing mode error=%v", err)
	}
	if _, err := Init(InitOptions{Root: fixture.root, ConfigPath: fixture.config, DryRun: true, Yes: true}); !hasInitError(err, "ci-cd.init.flags", 2) {
		t.Fatalf("both modes error=%v", err)
	}
	unsupported := newInitFixture(t, manifest.SchemaVersionV2, strings.Replace(initDecision, "cloud_authority: none", "cloud_authority: aws", 1))
	if _, err := Init(InitOptions{Root: unsupported.root, ConfigPath: unsupported.config, DryRun: true}); !hasInitError(err, "ci-cd.init.unsafe", 4) {
		t.Fatalf("unsupported plan error=%v", err)
	}
}

type initFixture struct {
	root   string
	config string
}

func newInitFixture(t *testing.T, version int, decision string) initFixture {
	t.Helper()
	root := t.TempDir()
	sourceRoot := repositoryRoot(t)
	for _, relative := range []string{
		"verification/profiles.yml",
		"soku/catalog/index-v2.json",
		"soku/catalog/ci-cd-adapter-mapping-v1.json",
	} {
		copyInitFile(t, filepath.Join(sourceRoot, relative), filepath.Join(root, relative))
	}
	copyInitTree(t, filepath.Join(sourceRoot, "soku", "internal", "cicd", "testdata", "upstream"), filepath.Join(root, "soku", "internal", "cicd", "testdata", "upstream"))
	runInitGit(t, root, "init", "-q", "-b", "main")
	runInitGit(t, root, "config", "user.email", "test@example.invalid")
	runInitGit(t, root, "config", "user.name", "Soku test")
	runInitGit(t, root, "config", "commit.gpgsign", "false")
	runInitGit(t, root, "remote", "add", "origin", "https://github.com/example/repository")
	runInitGit(t, root, "add", ".")
	runInitGit(t, root, "commit", "-qm", "fixture")

	config := filepath.Join(root, ".soku", "ci-cd-decision.yml")
	writeInitFile(t, config, decision, 0o600)
	document := manifest.Document{
		SchemaVersion: version,
		SokuVersion:   "v1.0.0",
		Boilerplate: manifest.Boilerplate{
			Source: "https://github.com/example/boilerplate", Release: "v1.0.0",
			ResolvedCommit: "0123456789abcdef0123456789abcdef01234567",
		},
		Selection:    manifest.Selection{Profile: "standard", Stacks: []string{}},
		Files:        []manifest.File{},
		Integrations: []manifest.Integration{},
	}
	if version == manifest.SchemaVersionV3 {
		writeInitFile(t, filepath.Join(root, "README.md"), "project\n", 0o644)
		document.Selection.ProjectOwnedOverrides = []string{"README.md"}
		document.Files = append(document.Files, manifest.File{Path: "README.md", Owner: "project", Class: "project-owned", LifecycleState: "unmanaged-expected"})
	}
	document.Selection.ConfigurationHash, _ = manifest.HashSelection(document.Selection)
	if err := manifest.NewStore(root).Write(document); err != nil {
		t.Fatal(err)
	}
	return initFixture{root: root, config: config}
}

func copyInitTree(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		copyInitFile(t, path, target)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func copyInitFile(t *testing.T, source, destination string) {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	writeInitFile(t, destination, string(data), 0o644)
}

func writeInitFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func mustReadInitFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func runInitGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func initManifestFile(document manifest.Document, path string) (manifest.File, bool) {
	for _, file := range document.Files {
		if file.Path == path {
			return file, true
		}
	}
	return manifest.File{}, false
}

func hasInitError(err error, code string, exitCode int) bool {
	var initError *InitError
	return errors.As(err, &initError) && initError.Code == code && initError.ExitCode == exitCode
}
