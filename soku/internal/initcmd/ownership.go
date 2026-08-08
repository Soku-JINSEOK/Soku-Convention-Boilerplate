package initcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// OwnershipState is the before or after ownership record shown in a handoff plan.
type OwnershipState struct {
	Owner          string `json:"owner"`
	Class          string `json:"class"`
	ContentMode    string `json:"content_mode,omitempty"`
	BaselineSHA256 string `json:"baseline_sha256,omitempty"`
	LifecycleState string `json:"lifecycle_state"`
}

// HandoffReport is the deterministic plan and result for one ownership handoff.
type HandoffReport struct {
	State                string         `json:"state"`
	Path                 string         `json:"path"`
	ExpectedSHA256       string         `json:"expected_sha256"`
	CurrentSHA256        string         `json:"current_sha256"`
	ManifestSchemaBefore int            `json:"manifest_schema_before"`
	ManifestSchemaAfter  int            `json:"manifest_schema_after"`
	Before               OwnershipState `json:"before"`
	After                OwnershipState `json:"after"`
	Recovery             Recovery       `json:"recovery"`
}

// HandoffOptions contains the complete explicit input for a manifest-only handoff.
type HandoffOptions struct {
	Root           string
	Path           string
	ExpectedSHA256 string
	DryRun         bool
	Yes            bool
	Interactive    bool
	Confirm        func(HandoffReport) (bool, error)
	SokuVersion    string
	ApplyHook      ApplyHook
}

type handoffTarget struct {
	content []byte
	mode    fs.FileMode
	hash    string
}

// HandoffOwnership plans or applies one explicit core-to-project ownership handoff.
func HandoffOwnership(options HandoffOptions) (HandoffReport, error) {
	if options.Root == "" {
		return HandoffReport{}, fail(2, "path.invalid", "target root is required")
	}
	if err := manifest.ValidatePath(options.Path); err != nil {
		return HandoffReport{}, fail(2, "ownership.path.invalid", "%v", err)
	}
	if !sha256Pattern.MatchString(options.ExpectedSHA256) {
		return HandoffReport{}, fail(2, "ownership.hash.invalid", "--expected-sha256 must be a lowercase 64-character SHA-256")
	}
	if err := ensureNoState(options.Root); err != nil {
		return HandoffReport{}, err
	}
	if err := validateRepositoryPaths(options.Root); err != nil {
		return HandoffReport{}, err
	}

	document, err := manifest.NewStore(options.Root).Load()
	if err != nil {
		if errors.Is(err, manifest.ErrNotInitialized) {
			return HandoffReport{}, fail(2, "manifest.missing", "managed state is not initialized; run soku init first")
		}
		if errors.Is(err, manifest.ErrRecoveryRequired) {
			return HandoffReport{}, fail(8, "recovery.required", "manifest recovery is required before ownership handoff")
		}
		var unsupported *manifest.UnsupportedSchemaError
		if errors.As(err, &unsupported) {
			return HandoffReport{}, fail(5, "manifest.incompatible", "%v", err)
		}
		return HandoffReport{}, fail(2, "manifest.invalid", "%v", err)
	}

	index := -1
	for position, file := range document.Files {
		if file.Path == options.Path {
			index = position
			break
		}
		if strings.EqualFold(file.Path, options.Path) {
			return HandoffReport{}, fail(4, "ownership.path.case_mismatch", "path %q does not exactly match recorded path %q", options.Path, file.Path)
		}
	}
	if index < 0 {
		return HandoffReport{}, fail(4, "ownership.path.unmanaged", "path %q is not a recorded managed file", options.Path)
	}
	file := document.Files[index]
	if file.Owner != "core" || file.Class != "core-managed" {
		return HandoffReport{}, fail(4, "ownership.class.ineligible", "path %q is %s owned by %s; only current core-managed files are eligible", options.Path, file.Class, file.Owner)
	}
	if file.LifecycleState != "current" {
		return HandoffReport{}, fail(4, "ownership.state.ineligible", "path %q is %s; only current files are eligible", options.Path, file.LifecycleState)
	}
	target, err := inspectHandoffTarget(options.Root, file)
	if err != nil {
		return HandoffReport{}, err
	}
	if target.hash == file.BaselineSHA256 {
		return HandoffReport{}, fail(4, "ownership.path.clean", "path %q still matches its managed baseline", options.Path)
	}
	if target.hash != options.ExpectedSHA256 {
		return HandoffReport{}, fail(4, "ownership.hash.stale", "path %q current hash is %s, not the explicitly expected hash", options.Path, target.hash)
	}

	report := HandoffReport{
		State: "planned", Path: options.Path, ExpectedSHA256: options.ExpectedSHA256,
		CurrentSHA256: target.hash, ManifestSchemaBefore: document.SchemaVersion,
		ManifestSchemaAfter: manifest.SchemaVersionV3,
		Before: OwnershipState{
			Owner: file.Owner, Class: file.Class, ContentMode: file.ContentMode,
			BaselineSHA256: file.BaselineSHA256, LifecycleState: file.LifecycleState,
		},
		After: OwnershipState{
			Owner: "project", Class: "project-owned", LifecycleState: "unmanaged-expected",
		},
		Recovery: Recovery{Instructions: []string{}},
	}

	next := document
	next.SchemaVersion = manifest.SchemaVersionV3
	next.Files = append([]manifest.File(nil), document.Files...)
	next.Selection.Stacks = append([]string(nil), document.Selection.Stacks...)
	next.Selection.ProjectOwnedOverrides = append([]string(nil), document.Selection.ProjectOwnedOverrides...)
	next.Components = append([]manifest.Component(nil), document.Components...)
	next.Integrations = append([]manifest.Integration(nil), document.Integrations...)
	next.Files[index] = manifest.File{
		Path: options.Path, Owner: "project", Class: "project-owned", LifecycleState: "unmanaged-expected",
	}
	next.Selection.ProjectOwnedOverrides = append(next.Selection.ProjectOwnedOverrides, options.Path)
	sort.Strings(next.Selection.ProjectOwnedOverrides)
	next.Selection.ConfigurationHash, err = manifest.HashSelection(next.Selection)
	if err != nil {
		return HandoffReport{}, fail(2, "ownership.manifest.invalid", "hash manifest selection: %v", err)
	}
	if strings.TrimSpace(options.SokuVersion) != "" {
		next.SokuVersion = options.SokuVersion
	}
	if err := manifest.Validate(next); err != nil {
		return HandoffReport{}, fail(2, "ownership.manifest.invalid", "construct manifest v3: %v", err)
	}
	if options.DryRun {
		report.State = "dry-run"
		return report, nil
	}
	if !options.Yes {
		if !options.Interactive || options.Confirm == nil {
			return HandoffReport{}, fail(2, "confirmation.required", "ownership handoff requires --dry-run, --yes, or interactive confirmation")
		}
		approved, confirmErr := options.Confirm(report)
		if confirmErr != nil {
			return HandoffReport{}, fail(2, "confirmation.failed", "read confirmation: %v", confirmErr)
		}
		if !approved {
			report.State = "cancelled"
			return report, nil
		}
	}

	hook := func(stage, path string) error {
		if options.ApplyHook != nil {
			if err := options.ApplyHook(stage, path); err != nil {
				return err
			}
		}
		if stage == "before-manifest" {
			current, verifyErr := inspectHandoffTarget(options.Root, file)
			if verifyErr != nil {
				return verifyErr
			}
			if current.hash != target.hash || current.mode != target.mode || !bytes.Equal(current.content, target.content) {
				return fmt.Errorf("selected project file changed after the approved plan")
			}
		}
		return nil
	}
	transactionID, err := applyTransaction(options.Root, nil, next, hook)
	if err != nil {
		if failure, ok := err.(*Failure); ok {
			switch failure.Code {
			case 8:
				report.Recovery = Recovery{Required: true, TransactionID: transactionID, Instructions: []string{"preserve .soku/transactions/" + transactionID, "run soku status", "restore only from the recorded manifest backup"}}
			case 7:
				report.Recovery = Recovery{Instructions: []string{"rollback restored the exact previous manifest"}}
			}
			failure.Data = report
		}
		return HandoffReport{}, err
	}
	report.State = "applied"
	return report, nil
}

