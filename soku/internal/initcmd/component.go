package initcmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/projectsync"
)

func projectSyncChanges(root string, number int, document *manifest.Document) ([]Change, error) {
	if err := validateRepositoryPaths(root); err != nil {
		return nil, err
	}
	assets, err := projectsync.Assets(number)
	if err != nil {
		return nil, fail(2, "project-sync.project-number.invalid", "%v", err)
	}
	recorded := map[string]manifest.File{}
	componentInstalled := false
	if document != nil {
		for _, file := range document.Files {
			recorded[strings.ToLower(file.Path)] = file
		}
		for _, component := range document.Components {
			if component.ID != projectsync.ComponentID {
				continue
			}
			componentInstalled = true
			if component.CatalogVersion != projectsync.CatalogVersion || component.ConfigurationPath != projectsync.ConfigPath {
				return nil, fail(4, "project-sync.component.conflict", "github-project-sync is installed with an unsupported catalog or configuration path")
			}
		}
	}

	changes := make([]Change, 0, len(assets))
	for _, asset := range assets {
		recordedFile, recordedOK := recorded[strings.ToLower(asset.Path)]
		if componentInstalled {
			change, err := validateInstalledProjectSyncAsset(root, asset, recordedFile, recordedOK)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
			continue
		}
		if recordedOK {
			return nil, fail(4, "project-sync.component.collision", "component output %q is already recorded by another lifecycle owner", asset.Path)
		}
		if err := ensureNoSymlink(root, asset.Path); err != nil {
			return nil, err
		}
		if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(asset.Path))); err == nil {
			return nil, fail(4, "project-sync.component.collision", "component output %q already exists; existing Project Sync files are not adopted", asset.Path)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, fail(4, "project-sync.component.collision", "cannot inspect component output %q: %v", asset.Path, err)
		}
		change, err := assetChange(asset, "create")
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes, nil
}

func validateInstalledProjectSyncAsset(root string, asset projectsync.Asset, recorded manifest.File, recordedOK bool) (Change, error) {
	if !recordedOK {
		return Change{}, fail(4, "project-sync.component.drift", "component manifest entry %q is missing", asset.Path)
	}
	if recorded.Path != asset.Path {
		return Change{}, fail(4, "project-sync.component.drift", "component manifest path %q is not canonical", recorded.Path)
	}
	if asset.Class == "project-owned" {
		if recorded.Owner != "project" || recorded.Class != "project-owned" || recorded.LifecycleState != "unmanaged-expected" {
			return Change{}, fail(4, "project-sync.component.drift", "project-owned component configuration %q has invalid manifest ownership", asset.Path)
		}
		if err := ensureNoSymlink(root, asset.Path); err != nil {
			return Change{}, err
		}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(asset.Path)))
		if errors.Is(err, fs.ErrNotExist) {
			return Change{}, fail(4, "project-sync.component.drift", "project-owned component configuration %q is missing", asset.Path)
		}
		if err != nil || !info.Mode().IsRegular() {
			return Change{}, fail(4, "project-sync.component.drift", "project-owned component configuration %q is not a regular file", asset.Path)
		}
		return Change{Path: asset.Path, Action: "unchanged", Owner: "project", Class: "project-owned", Content: asset.Content}, nil
	}
	expected, err := manifest.HashContent(asset.Content, asset.ContentMode)
	if err != nil {
		return Change{}, err
	}
	if recorded.Owner != asset.Owner || recorded.Class != asset.Class || recorded.ContentMode != asset.ContentMode || recorded.BaselineSHA256 != expected || recorded.LifecycleState != "current" {
		return Change{}, fail(4, "project-sync.component.drift", "component manifest entry %q does not match catalog v%s", asset.Path, projectsync.CatalogVersion)
	}
	if err := ensureNoSymlink(root, asset.Path); err != nil {
		return Change{}, err
	}
	current, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(asset.Path)))
	if errors.Is(err, fs.ErrNotExist) {
		return Change{}, fail(4, "project-sync.component.drift", "component output %q is missing", asset.Path)
	}
	if err != nil {
		return Change{}, fail(4, "project-sync.component.drift", "cannot read component output %q: %v", asset.Path, err)
	}
	currentHash, err := manifest.HashContent(current, asset.ContentMode)
	if err != nil || currentHash != recorded.BaselineSHA256 {
		return Change{}, fail(4, "project-sync.component.drift", "component output %q has local modifications", asset.Path)
	}
	return Change{Path: asset.Path, Action: "unchanged", Owner: asset.Owner, Class: asset.Class, ContentMode: asset.ContentMode, BaselineSHA256: expected, Content: asset.Content}, nil
}

