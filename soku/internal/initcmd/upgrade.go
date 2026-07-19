package initcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

// TransitionOptions contains the complete input for diff and upgrade.
type TransitionOptions struct {
	Root                  string
	ConfigPath            string
	TargetRelease         string
	TargetProfile         string
	DryRun                bool
	Yes                   bool
	Interactive           bool
	Confirm               func(TransitionReport) (bool, error)
	SokuVersion           string
	ApplyHook             ApplyHook
	IntegrationSource     string
	IntegrationRef        string
	IntegrationConfigPath string
	IntegrationFetcher    IntegrationFetcher
}

// TransitionReport is the stable, ordered comparison emitted by diff and upgrade.
type TransitionReport struct {
	State          string                 `json:"state"`
	Source         string                 `json:"source"`
	CurrentRelease string                 `json:"current_release"`
	TargetRelease  string                 `json:"target_release"`
	CurrentCommit  string                 `json:"current_commit"`
	TargetCommit   string                 `json:"target_commit"`
	CurrentProfile string                 `json:"current_profile"`
	TargetProfile  string                 `json:"target_profile"`
	Changes        []Change               `json:"changes"`
	Recovery       Recovery               `json:"recovery"`
	HasChanges     bool                   `json:"has_changes"`
	Integrations   []manifest.Integration `json:"integrations"`
}

