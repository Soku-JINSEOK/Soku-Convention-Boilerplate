// Package status performs read-only lifecycle diagnostics from a manifest snapshot.
package status

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

// Report is the stable JSON data payload for soku status.
type Report struct {
	State        string                  `json:"state"`
	Manifest     ManifestDiagnostic      `json:"manifest"`
	Boilerplate  BoilerplateDiagnostic   `json:"boilerplate"`
	Counts       Counts                  `json:"counts"`
	Files        []FileDiagnostic        `json:"files"`
	Integrations []IntegrationDiagnostic `json:"integrations"`
	Guidance     []string                `json:"guidance"`
}

// ManifestDiagnostic describes the portable snapshot used by the check.
type ManifestDiagnostic struct {
	Path          string `json:"path"`
	SchemaVersion int    `json:"schema_version"`
	SokuVersion   string `json:"soku_version"`
}

// BoilerplateDiagnostic identifies the immutable source snapshot.
type BoilerplateDiagnostic struct {
	Source         string `json:"source"`
	Release        string `json:"release"`
	ResolvedCommit string `json:"resolved_commit"`
}

// Counts contains all diagnostic categories in fixed wire order.
type Counts struct {
	Clean              int `json:"clean"`
	Missing            int `json:"missing"`
	Changed            int `json:"changed"`
	Obsolete           int `json:"obsolete"`
	UnmanagedExpected  int `json:"unmanaged_expected"`
	TypeMismatch       int `json:"type_mismatch"`
	Unreadable         int `json:"unreadable"`
	SymlinkEscape      int `json:"symlink_escape"`
	IntegrationPending int `json:"integration_pending"`
	IntegrationDrift   int `json:"integration_drift"`
}

// FileDiagnostic is one path-sorted filesystem result.
type FileDiagnostic struct {
	Path    string `json:"path"`
	Owner   string `json:"owner"`
	Class   string `json:"class"`
	State   string `json:"state"`
	Message string `json:"message"`
}

// IntegrationDiagnostic is one ID-sorted provider snapshot result.
type IntegrationDiagnostic struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	Message string `json:"message"`
}

// Result combines the report with its human rendering and process outcome.
type Result struct {
	Report Report
	Human  string
	Code   int
}