func assetChange(asset projectsync.Asset, action string) (Change, error) {
	change := Change{Path: asset.Path, Action: action, Owner: asset.Owner, Class: asset.Class, ContentMode: asset.ContentMode, Content: append([]byte(nil), asset.Content...)}
	if asset.Class != "project-owned" {
		hash, err := manifest.HashContent(asset.Content, asset.ContentMode)
		if err != nil {
			return Change{}, fail(2, "project-sync.component.invalid", "hash %q: %v", asset.Path, err)
		}
		change.BaselineSHA256 = hash
	}
	return change, nil
}

func installProjectSync(options Options, document manifest.Document) (Report, error) {
	if options.ProjectSyncProjectNumber < 1 {
		return Report{}, fail(2, "project-sync.project-number.required", "--project-sync-project-number must be a positive integer for non-interactive installation")
	}
	changes, err := projectSyncChanges(options.Root, options.ProjectSyncProjectNumber, &document)
	if err != nil {
		return Report{}, err
	}
	components := append([]manifest.Component(nil), document.Components...)
	if !hasProjectSyncComponent(document) {
		components = append(components, manifest.Component{ID: projectsync.ComponentID, CatalogVersion: projectsync.CatalogVersion, ConfigurationPath: projectsync.ConfigPath})
	}
	sort.Slice(components, func(i, j int) bool { return components[i].ID < components[j].ID })
	report := Report{
		State: "planned", Source: document.Boilerplate.Source, Release: document.Boilerplate.Release,
		ResolvedCommit: document.Boilerplate.ResolvedCommit, Profile: document.Selection.Profile,
		Stacks: append([]string(nil), document.Selection.Stacks...), SelectionHash: "", ConfigurationHash: document.Selection.ConfigurationHash,
		Changes: changes, Verification: []Verification{}, Recovery: Recovery{Instructions: []string{}},
		Integrations: append([]manifest.Integration(nil), document.Integrations...), Components: components,
	}
	if hasProjectSyncComponent(document) {
		report.State = "no-op"
		return report, nil
	}
	if options.DryRun {
		report.State = "dry-run"
		return report, nil
	}
	if !options.Yes {
		if !options.Interactive || options.Confirm == nil {
			return Report{}, fail(2, "confirmation.required", "Project Sync installation requires --dry-run, --yes, or interactive confirmation")
		}
		approved, confirmErr := options.Confirm(report)
		if confirmErr != nil {
			return Report{}, fail(2, "confirmation.failed", "read confirmation: %v", confirmErr)
		}
		if !approved {
			report.State = "cancelled"
			return report, nil
		}
	}
	next := document
	next.Files = append([]manifest.File(nil), document.Files...)
	next.Components = append([]manifest.Component(nil), document.Components...)
	next.SchemaVersion = manifest.SchemaVersionV2
	if options.SokuVersion != "" {
		next.SokuVersion = options.SokuVersion
	}
	for _, change := range changes {
		next.Files = append(next.Files, manifestFileForChange(change))
	}
	next.Components = append(next.Components, manifest.Component{ID: projectsync.ComponentID, CatalogVersion: projectsync.CatalogVersion, ConfigurationPath: projectsync.ConfigPath})
	sort.Slice(next.Files, func(i, j int) bool { return next.Files[i].Path < next.Files[j].Path })
	sort.Slice(next.Components, func(i, j int) bool { return next.Components[i].ID < next.Components[j].ID })
	if err := manifest.Validate(next); err != nil {
		return Report{}, fail(2, "project-sync.manifest.invalid", "construct manifest v2: %v", err)
	}
	transactionID, err := applyTransaction(options.Root, changes, next, options.ApplyHook)
	if err != nil {
		if failure, ok := err.(*Failure); ok {
			switch failure.Code {
			case 8:
				report.Recovery = Recovery{Required: true, TransactionID: transactionID, Instructions: []string{"preserve .soku/transactions/" + transactionID, "run soku status", "restore files only from the recorded backup"}}
			case 7:
				report.Recovery = Recovery{Instructions: []string{"rollback restored the complete previous state"}}
			}
			failure.Data = report
		}
		return Report{}, err
	}
	report.State = "applied"
	report.Components = append([]manifest.Component(nil), next.Components...)
	return report, nil
}