// RunTransition plans a diff or applies an upgrade using only manifest state and
// an explicitly selected immutable target release.
func RunTransition(ctx context.Context, options TransitionOptions, fetcher Fetcher, apply bool) (TransitionReport, error) {
	if options.Root == "" {
		return TransitionReport{}, fail(2, "path.invalid", "target root is required")
	}
	if err := ensureNoState(options.Root); err != nil {
		return TransitionReport{}, err
	}
	if !releasePattern.MatchString(options.TargetRelease) {
		return TransitionReport{}, fail(2, "selection.invalid", "--boilerplate-release must be an exact vMAJOR.MINOR.PATCH without prerelease")
	}
	document, err := manifest.NewStore(options.Root).Load()
	if err != nil {
		if errors.Is(err, manifest.ErrNotInitialized) {
			return TransitionReport{}, fail(2, "manifest.missing", "managed state is not initialized; run soku init first")
		}
		if errors.Is(err, manifest.ErrRecoveryRequired) {
			return TransitionReport{}, fail(8, "recovery.required", "manifest recovery is required; run soku status and preserve .soku state")
		}
		var unsupported *manifest.UnsupportedSchemaError
		if errors.As(err, &unsupported) {
			return TransitionReport{}, fail(5, "manifest.incompatible", "%v", err)
		}
		return TransitionReport{}, fail(2, "manifest.invalid", "%v", err)
	}
	targetProfile := options.TargetProfile
	if options.ConfigPath != "" {
		fileConfig, configErr := LoadConfig(options.ConfigPath)
		if configErr != nil {
			return TransitionReport{}, configErr
		}
		if fileConfig.BoilerplateSource != "" && fileConfig.BoilerplateSource != document.Boilerplate.Source {
			return TransitionReport{}, fail(5, "source.change", "changing boilerplate source during an upgrade is not supported")
		}
		if len(fileConfig.Stacks) > 0 && strings.Join(fileConfig.Stacks, "\x00") != strings.Join(document.Selection.Stacks, "\x00") {
			return TransitionReport{}, fail(5, "selection.change", "stack changes require a compatible migration")
		}
		if targetProfile == "" {
			targetProfile = fileConfig.Profile
		}
	}
	if compareRelease(options.TargetRelease, document.Boilerplate.Release) < 0 {
		return TransitionReport{}, fail(5, "release.downgrade", "downgrade from %s to %s is not supported", document.Boilerplate.Release, options.TargetRelease)
	}
	if fetcher == nil {
		fetcher = NewSourceClient()
	}
	base, err := fetcher.Fetch(ctx, document.Boilerplate.Source, document.Boilerplate.Release)
	if err != nil {
		return TransitionReport{}, err
	}
	if base.Source != document.Boilerplate.Source || base.Release != document.Boilerplate.Release || base.ResolvedCommit != document.Boilerplate.ResolvedCommit {
		return TransitionReport{}, fail(5, "source.moved", "recorded release %s no longer resolves to manifest commit %s", document.Boilerplate.Release, document.Boilerplate.ResolvedCommit)
	}
	target := base
	if options.TargetRelease != base.Release {
		target, err = fetcher.Fetch(ctx, document.Boilerplate.Source, options.TargetRelease)
		if err != nil {
			return TransitionReport{}, err
		}
	}
	if target.Source != document.Boilerplate.Source || target.Release != options.TargetRelease || !lowerCommit(target.ResolvedCommit) {
		return TransitionReport{}, fail(6, "source.invalid", "source resolver returned an inconsistent immutable target identity")
	}
	baseConfig, err := configFromManifest(document)
	if err != nil {
		return TransitionReport{}, err
	}
	targetConfig := baseConfig
	if targetProfile != "" {
		if !contains([]string{ProfileBootstrap, ProfileStandard, ProfileScaled}, targetProfile) {
			return TransitionReport{}, fail(2, "selection.invalid", "profile must be bootstrap, standard, or scaled")
		}
		targetConfig.Profile = targetProfile
	}
	baseTree, err := renderSnapshot(base, baseConfig)
	if err != nil {
		return TransitionReport{}, err
	}
	targetTree, err := renderSnapshot(target, targetConfig)
	if err != nil {
		return TransitionReport{}, err
	}
	providerBase, err := providerBaselineChanges(options.Root, document)
	if err != nil {
		return TransitionReport{}, err
	}
	baseTree = append(baseTree, providerBase...)
	targetTree = append(targetTree, providerBase...)
	targetIntegrations := append([]manifest.Integration(nil), document.Integrations...)
	if options.IntegrationSource != "" || options.IntegrationRef != "" || options.IntegrationConfigPath != "" {
		integration, integrationErr := planIntegration(ctx, options.IntegrationSource, options.IntegrationRef, options.IntegrationConfigPath, targetConfig.Profile, options.IntegrationFetcher)
		if integrationErr != nil {
			return TransitionReport{}, integrationErr
		}
		filteredTree := targetTree[:0]
		for _, change := range targetTree {
			if change.Owner != integration.Integration.ID {
				filteredTree = append(filteredTree, change)
			}
		}
		targetTree = append(filteredTree, integration.Changes...)
		replaced := false
		for position := range targetIntegrations {
			if targetIntegrations[position].ID == integration.Integration.ID {
				targetIntegrations[position] = integration.Integration
				replaced = true
			}
		}
		if !replaced {
			targetIntegrations = append(targetIntegrations, integration.Integration)
		}
		sort.Slice(targetIntegrations, func(i, j int) bool { return targetIntegrations[i].ID < targetIntegrations[j].ID })
	}
	if err := validateChangeOwnership(baseTree); err != nil {
		return TransitionReport{}, err
	}
	if err := validateChangeOwnership(targetTree); err != nil {
		return TransitionReport{}, err
	}
	report := TransitionReport{
		State: "planned", Source: document.Boilerplate.Source,
		CurrentRelease: document.Boilerplate.Release, TargetRelease: options.TargetRelease,
		CurrentCommit: document.Boilerplate.ResolvedCommit, TargetCommit: target.ResolvedCommit,
		CurrentProfile: baseConfig.Profile, TargetProfile: targetConfig.Profile,
		Recovery:     Recovery{Instructions: []string{}},
		HasChanges:   document.Boilerplate.Release != options.TargetRelease || document.Boilerplate.ResolvedCommit != target.ResolvedCommit || baseConfig.Profile != targetConfig.Profile,
		Integrations: targetIntegrations,
	}
	report.Changes, err = planTransition(options.Root, document, baseTree, targetTree)
	if err != nil {
		if failure, ok := err.(*Failure); ok {
			failure.Data = report
		}
		return TransitionReport{}, err
	}
	for _, change := range report.Changes {
		if change.Action != "unchanged" {
			report.HasChanges = true
		}
		if change.Action == "conflict" || change.Action == "locally-modified" {
			failure := fail(4, "upgrade.conflict", "managed path %q requires manual resolution (%s)", change.Path, change.Action)
			failure.Data = report
			if apply {
				return TransitionReport{}, failure
			}
		}
	}
	previousIntegrations, _ := json.Marshal(document.Integrations)
	nextIntegrations, _ := json.Marshal(targetIntegrations)
	if !bytes.Equal(previousIntegrations, nextIntegrations) {
		report.HasChanges = true
	}
	if !apply {
		if report.HasChanges {
			report.State = "changes"
		} else {
			report.State = "no-op"
		}
		return report, nil
	}
	if !report.HasChanges {
		report.State = "no-op"
		return report, nil
	}
	if options.DryRun {
		report.State = "dry-run"
		return report, nil
	}
	if !options.Yes {
		if !options.Interactive || options.Confirm == nil {
			return TransitionReport{}, fail(2, "confirmation.required", "mutation requires --yes or interactive confirmation")
		}
		approved, confirmErr := options.Confirm(report)
		if confirmErr != nil {
			return TransitionReport{}, fail(2, "confirmation.failed", "read confirmation: %v", confirmErr)
		}
		if !approved {
			report.State = "cancelled"
			return report, nil
		}
	}
	targetDocument, err := buildTransitionManifestWithIntegrations(options.SokuVersion, document, target, targetConfig, report.Changes, targetIntegrations)
	if err != nil {
		return TransitionReport{}, err
	}
	id, err := applyTransaction(options.Root, report.Changes, targetDocument, options.ApplyHook)
	if err != nil {
		if failure, ok := err.(*Failure); ok {
			switch failure.Code {
			case 8:
				report.Recovery = Recovery{Required: true, TransactionID: id, Instructions: []string{"preserve .soku/transactions/" + id, "run soku status", "restore files only from the recorded backup"}}
			case 7:
				report.Recovery = Recovery{Instructions: []string{"rollback restored the complete previous state"}}
			}
			failure.Data = report
		}
		return TransitionReport{}, err
	}
	report.State = "applied"
	return report, nil
}

