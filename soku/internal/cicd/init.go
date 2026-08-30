package cicd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const (
	// CICDComponentID is the durable component identity for a validation-only
	// caller. It is deliberately distinct from provider and delivery state.
	CICDComponentID = "ci-cd-validation"

	// CICDComponentCatalogVersion identifies the component wire format. The
	// resolved mapping and its immutable binding hashes are appended to this
	// value in the manifest component record.
	CICDComponentCatalogVersion = "1"
)

// InitOptions is the complete non-interactive input for CI/CD installation.
// Exactly one of DryRun and Yes must be true. There is intentionally no
// interactive fallback: a caller must either inspect a plan or explicitly
// approve the already validated plan.
type InitOptions struct {
	Root        string
	ConfigPath  string
	DryRun      bool
	Yes         bool
	SokuVersion string
	ApplyHook   initcmd.ApplyHook
}

// InitError is a stable CI/CD installer error. ExitCode uses the public CLI
// lifecycle values without importing the CLI package and creating a cycle.
type InitError struct {
	Code     string
	Message  string
	ExitCode int
	Cause    error
	Data     any
}

func (e *InitError) Error() string { return e.Message }
func (e *InitError) Unwrap() error { return e.Cause }

func initInvalid(exitCode int, code, format string, args ...any) *InitError {
	return &InitError{ExitCode: exitCode, Code: code, Message: fmt.Sprintf(format, args...)}
}

// InitChange is the redacted, portable change summary emitted by Init. File
// bytes are never retained in a report or manifest outside the transaction.
type InitChange struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	SHA256 string `json:"sha256"`
}

type InitRecovery struct {
	Required      bool     `json:"required"`
	TransactionID string   `json:"transaction_id,omitempty"`
	Instructions  []string `json:"instructions"`
}

// InitReport is the stable output for the transactional CI/CD installer. The
// embedded plan contains only portable repository identity and no filesystem
// path, repository name, timestamp, inventory, or raw configuration.
type InitReport struct {
	State                 string       `json:"state"`
	ConfigPath            string       `json:"config_path"`
	Plan                  Plan         `json:"plan"`
	ComponentID           string       `json:"component_id"`
	ComponentCatalog      string       `json:"component_catalog"`
	Platform              string       `json:"platform"`
	MappingID             string       `json:"mapping_id"`
	AdapterID             string       `json:"adapter_id"`
	AdapterRef            string       `json:"adapter_ref"`
	ImplementationSHA256  string       `json:"implementation_sha256"`
	RendererSHA256        string       `json:"renderer_sha256"`
	OutputPath            string       `json:"output_path"`
	OutputSHA256          string       `json:"output_sha256"`
	ManifestSchemaVersion int          `json:"manifest_schema_version"`
	Changes               []InitChange `json:"changes"`
	Recovery              InitRecovery `json:"recovery"`
}

type preparedInit struct {
	document manifest.Document
	change   initcmd.Change
	state    string
}