func hasProjectSyncComponent(document manifest.Document) bool {
	for _, component := range document.Components {
		if component.ID == projectsync.ComponentID {
			return true
		}
	}
	return false
}

func manifestFileForChange(change Change) manifest.File {
	if change.Class == "project-owned" {
		return manifest.File{Path: change.Path, Owner: "project", Class: "project-owned", LifecycleState: "unmanaged-expected"}
	}
	return manifest.File{Path: change.Path, Owner: change.Owner, Class: change.Class, ContentMode: change.ContentMode, BaselineSHA256: change.BaselineSHA256, LifecycleState: "current"}
}

func componentBaselineChanges(root string, document manifest.Document) ([]Change, []Change, error) {
	if !hasProjectSyncComponent(document) {
		return nil, nil, nil
	}
	for _, component := range document.Components {
		if component.ID == projectsync.ComponentID &&
			(component.CatalogVersion != projectsync.CatalogVersion || component.ConfigurationPath != projectsync.ConfigPath) {
			return nil, nil, fail(5, "project-sync.component.incompatible", "github-project-sync is recorded with an unsupported catalog or configuration path")
		}
	}
	assets, err := projectsync.ManagedAssets()
	if err != nil {
		return nil, nil, fail(5, "project-sync.component.incompatible", "%v", err)
	}
	recorded := map[string]manifest.File{}
	for _, file := range document.Files {
		recorded[strings.ToLower(file.Path)] = file
	}
	result := make([]Change, 0, len(assets))
	target := make([]Change, 0, len(assets))
	for _, asset := range assets {
		file, ok := recorded[strings.ToLower(asset.Path)]
		if !ok || file.Owner != asset.Owner || file.Class != asset.Class || file.ContentMode != asset.ContentMode || file.LifecycleState != "current" {
			return nil, nil, fail(5, "project-sync.component.incompatible", "component-managed path %q is missing or has incompatible ownership metadata", asset.Path)
		}
		hash, hashErr := manifest.HashContent(asset.Content, asset.ContentMode)
		if hashErr != nil {
			return nil, nil, hashErr
		}
		if err := ensureNoSymlink(root, asset.Path); err != nil {
			return nil, nil, err
		}
		result = append(result, Change{Path: asset.Path, Action: "unchanged", Owner: asset.Owner, Class: asset.Class, ContentMode: asset.ContentMode, BaselineSHA256: file.BaselineSHA256, Content: asset.Content})
		target = append(target, Change{Path: asset.Path, Action: "unchanged", Owner: asset.Owner, Class: asset.Class, ContentMode: asset.ContentMode, BaselineSHA256: hash, Content: asset.Content})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	sort.Slice(target, func(i, j int) bool { return target[i].Path < target[j].Path })
	return result, target, nil
}