func renderSnapshot(snapshot SourceSnapshot, config Config) ([]Change, error) {
	catalog, err := DecodeCatalog(snapshot.Files[CatalogPath])
	if err != nil {
		return nil, err
	}
	return renderProfileCatalog(snapshot, catalog, config)
}

func configFromManifest(document manifest.Document) (Config, error) {
	selection := document.Selection
	config := Config{SchemaVersion: 1, BoilerplateSource: document.Boilerplate.Source, BoilerplateRelease: document.Boilerplate.Release, Profile: selection.Profile, Stacks: append([]string(nil), selection.Stacks...), ProjectName: selection.ProjectName, ModulePath: selection.ModulePath, JavaGroup: selection.JavaGroup, ServiceName: selection.ServiceName}
	hash, err := configHash(config)
	if err != nil {
		return Config{}, err
	}
	if hash != selection.ConfigurationHash {
		return Config{}, fail(5, "selection.incompatible", "manifest selection hash cannot reproduce the recorded configuration")
	}
	return config, nil
}

func planTransition(root string, document manifest.Document, baseTree, targetTree []Change) ([]Change, error) {
	if err := validateRepositoryPaths(root); err != nil {
		return nil, err
	}
	base := changeMap(baseTree)
	target := changeMap(targetTree)
	recorded := map[string]manifest.File{}
	paths := map[string]bool{}
	for _, file := range document.Files {
		recorded[file.Path] = file
		paths[file.Path] = true
	}
	for path := range base {
		paths[path] = true
	}
	for path := range target {
		paths[path] = true
	}
	ordered := make([]string, 0, len(paths))
	for path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	result := make([]Change, 0, len(ordered))
	for _, path := range ordered {
		old, wasRendered := base[path]
		next, willRender := target[path]
		file, wasRecorded := recorded[path]
		if wasRecorded && file.Class == "project-owned" {
			if willRender {
				return nil, fail(4, "ownership.conflict", "target output %q collides with a project-owned path", path)
			}
			continue
		}
		if wasRendered != wasRecorded && (!wasRecorded || file.Class != "project-owned") {
			return nil, fail(5, "manifest.incompatible", "recorded managed paths cannot be reproduced from the pinned release at %q", path)
		}
		if wasRecorded && (!wasRendered || old.Owner != file.Owner || old.Class != file.Class || old.ContentMode != file.ContentMode) {
			return nil, fail(5, "manifest.incompatible", "recorded ownership for %q cannot be reproduced from the pinned release", path)
		}
		if wasRecorded && file.Class == "core-managed" && old.BaselineSHA256 != file.BaselineSHA256 {
			return nil, fail(5, "manifest.incompatible", "recorded baseline for %q cannot be reproduced from the pinned release", path)
		}
		if wasRecorded && willRender && (next.Owner != file.Owner || next.Class != file.Class || next.ContentMode != file.ContentMode) {
			return nil, fail(4, "ownership.conflict", "target changes ownership or content mode for %q", path)
		}
		if err := ensureNoSymlink(root, path); err != nil {
			return nil, err
		}
		current, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if !wasRecorded {
			if !willRender {
				continue
			}
			if readErr == nil {
				return nil, fail(4, "ownership.conflict", "target output %q collides with an existing project-owned path", path)
			}
			if !errors.Is(readErr, fs.ErrNotExist) {
				return nil, fail(4, "path.conflict", "target path %q is not writable: %v", path, readErr)
			}
			next.Action = "added"
			result = append(result, next)
			continue
		}
		if readErr != nil {
			change := old
			change.Action = "conflict"
			change.Content = nil
			result = append(result, change)
			continue
		}
		currentHash, hashErr := manifest.HashContent(current, file.ContentMode)
		if hashErr != nil {
			return nil, fail(4, "path.conflict", "hash current path %q: %v", path, hashErr)
		}
		if !willRender {
			change := old
			change.Content = nil
			if currentHash == file.BaselineSHA256 {
				change.Action = "removed"
			} else {
				change.Action = "conflict"
			}
			result = append(result, change)
			continue
		}
		if file.Class != "mergeable" {
			if currentHash == next.BaselineSHA256 {
				next.Action = "unchanged"
				result = append(result, next)
				continue
			}
			if currentHash == file.BaselineSHA256 {
				next.Action = "updated"
				result = append(result, next)
				continue
			}
			next.Action = "locally-modified"
			next.Content = nil
			result = append(result, next)
			continue
		}
		var merged []byte
		var mergeErr error
		switch path {
		case ".gitignore":
			merged, mergeErr = mergeGitignoreThreeWay(old.Content, current, next.Content)
		case ".editorconfig":
			merged, mergeErr = mergeEditorconfigThreeWay(old.Content, current, next.Content)
		default:
			mergeErr = fail(4, "merge.conflict", "no deterministic merge strategy exists for %q", path)
		}
		if mergeErr != nil {
			next.Action = "conflict"
			next.Content = nil
			result = append(result, next)
			continue
		}
		next.Content = merged
		next.BaselineSHA256, mergeErr = manifest.HashContent(merged, next.ContentMode)
		if mergeErr != nil {
			return nil, mergeErr
		}
		if currentHash == next.BaselineSHA256 && currentHash == file.BaselineSHA256 {
			next.Action = "unchanged"
		} else {
			next.Action = "merged"
		}
		result = append(result, next)
	}
	return result, nil
}