// Init computes a fresh plan, validates the trusted renderer and current
// repository state, then re-computes all identities immediately before using
// the shared manifest-last transaction. It never accepts a cached plan.
func Init(options InitOptions) (InitReport, error) {
	if options.DryRun == options.Yes {
		return InitReport{}, initInvalid(2, "ci-cd.init.flags", "soku ci-cd init requires exactly one of --dry-run or --yes")
	}
	root, configPath, configRelative, err := resolveInitPaths(options.Root, options.ConfigPath)
	if err != nil {
		return InitReport{}, err
	}
	if err := initcmd.CheckTransactionState(root); err != nil {
		return InitReport{}, initErrorFromFailure(err, nil)
	}
	initialWorktree, err := worktreeFingerprint(root)
	if err != nil {
		return InitReport{}, initInvalid(2, "ci-cd.init.repository", "read repository worktree state: %v", err)
	}

	plan, err := BuildPlan(root, configPath)
	if err != nil {
		return InitReport{}, initErrorFromPlanning(err)
	}
	if !plan.Installability || plan.AdapterResolution.Status != "resolved" {
		return InitReport{}, initInvalid(4, "ci-cd.init.unsafe", "CI/CD requirements are not installable: %s", plan.AdapterResolution.Reason)
	}
	rendered, err := RenderMapping(root, plan.AdapterResolution.MappingID)
	if err != nil {
		return InitReport{}, initInvalid(4, "ci-cd.init.renderer", "%v", err)
	}
	if err := validateInitOutput(root, rendered.Path); err != nil {
		return InitReport{}, err
	}
	document, err := loadManifestForInit(root)
	if err != nil {
		return InitReport{}, err
	}
	prepared, err := prepareInit(root, document, configRelative, rendered)
	if err != nil {
		return InitReport{}, err
	}
	report := makeInitReport(plan, configRelative, rendered, prepared)
	if prepared.state == "no-op" {
		return report, nil
	}
	if options.DryRun {
		report.State = "dry-run"
		return report, nil
	}

	// The CLI parser enforces this too; keeping the check in the lifecycle
	// package prevents alternate callers from introducing an implicit prompt.
	if !options.Yes {
		return InitReport{}, initInvalid(2, "ci-cd.init.confirmation", "soku ci-cd init requires exactly one of --dry-run or --yes")
	}

	if err := initcmd.CheckTransactionState(root); err != nil {
		return InitReport{}, initErrorFromFailure(err, report)
	}
	currentWorktree, currentPlan, currentRendered, currentDocument, currentPrepared, err := readInitState(root, configPath, configRelative)
	if err != nil {
		return InitReport{}, err
	}
	if err := verifyInitWriteState(root, initialWorktree, plan, rendered, document, prepared, currentWorktree, currentPlan, currentRendered, currentDocument, currentPrepared); err != nil {
		return InitReport{}, err
	}

	guard := func(stage, path string) error {
		if options.ApplyHook != nil {
			if err := options.ApplyHook(stage, path); err != nil {
				return err
			}
		}
		if stage == "before-write" && path == currentRendered.Path {
			writeWorktree, writePlan, writeRendered, writeDocument, writePrepared, err := readInitState(root, configPath, configRelative)
			if err != nil {
				return err
			}
			if err := verifyInitWriteState(root, currentWorktree, currentPlan, currentRendered, currentDocument, currentPrepared, writeWorktree, writePlan, writeRendered, writeDocument, writePrepared); err != nil {
				return err
			}
			if err := ensureOutputAbsent(root, path); err != nil {
				return err
			}
		}
		return nil
	}
	transactionID, err := initcmd.ApplyTransaction(root, []initcmd.Change{currentPrepared.change}, currentPrepared.document, guard)
	if err != nil {
		if failure, ok := asInitFailure(err); ok {
			report.Recovery = recoveryFromFailure(failure, transactionID)
			report.State = "failed"
			return InitReport{}, &InitError{Code: failure.Key, Message: failure.Message, ExitCode: failure.Code, Cause: failure, Data: report}
		}
		return InitReport{}, err
	}
	report.State = "applied"
	report.ManifestSchemaVersion = currentPrepared.document.SchemaVersion
	return report, nil
}

func readInitState(root, configPath, configRelative string) (string, Plan, RenderedMapping, manifest.Document, preparedInit, error) {
	worktree, err := worktreeFingerprint(root)
	if err != nil {
		return "", Plan{}, RenderedMapping{}, manifest.Document{}, preparedInit{}, initInvalid(2, "ci-cd.init.repository", "read repository worktree state before write: %v", err)
	}
	plan, err := BuildPlan(root, configPath)
	if err != nil {
		return "", Plan{}, RenderedMapping{}, manifest.Document{}, preparedInit{}, initErrorFromPlanning(err)
	}
	if !plan.Installability || plan.AdapterResolution.Status != "resolved" {
		return "", Plan{}, RenderedMapping{}, manifest.Document{}, preparedInit{}, initInvalid(4, "ci-cd.init.unsafe", "CI/CD requirements are not installable: %s", plan.AdapterResolution.Reason)
	}
	rendered, err := RenderMapping(root, plan.AdapterResolution.MappingID)
	if err != nil {
		return "", Plan{}, RenderedMapping{}, manifest.Document{}, preparedInit{}, initInvalid(4, "ci-cd.init.renderer", "%v", err)
	}
	document, err := loadManifestForInit(root)
	if err != nil {
		return "", Plan{}, RenderedMapping{}, manifest.Document{}, preparedInit{}, err
	}
	prepared, err := prepareInit(root, document, configRelative, rendered)
	if err != nil {
		return "", Plan{}, RenderedMapping{}, manifest.Document{}, preparedInit{}, err
	}
	return worktree, plan, rendered, document, prepared, nil
}