// ValidationError identifies malformed or unreadable lifecycle state.
type ValidationError struct {
	Cause error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("manifest state is invalid or unreadable: %v", e.Cause)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// Inspect loads and diagnoses root without mutating any file.
func Inspect(root string) (Result, error) {
	document, err := manifest.NewStore(root).Load()
	if err != nil {
		return inspectLoadError(err)
	}
	report := Report{
		State: "clean",
		Manifest: ManifestDiagnostic{
			Path:          manifest.ManifestPath,
			SchemaVersion: document.SchemaVersion,
			SokuVersion:   document.SokuVersion,
		},
		Boilerplate: BoilerplateDiagnostic{
			Source:         document.Boilerplate.Source,
			Release:        document.Boilerplate.Release,
			ResolvedCommit: document.Boilerplate.ResolvedCommit,
		},
		Files:        make([]FileDiagnostic, 0, len(document.Files)),
		Integrations: make([]IntegrationDiagnostic, 0, len(document.Integrations)),
		Guidance:     []string{},
	}

	rootPath, err := filepath.Abs(root)
	if err != nil {
		return Result{}, fmt.Errorf("resolve repository root: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return Result{}, fmt.Errorf("resolve repository root: %w", err)
	}
	for _, file := range document.Files {
		diagnostic := inspectFile(resolvedRoot, file)
		report.Files = append(report.Files, diagnostic)
		incrementFileCount(&report.Counts, diagnostic.State)
	}
	for _, integration := range document.Integrations {
		diagnostic := inspectIntegration(integration)
		report.Integrations = append(report.Integrations, diagnostic)
		switch diagnostic.State {
		case "pending":
			report.Counts.IntegrationPending++
		case "drifted":
			report.Counts.IntegrationDrift++
		}
	}
	sort.Slice(report.Files, func(i, j int) bool { return report.Files[i].Path < report.Files[j].Path })
	sort.Slice(report.Integrations, func(i, j int) bool { return report.Integrations[i].ID < report.Integrations[j].ID })

	code := 0
	if report.Counts.Unreadable > 0 {
		return Result{}, &ValidationError{Cause: errors.New("one or more managed files are unreadable")}
	}
	if hasCompatibilityFailure(document) {
		report.State = "incompatible"
		report.Guidance = append(report.Guidance, "Install a compatible soku/provider version before making changes.")
		code = 5
	} else if hasDrift(report.Counts) {
		report.State = "drifted"
		report.Guidance = append(report.Guidance, "Review the diagnostics before a future diff or upgrade.")
		code = 3
	}
	return Result{Report: report, Human: renderHuman(report), Code: code}, nil
}

func inspectLoadError(err error) (Result, error) {
	base := Report{
		Manifest: ManifestDiagnostic{Path: manifest.ManifestPath},
		Files:    []FileDiagnostic{}, Integrations: []IntegrationDiagnostic{}, Guidance: []string{},
	}
	if errors.Is(err, manifest.ErrNotInitialized) {
		base.State = "uninitialized"
		base.Guidance = append(base.Guidance, "Run soku init when initialization becomes available.")
		return Result{Report: base, Human: renderHuman(base), Code: 3}, nil
	}
	if errors.Is(err, manifest.ErrRecoveryRequired) {
		base.State = "recovery-required"
		base.Guidance = append(base.Guidance, "Preserve .soku/manifest.json.pending and run explicit manifest recovery before mutation.")
		return Result{Report: base, Human: renderHuman(base), Code: 3}, nil
	}
	var unsupported *manifest.UnsupportedSchemaError
	if errors.As(err, &unsupported) {
		base.State = "incompatible"
		base.Manifest.SchemaVersion = unsupported.Version
		base.Guidance = append(base.Guidance, "Use a soku version that supports this manifest schema; the repository was not changed.")
		return Result{Report: base, Human: renderHuman(base), Code: 5}, nil
	}
	return Result{}, &ValidationError{Cause: err}
}

func inspectFile(root string, file manifest.File) FileDiagnostic {
	diagnostic := FileDiagnostic{Path: file.Path, Owner: file.Owner, Class: file.Class}
	fullPath := filepath.Join(root, filepath.FromSlash(file.Path))
	info, err := os.Lstat(fullPath)
	if errors.Is(err, os.ErrNotExist) {
		diagnostic.State = "missing"
		diagnostic.Message = "expected path is missing"
		return diagnostic
	}
	if err != nil {
		diagnostic.State = "unreadable"
		diagnostic.Message = "path metadata cannot be read"
		return diagnostic
	}
	resolvedPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		diagnostic.State = "unreadable"
		diagnostic.Message = "path cannot be resolved"
		return diagnostic
	}
	if !withinRoot(root, resolvedPath) {
		diagnostic.State = "symlink-escape"
		diagnostic.Message = "path resolves outside the repository"
		return diagnostic
	}
	if info.Mode()&os.ModeSymlink != 0 {
		info, err = os.Stat(fullPath)
		if err != nil {
			diagnostic.State = "unreadable"
			diagnostic.Message = "symbolic link target cannot be read"
			return diagnostic
		}
	}
	if !info.Mode().IsRegular() {
		diagnostic.State = "type-mismatch"
		diagnostic.Message = "expected a regular file"
		return diagnostic
	}
	if file.LifecycleState == "obsolete" {
		diagnostic.State = "obsolete"
		diagnostic.Message = "obsolete managed path is still present"
		return diagnostic
	}
	if file.LifecycleState == "unmanaged-expected" || file.Class == "project-owned" {
		diagnostic.State = "unmanaged-expected"
		diagnostic.Message = "path is present and intentionally unmanaged"
		return diagnostic
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		diagnostic.State = "unreadable"
		diagnostic.Message = "file content cannot be read"
		return diagnostic
	}
	hash, err := manifest.HashContent(content, file.ContentMode)
	if err != nil {
		diagnostic.State = "unreadable"
		diagnostic.Message = err.Error()
		return diagnostic
	}
	if hash != file.BaselineSHA256 {
		diagnostic.State = "changed"
		diagnostic.Message = "content differs from the recorded baseline"
		return diagnostic
	}
	diagnostic.State = "clean"
	diagnostic.Message = "content matches the recorded baseline"
	return diagnostic
}

func inspectIntegration(integration manifest.Integration) IntegrationDiagnostic {
	diagnostic := IntegrationDiagnostic{ID: integration.ID, State: integration.LifecycleState}
	if integration.ProviderAPIVersion != "1" || integration.ProviderSchemaVersion != "1" {
		diagnostic.State = "incompatible"
		diagnostic.Message = fmt.Sprintf(
			"provider API/schema version %s/%s is not supported",
			integration.ProviderAPIVersion,
			integration.ProviderSchemaVersion,
		)
		return diagnostic
	}
	switch integration.LifecycleState {
	case "connected":
		diagnostic.Message = "integration snapshot is current"
	case "pending":
		diagnostic.Message = "integration is waiting for exact provider data"
	case "drifted":
		diagnostic.Message = "integration state differs from its recorded snapshot"
	case "incompatible":
		diagnostic.Message = "integration is recorded as incompatible"
	}
	return diagnostic
}

func withinRoot(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relative)
}

