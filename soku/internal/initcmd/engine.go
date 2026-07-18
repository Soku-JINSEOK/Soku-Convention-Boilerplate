package initcmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

type Fetcher interface {
	Fetch(context.Context, string, string) (SourceSnapshot, error)
}

func Run(ctx context.Context, options Options, fetcher Fetcher) (Report, error) {
	if options.Root == "" {
		return Report{}, fail(2, "path.invalid", "target root is required")
	}
	if err := ensureNoState(options.Root); err != nil {
		return Report{}, err
	}
	fileConfig, err := LoadConfig(options.ConfigPath)
	if err != nil {
		return Report{}, err
	}
	source, release := fileConfig.BoilerplateSource, fileConfig.BoilerplateRelease
	if options.Explicit.SourceSet {
		source = options.Explicit.Source
	}
	if options.Explicit.ReleaseSet {
		release = options.Explicit.Release
	}
	if source == "" || release == "" {
		return Report{}, fail(2, "selection.invalid", "--boilerplate-source and --boilerplate-release are required")
	}
	if !releasePattern.MatchString(release) {
		return Report{}, fail(2, "selection.invalid", "boilerplate_release must be an exact vMAJOR.MINOR.PATCH without prerelease")
	}
	if fetcher == nil {
		fetcher = NewSourceClient()
	}
	snapshot, err := fetcher.Fetch(ctx, source, release)
	if err != nil {
		return Report{}, err
	}
	if snapshot.Source != source || snapshot.Release != release || !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(snapshot.ResolvedCommit) {
		return Report{}, fail(6, "source.invalid", "source resolver returned an inconsistent immutable identity")
	}
	catalog, err := DecodeCatalog(snapshot.Files[CatalogPath])
	if err != nil {
		return Report{}, err
	}
	config, err := ResolveConfig(options.Root, fileConfig, options.Explicit, catalog)
	if err != nil {
		return Report{}, err
	}
	configurationHash, err := configHash(config)
	if err != nil {
		return Report{}, err
	}
	selectionHash, err := canonicalHash(struct {
		Source, Release, Commit, Profile, ConfigurationHash string
		Stacks                                              []string
	}{snapshot.Source, snapshot.Release, snapshot.ResolvedCommit, config.Profile, configurationHash, config.Stacks})
	if err != nil {
		return Report{}, err
	}
	report := Report{State: "planned", Source: snapshot.Source, Release: snapshot.Release, ResolvedCommit: snapshot.ResolvedCommit, Profile: config.Profile, Stacks: append([]string(nil), config.Stacks...), SelectionHash: selectionHash, ConfigurationHash: configurationHash, Changes: []Change{}, Verification: []Verification{}, Recovery: Recovery{Instructions: []string{}}}
	if existing, loadErr := manifest.NewStore(options.Root).Load(); loadErr == nil {
		state, rerunErr := checkRerun(options.Root, existing, snapshot, config, configurationHash)
		if rerunErr != nil {
			return Report{}, rerunErr
		}
		if state {
			report.State = "no-op"
			return report, nil
		}
	} else if !errors.Is(loadErr, manifest.ErrNotInitialized) {
		if errors.Is(loadErr, manifest.ErrRecoveryRequired) {
			return Report{}, fail(8, "recovery.required", "manifest recovery is required; run soku status and preserve .soku state")
		}
		return Report{}, fail(4, "manifest.conflict", "existing manifest is invalid or incompatible: %v", loadErr)
	}
	changes, err := renderCatalog(snapshot, catalog, config)
	if err != nil {
		return Report{}, err
	}
	changes, err = preflight(options.Root, changes)
	if err != nil {
		return Report{}, err
	}
	report.Changes = changes
	if config.Verify {
		report.Verification, err = verifyPlan(ctx, options.Root, changes, config.Stacks, nil)
		if err != nil {
			return Report{}, err
		}
	}
	if options.DryRun {
		report.State = "dry-run"
		return report, nil
	}
	if !options.Yes {
		if !options.Interactive || options.Confirm == nil {
			return Report{}, fail(2, "confirmation.required", "mutation requires --yes or interactive confirmation")
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
	document, err := buildManifest(options.SokuVersion, snapshot, config, configurationHash, changes)
	if err != nil {
		return Report{}, err
	}
	transactionID, err := applyTransaction(options.Root, changes, document, nil)
	if err != nil {
		if failure, ok := err.(*Failure); ok {
			switch failure.Code {
			case 8:
				report.Recovery = Recovery{Required: true, TransactionID: transactionID, Instructions: []string{"preserve .soku/transactions/" + transactionID, "run soku status", "restore files only from the recorded backup"}}
			case 7:
				report.Recovery = Recovery{Required: false, Instructions: []string{"rollback restored the complete previous state"}}
			}
			if failure.Code == 7 || failure.Code == 8 {
				failure.Data = report
			}
		}
		return Report{}, err
	}
	report.State = "applied"
	return report, nil
}

func checkRerun(root string, document manifest.Document, snapshot SourceSnapshot, config Config, configurationHash string) (bool, error) {
	if document.Boilerplate.Source != snapshot.Source || document.Boilerplate.Release != snapshot.Release || document.Boilerplate.ResolvedCommit != snapshot.ResolvedCommit || document.Selection.Profile != config.Profile || document.Selection.ConfigurationHash != configurationHash || strings.Join(document.Selection.Stacks, "\x00") != strings.Join(config.Stacks, "\x00") {
		return false, fail(4, "init.conflict", "repository is initialized with a different source or selection; use soku status or soku upgrade")
	}
	for _, file := range document.Files {
		if file.Class == "project-owned" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil {
			return false, fail(4, "init.drift", "managed path %q is missing or unreadable; run soku status", file.Path)
		}
		hash, err := manifest.HashContent(data, file.ContentMode)
		if err != nil || hash != file.BaselineSHA256 {
			return false, fail(4, "init.drift", "managed path %q has drifted; run soku status or soku upgrade", file.Path)
		}
	}
	return true, nil
}

func buildManifest(version string, snapshot SourceSnapshot, config Config, configurationHash string, changes []Change) (manifest.Document, error) {
	if strings.TrimSpace(version) == "" {
		version = "dev"
	}
	files := make([]manifest.File, 0, len(changes))
	for _, change := range changes {
		files = append(files, manifest.File{Path: change.Path, Owner: change.Owner, Class: change.Class, ContentMode: change.ContentMode, BaselineSHA256: change.BaselineSHA256, LifecycleState: "current"})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	document := manifest.Document{SchemaVersion: manifest.SchemaVersion, SokuVersion: version, Boilerplate: manifest.Boilerplate{Source: snapshot.Source, Release: snapshot.Release, ResolvedCommit: snapshot.ResolvedCommit}, Selection: manifest.Selection{Profile: config.Profile, Stacks: append([]string(nil), config.Stacks...), ConfigurationHash: configurationHash}, Files: files, Integrations: []manifest.Integration{}}
	if err := manifest.Validate(document); err != nil {
		return manifest.Document{}, fail(2, "manifest.invalid", "construct manifest: %v", err)
	}
	return document, nil
}

func Human(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Soku init: %s\nSource: %s@%s (%s)\nProfile: %s\nStacks: %s\n", report.State, report.Source, report.Release, report.ResolvedCommit, report.Profile, strings.Join(report.Stacks, ", "))
	if len(report.Changes) > 0 {
		b.WriteString("Changes:\n")
		for _, change := range report.Changes {
			fmt.Fprintf(&b, "  %s %s\n", change.Action, change.Path)
		}
	}
	if len(report.Verification) > 0 {
		b.WriteString("Verification:\n")
		for _, result := range report.Verification {
			fmt.Fprintf(&b, "  %s: %s (%s)\n", result.Stack, strings.Join(result.Command, " "), result.Status)
		}
	}
	return b.String()
}