func verifyInitWriteState(root string, expectedWorktree string, expectedPlan Plan, expectedRendered RenderedMapping, expectedDocument manifest.Document, expectedPrepared preparedInit, observedWorktree string, observedPlan Plan, observedRendered RenderedMapping, observedDocument manifest.Document, observedPrepared preparedInit) error {
	if observedWorktree != expectedWorktree || observedPlan.PlanDigest != expectedPlan.PlanDigest ||
		!reflect.DeepEqual(observedRendered.Mapping, expectedRendered.Mapping) ||
		observedRendered.Path != expectedRendered.Path || !bytes.Equal(observedRendered.Content, expectedRendered.Content) ||
		!sameManifest(expectedDocument, observedDocument) || !samePrepared(expectedPrepared, observedPrepared) {
		return initInvalid(4, "ci-cd.init.stale", "repository, configuration, mapping, or manifest state changed before write")
	}
	if err := validateInitOutput(root, observedRendered.Path); err != nil {
		return err
	}
	return nil
}

func resolveInitPaths(rootValue, configValue string) (string, string, string, error) {
	if strings.TrimSpace(rootValue) == "" {
		return "", "", "", initInvalid(2, "ci-cd.init.repository", "repository root is required")
	}
	root, err := filepath.Abs(rootValue)
	if err != nil {
		return "", "", "", initInvalid(2, "ci-cd.init.repository", "resolve repository root: %v", err)
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return "", "", "", initInvalid(2, "ci-cd.init.repository", "repository root is unavailable")
	}
	if strings.TrimSpace(configValue) == "" {
		return "", "", "", initInvalid(2, "ci-cd.init.config", "decision configuration path is required")
	}
	configPath := configValue
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, filepath.FromSlash(configPath))
	}
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return "", "", "", initInvalid(2, "ci-cd.init.config", "resolve decision configuration: %v", err)
	}
	relative, err := filepath.Rel(root, configPath)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return "", "", "", initInvalid(4, "ci-cd.init.config", "decision configuration must be repository-relative")
	}
	relative = filepath.ToSlash(relative)
	if err := manifest.ValidateComponentPath(relative); err != nil {
		return "", "", "", initInvalid(4, "ci-cd.init.config", "decision configuration path is unsafe: %v", err)
	}
	configInfo, err := os.Lstat(configPath)
	if errors.Is(err, fs.ErrNotExist) || err != nil || configInfo.Mode()&os.ModeSymlink != 0 || !configInfo.Mode().IsRegular() {
		return "", "", "", initInvalid(2, "ci-cd.init.config", "decision configuration is not a readable regular file")
	}
	return root, configPath, relative, nil
}

func worktreeFingerprint(root string) (string, error) {
	command := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=all")
	data, err := command.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "?? .soku/transactions/") {
			continue
		}
		filtered = append(filtered, line)
	}
	return sha256Hex([]byte(strings.Join(filtered, "\n"))), nil
}

func loadManifestForInit(root string) (manifest.Document, error) {
	document, err := manifest.NewStore(root).Load()
	if err == nil {
		return document, nil
	}
	if errors.Is(err, manifest.ErrNotInitialized) {
		return manifest.Document{}, initInvalid(4, "ci-cd.init.manifest-missing", "managed repository state is required; run soku init first")
	}
	if errors.Is(err, manifest.ErrRecoveryRequired) {
		return manifest.Document{}, initInvalid(8, "ci-cd.init.recovery-required", "manifest recovery is required; preserve .soku state before CI/CD installation")
	}
	var unsupported *manifest.UnsupportedSchemaError
	if errors.As(err, &unsupported) {
		return manifest.Document{}, initInvalid(5, "ci-cd.init.manifest-incompatible", "%v", err)
	}
	return manifest.Document{}, initInvalid(5, "ci-cd.init.manifest-invalid", "existing manifest is invalid or incompatible: %v", err)
}

func validateInitOutput(root, relative string) error {
	if err := manifest.ValidateManagedPath(relative); err != nil {
		return initInvalid(4, "ci-cd.init.output", "renderer output path is unsafe: %v", err)
	}
	if err := initcmd.CheckWritePath(root, relative); err != nil {
		return initErrorFromFailure(err, nil)
	}
	return nil
}

func ensureOutputAbsent(root, relative string) error {
	target := filepath.Join(root, filepath.FromSlash(relative))
	info, err := os.Lstat(target)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return initInvalid(4, "ci-cd.init.stale", "recheck renderer output %q: %v", relative, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return initInvalid(4, "ci-cd.init.collision", "renderer output %q is no longer an unmanaged regular path", relative)
	}
	return initInvalid(4, "ci-cd.init.collision", "renderer output %q appeared before write", relative)
}