func hasCompatibilityFailure(document manifest.Document) bool {
	for _, integration := range document.Integrations {
		if integration.LifecycleState == "incompatible" || integration.ProviderAPIVersion != "1" || integration.ProviderSchemaVersion != "1" {
			return true
		}
	}
	return false
}

func hasDrift(counts Counts) bool {
	return counts.Missing+counts.Changed+counts.Obsolete+counts.UnmanagedExpected+
		counts.TypeMismatch+counts.Unreadable+counts.SymlinkEscape+
		counts.IntegrationPending+counts.IntegrationDrift > 0
}

func incrementFileCount(counts *Counts, state string) {
	switch state {
	case "clean":
		counts.Clean++
	case "missing":
		counts.Missing++
	case "changed":
		counts.Changed++
	case "obsolete":
		counts.Obsolete++
	case "unmanaged-expected":
		counts.UnmanagedExpected++
	case "type-mismatch":
		counts.TypeMismatch++
	case "unreadable":
		counts.Unreadable++
	case "symlink-escape":
		counts.SymlinkEscape++
	}
}

func renderHuman(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku status: %s\n", report.State)
	if report.Manifest.SchemaVersion != 0 {
		fmt.Fprintf(&builder, "Manifest: %s (schema %d, soku %s)\n", report.Manifest.Path, report.Manifest.SchemaVersion, report.Manifest.SokuVersion)
	} else {
		fmt.Fprintf(&builder, "Manifest: %s\n", report.Manifest.Path)
	}
	if len(report.Files) > 0 || len(report.Integrations) > 0 {
		fmt.Fprintf(&builder, "Summary: %d clean, %d missing, %d changed, %d obsolete, %d unmanaged expected\n",
			report.Counts.Clean, report.Counts.Missing, report.Counts.Changed, report.Counts.Obsolete, report.Counts.UnmanagedExpected)
	}
	for _, file := range report.Files {
		if file.State != "clean" {
			fmt.Fprintf(&builder, "- %s: %s (%s)\n", file.Path, file.State, file.Message)
		}
	}
	for _, integration := range report.Integrations {
		if integration.State != "connected" {
			fmt.Fprintf(&builder, "- integration %s: %s (%s)\n", integration.ID, integration.State, integration.Message)
		}
	}
	for _, guidance := range report.Guidance {
		fmt.Fprintf(&builder, "Next: %s\n", guidance)
	}
	return builder.String()
}