func inspectHandoffTarget(root string, file manifest.File) (handoffTarget, error) {
	if err := ensureNoSymlink(root, file.Path); err != nil {
		return handoffTarget{}, err
	}
	fullPath := filepath.Join(root, filepath.FromSlash(file.Path))
	info, err := os.Lstat(fullPath)
	if errors.Is(err, fs.ErrNotExist) {
		return handoffTarget{}, fail(4, "ownership.path.missing", "path %q is missing", file.Path)
	}
	if err != nil {
		return handoffTarget{}, fail(4, "ownership.path.unreadable", "inspect path %q: %v", file.Path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return handoffTarget{}, fail(4, "ownership.path.symlink", "path %q is a symbolic link", file.Path)
	}
	if !info.Mode().IsRegular() {
		return handoffTarget{}, fail(4, "ownership.path.type_mismatch", "path %q is not a regular file", file.Path)
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return handoffTarget{}, fail(4, "ownership.path.unreadable", "read path %q: %v", file.Path, err)
	}
	hash, err := manifest.HashContent(content, file.ContentMode)
	if err != nil {
		return handoffTarget{}, fail(4, "ownership.path.unreadable", "hash path %q: %v", file.Path, err)
	}
	return handoffTarget{content: content, mode: info.Mode(), hash: hash}, nil
}

// HumanHandoff renders the exact ownership transition without exposing file content.
func HumanHandoff(report HandoffReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku ownership handoff: %s\n", report.State)
	fmt.Fprintf(&builder, "Path: %s\nCurrent SHA-256: %s\n", report.Path, report.CurrentSHA256)
	fmt.Fprintf(&builder, "Manifest: v%d -> v%d\n", report.ManifestSchemaBefore, report.ManifestSchemaAfter)
	fmt.Fprintf(&builder, "Ownership: %s/%s/%s -> %s/%s/%s\n",
		report.Before.Owner, report.Before.Class, report.Before.LifecycleState,
		report.After.Owner, report.After.Class, report.After.LifecycleState)
	return builder.String()
}