func prepareInit(root string, document manifest.Document, configPath string, rendered RenderedMapping) (preparedInit, error) {
	expectedComponent := manifest.Component{
		ID:                CICDComponentID,
		CatalogVersion:    componentCatalogVersion(rendered.Mapping),
		ConfigurationPath: configPath,
	}
	for _, component := range document.Components {
		if component.ID != CICDComponentID {
			continue
		}
		if component.ConfigurationPath != configPath {
			return preparedInit{}, initInvalid(4, "ci-cd.init.component-conflict", "CI/CD component is bound to a different configuration path")
		}
		if component.CatalogVersion != expectedComponent.CatalogVersion {
			return preparedInit{}, initInvalid(4, "ci-cd.init.binding-drift", "CI/CD component binding does not match the trusted mapping")
		}
		file, ok := manifestFile(document, rendered.Path)
		if !ok || file.Owner != "core" || file.Class != "core-managed" || file.ContentMode != "text" || file.LifecycleState != "current" {
			return preparedInit{}, initInvalid(5, "ci-cd.init.component-invalid", "CI/CD component caller state is missing or incompatible")
		}
		expectedHash, err := manifest.HashContent(rendered.Content, "text")
		if err != nil || file.BaselineSHA256 != expectedHash {
			return preparedInit{}, initInvalid(5, "ci-cd.init.component-invalid", "CI/CD component caller baseline is stale")
		}
		if err := ensureOutputPresentAndClean(root, rendered.Path, rendered.Content, expectedHash); err != nil {
			return preparedInit{}, err
		}
		return preparedInit{document: document, state: "no-op", change: initcmd.Change{Path: rendered.Path, Action: "unchanged", Owner: "core", Class: "core-managed", ContentMode: "text", BaselineSHA256: expectedHash, Content: append([]byte(nil), rendered.Content...)}}, nil
	}

	if _, exists := manifestFile(document, rendered.Path); exists {
		return preparedInit{}, initInvalid(4, "ci-cd.init.collision", "renderer output %q is already recorded by another lifecycle owner", rendered.Path)
	}
	if err := validateInitOutput(root, rendered.Path); err != nil {
		return preparedInit{}, err
	}
	if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rendered.Path))); err == nil {
		return preparedInit{}, initInvalid(4, "ci-cd.init.collision", "renderer output %q already exists; existing files are not adopted", rendered.Path)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return preparedInit{}, initInvalid(4, "ci-cd.init.collision", "cannot inspect renderer output %q: %v", rendered.Path, err)
	}
	expectedHash, err := manifest.HashContent(rendered.Content, "text")
	if err != nil {
		return preparedInit{}, initInvalid(2, "ci-cd.init.renderer", "hash renderer output: %v", err)
	}
	change := initcmd.Change{
		Path:           rendered.Path,
		Action:         "create",
		Owner:          "core",
		Class:          "core-managed",
		ContentMode:    "text",
		BaselineSHA256: expectedHash,
		Content:        append([]byte(nil), rendered.Content...),
	}
	next := document
	next.Files = append([]manifest.File(nil), document.Files...)
	next.Components = append([]manifest.Component(nil), document.Components...)
	if next.SchemaVersion == manifest.SchemaVersion {
		next.SchemaVersion = manifest.SchemaVersionV2
	}
	next.Files = append(next.Files, manifest.File{
		Path: rendered.Path, Owner: "core", Class: "core-managed", ContentMode: "text",
		BaselineSHA256: expectedHash, LifecycleState: "current",
	})
	next.Components = append(next.Components, expectedComponent)
	sort.Slice(next.Files, func(i, j int) bool { return next.Files[i].Path < next.Files[j].Path })
	sort.Slice(next.Components, func(i, j int) bool { return next.Components[i].ID < next.Components[j].ID })
	if err := manifest.Validate(next); err != nil {
		return preparedInit{}, initInvalid(5, "ci-cd.init.manifest-invalid", "construct manifest: %v", err)
	}
	return preparedInit{document: next, change: change, state: "planned"}, nil
}

func ensureOutputPresentAndClean(root, relative string, expected []byte, expectedHash string) error {
	if err := initcmd.CheckWritePath(root, relative); err != nil {
		return initErrorFromFailure(err, nil)
	}
	path := filepath.Join(root, filepath.FromSlash(relative))
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return initInvalid(4, "ci-cd.init.component-invalid", "CI/CD component caller %q is missing", relative)
	}
	if err != nil || !info.Mode().IsRegular() {
		return initInvalid(4, "ci-cd.init.component-invalid", "CI/CD component caller %q is not a regular file", relative)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return initInvalid(4, "ci-cd.init.component-invalid", "read CI/CD component caller %q: %v", relative, err)
	}
	actual, err := manifest.HashContent(data, "text")
	if err != nil || actual != expectedHash || !bytes.Equal(data, expected) {
		return initInvalid(4, "ci-cd.init.component-drift", "CI/CD component caller %q has local modifications", relative)
	}
	return nil
}