func buildTransitionManifestWithIntegrations(version string, previous manifest.Document, snapshot SourceSnapshot, config Config, changes []Change, integrations []manifest.Integration) (manifest.Document, error) {
	active := make([]Change, 0, len(changes))
	for _, change := range changes {
		if change.Action != "removed" {
			active = append(active, change)
		}
	}
	configurationHash, err := configHash(config)
	if err != nil {
		return manifest.Document{}, err
	}
	document, err := buildManifestWithIntegrations(version, snapshot, config, configurationHash, active, integrations)
	if err != nil {
		return manifest.Document{}, err
	}
	for _, file := range previous.Files {
		if file.Class == "project-owned" {
			document.Files = append(document.Files, file)
		}
	}
	sort.Slice(document.Files, func(i, j int) bool { return document.Files[i].Path < document.Files[j].Path })
	return document, manifest.Validate(document)
}

func changeMap(changes []Change) map[string]Change {
	result := make(map[string]Change, len(changes))
	for _, change := range changes {
		result[change.Path] = change
	}
	return result
}

func validateRepositoryPaths(root string) error {
	seen := map[string]string{}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fail(4, "path.invalid", "inspect target repository: %v", walkErr)
		}
		relative, _ := filepath.Rel(root, path)
		if relative == "." {
			return nil
		}
		relative = filepath.ToSlash(relative)
		if relative == ".git" || relative == ".soku" || strings.HasPrefix(relative, ".git/") || strings.HasPrefix(relative, ".soku/") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if err := manifest.ValidatePath(relative); err != nil {
			return fail(4, "path.invalid", "existing repository path: %v", err)
		}
		folded := strings.ToLower(relative)
		if prior, ok := seen[folded]; ok && prior != relative {
			return fail(4, "path.conflict", "repository paths %q and %q collide by case", prior, relative)
		}
		seen[folded] = relative
		return nil
	})
}

func compareRelease(left, right string) int {
	parse := func(value string) [3]uint64 {
		parts := strings.Split(strings.TrimPrefix(value, "v"), ".")
		var result [3]uint64
		for index := range result {
			result[index], _ = strconv.ParseUint(parts[index], 10, 64)
		}
		return result
	}
	a, b := parse(left), parse(right)
	for index := range a {
		if a[index] < b[index] {
			return -1
		}
		if a[index] > b[index] {
			return 1
		}
	}
	return 0
}

func lowerCommit(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, char := range value {
		if !strings.ContainsRune("0123456789abcdef", char) {
			return false
		}
	}
	return true
}

// HumanTransition renders a deterministic human-readable transition report.
func HumanTransition(command string, report TransitionReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku %s: %s\nSource: %s\nRelease: %s -> %s\nCommit: %s -> %s\nProfile: %s -> %s\n", command, report.State, report.Source, report.CurrentRelease, report.TargetRelease, report.CurrentCommit, report.TargetCommit, report.CurrentProfile, report.TargetProfile)
	if len(report.Changes) > 0 {
		builder.WriteString("Changes:\n")
		for _, change := range report.Changes {
			fmt.Fprintf(&builder, "  %s %s\n", change.Action, change.Path)
		}
	}
	if len(report.Integrations) > 0 {
		builder.WriteString("Integrations:\n")
		for _, integration := range report.Integrations {
			fmt.Fprintf(&builder, "  %s %s\n", integration.LifecycleState, integration.ID)
		}
	}
	return builder.String()
}