func manifestFile(document manifest.Document, relative string) (manifest.File, bool) {
	for _, file := range document.Files {
		if strings.EqualFold(file.Path, relative) {
			return file, true
		}
	}
	return manifest.File{}, false
}

func componentCatalogVersion(mapping AdapterMapping) string {
	return strings.Join([]string{
		CICDComponentCatalogVersion,
		mapping.MappingID,
		mapping.AdapterRef,
		mapping.ImplementationSHA256,
		mapping.RendererSHA256,
	}, ":")
}

func makeInitReport(plan Plan, configPath string, rendered RenderedMapping, prepared preparedInit) InitReport {
	changeHash := sha256Hex(rendered.Content)
	changes := []InitChange{{Path: rendered.Path, Action: prepared.change.Action, SHA256: changeHash}}
	return InitReport{
		State:                 prepared.state,
		ConfigPath:            configPath,
		Plan:                  plan,
		ComponentID:           CICDComponentID,
		ComponentCatalog:      componentCatalogVersion(rendered.Mapping),
		Platform:              rendered.Mapping.Platform,
		MappingID:             rendered.Mapping.MappingID,
		AdapterID:             rendered.Mapping.AdapterID,
		AdapterRef:            rendered.Mapping.AdapterRef,
		ImplementationSHA256:  rendered.Mapping.ImplementationSHA256,
		RendererSHA256:        rendered.Mapping.RendererSHA256,
		OutputPath:            rendered.Path,
		OutputSHA256:          changeHash,
		ManifestSchemaVersion: prepared.document.SchemaVersion,
		Changes:               changes,
		Recovery:              InitRecovery{Instructions: []string{}},
	}
}

func sameManifest(left, right manifest.Document) bool {
	leftData, leftErr := manifest.MarshalCanonical(left)
	rightData, rightErr := manifest.MarshalCanonical(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftData, rightData)
}

func samePrepared(left, right preparedInit) bool {
	return left.state == right.state && left.change.Path == right.change.Path && left.change.Action == right.change.Action &&
		left.change.Owner == right.change.Owner && left.change.Class == right.change.Class && left.change.ContentMode == right.change.ContentMode &&
		left.change.BaselineSHA256 == right.change.BaselineSHA256 && bytes.Equal(left.change.Content, right.change.Content) && sameManifest(left.document, right.document)
}

func initErrorFromPlanning(err error) error {
	var planning *Error
	if errors.As(err, &planning) {
		return &InitError{Code: planning.Code, Message: planning.Message, ExitCode: 2, Cause: planning}
	}
	return err
}

func initErrorFromFailure(err error, data any) *InitError {
	var failure *initcmd.Failure
	if !errors.As(err, &failure) {
		return initInvalid(8, "ci-cd.init.transaction", "%v", err)
	}
	return &InitError{Code: failure.Key, Message: failure.Message, ExitCode: failure.Code, Cause: failure, Data: data}
}

func asInitFailure(err error) (*initcmd.Failure, bool) {
	var failure *initcmd.Failure
	return failure, errors.As(err, &failure)
}

func recoveryFromFailure(failure *initcmd.Failure, transactionID string) InitRecovery {
	if failure.Code == 8 {
		return InitRecovery{Required: true, TransactionID: transactionID, Instructions: []string{
			"preserve .soku/transactions/" + transactionID,
			"run soku status",
			"restore files only from the recorded backup",
		}}
	}
	return InitRecovery{Instructions: []string{"rollback restored the complete previous state"}}
}

// HumanInit renders the same data model as JSON without exposing raw content
// or machine-local paths.
func HumanInit(report InitReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku CI/CD init: %s\nConfig: %s\nPlatform: %s\nMapping: %s\nOutput: %s\nOutput SHA-256: %s\nManifest schema: %d\n",
		report.State, report.ConfigPath, report.Platform, report.MappingID, report.OutputPath, report.OutputSHA256, report.ManifestSchemaVersion)
	if len(report.Changes) > 0 {
		builder.WriteString("Changes:\n")
		for _, change := range report.Changes {
			fmt.Fprintf(&builder, "  %s %s (%s)\n", change.Action, change.Path, change.SHA256)
		}
	}
	if report.Recovery.Required {
		builder.WriteString("Recovery required:\n")
		for _, instruction := range report.Recovery.Instructions {
			fmt.Fprintf(&builder, "  - %s\n", instruction)
		}
	}
	return builder.String()
}
