package manual

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const (
	componentID    = "docs-manual"
	catalogVersion = "1"
	componentOwner = "core"
	componentClass = "core-managed"
	componentMode  = "text"
)

//go:embed assets/runner/package.json assets/runner/package-lock.json assets/runner/tsconfig.json assets/runner/src/*.ts assets/schemas/*.json assets/catalogs/*.json assets/examples/*.yml assets/templates/*.md
var componentAssets embed.FS

// InitChange is one component installation action.
type InitChange struct {
	Path           string `json:"path"`
	Action         string `json:"action"`
	Class          string `json:"class"`
	BaselineSHA256 string `json:"baseline_sha256"`
	content        []byte
}

// InitReport is the deterministic docs-manual installation plan/result.
type InitReport struct {
	State                string       `json:"state"`
	ComponentID          string       `json:"component_id"`
	CatalogVersion       string       `json:"catalog_version"`
	ConfigurationPath    string       `json:"configuration_path"`
	ManifestSchemaBefore int          `json:"manifest_schema_before"`
	ManifestSchemaAfter  int          `json:"manifest_schema_after"`
	Changes              []InitChange `json:"changes"`
	Recovery             []string     `json:"recovery"`
}

// InitOptions controls the explicit component mutation.
type InitOptions struct {
	Root        string
	ConfigPath  string
	DryRun      bool
	Yes         bool
	SokuVersion string
	ApplyHook   func(stage, path string) error
}

// Init plans or transactionally installs the docs-manual component.
func Init(options InitOptions) (InitReport, error) {
	if options.Root == "" {
		return InitReport{}, failure(2, "manual.path.invalid", "target root is required")
	}
	if err := manifest.ValidatePath(options.ConfigPath); err != nil {
		return InitReport{}, failure(2, "manual.configuration.invalid", "config path is not portable: %v", err)
	}
	store := manifest.NewStore(options.Root)
	document, err := store.Load()
	if err != nil {
		if errors.Is(err, manifest.ErrNotInitialized) {
			return InitReport{}, failure(2, "manual.manifest.missing", "managed state is not initialized; run soku init first")
		}
		if errors.Is(err, manifest.ErrRecoveryRequired) {
			return InitReport{}, failure(8, "manual.recovery.required", "manifest recovery is required before component installation")
		}
		var unsupported *manifest.UnsupportedSchemaError
		if errors.As(err, &unsupported) {
			return InitReport{}, failure(5, "manual.manifest.incompatible", "%v", err)
		}
		return InitReport{}, failure(2, "manual.manifest.invalid", "%v", err)
	}
	report := InitReport{
		State: "planned", ComponentID: componentID, CatalogVersion: catalogVersion,
		ConfigurationPath: options.ConfigPath, ManifestSchemaBefore: document.SchemaVersion,
		ManifestSchemaAfter: manifest.SchemaVersionV2, Changes: []InitChange{},
		Recovery: []string{},
	}
	existingComponent := -1
	for index, component := range document.Components {
		if component.ID == componentID {
			existingComponent = index
			if component.CatalogVersion != catalogVersion || component.ConfigurationPath != options.ConfigPath {
				return InitReport{}, failure(4, "manual.component.conflict", "docs-manual is installed with a different catalog or configuration path")
			}
		}
	}
	assets, err := installationAssets()
	if err != nil {
		return InitReport{}, err
	}
	filesByPath := map[string]manifest.File{}
	for _, file := range document.Files {
		filesByPath[strings.ToLower(file.Path)] = file
	}
	for output, content := range assets {
		hash, hashErr := manifest.HashContent(content, componentMode)
		if hashErr != nil {
			return InitReport{}, hashErr
		}
		recorded, managed := filesByPath[strings.ToLower(output)]
		fullPath := filepath.Join(options.Root, filepath.FromSlash(output))
		current, readErr := os.ReadFile(fullPath)
		exists := readErr == nil
		if readErr != nil && !errors.Is(readErr, fs.ErrNotExist) {
			return InitReport{}, failure(4, "manual.component.collision", "cannot inspect %q: %v", output, readErr)
		}
		if existingComponent < 0 {
			if managed || exists {
				return InitReport{}, failure(4, "manual.component.collision", "component output %q is already owned or exists", output)
			}
			report.Changes = append(report.Changes, InitChange{Path: output, Action: "created", Class: componentClass, BaselineSHA256: hash, content: content})
			continue
		}
		if !managed || recorded.Owner != componentOwner || recorded.Class != componentClass || recorded.BaselineSHA256 != hash {
			return InitReport{}, failure(4, "manual.component.drift", "component manifest entry %q is missing or does not match catalog v1", output)
		}
		currentHash, currentHashErr := manifest.HashContent(current, componentMode)
		if currentHashErr != nil || currentHash != recorded.BaselineSHA256 {
			return InitReport{}, failure(4, "manual.component.drift", "component output %q has local modifications", output)
		}
		report.Changes = append(report.Changes, InitChange{Path: output, Action: "unchanged", Class: componentClass, BaselineSHA256: hash, content: content})
	}
	sort.Slice(report.Changes, func(i, j int) bool { return report.Changes[i].Path < report.Changes[j].Path })
	if existingComponent >= 0 {
		report.State = "no-op"
		report.ManifestSchemaAfter = document.SchemaVersion
		return report, nil
	}
	if options.DryRun {
		report.State = "dry-run"
		return report, nil
	}
	if !options.Yes {
		return InitReport{}, failure(2, "manual.confirmation.required", "component installation requires --dry-run or --yes")
	}
	for _, change := range report.Changes {
		document.Files = append(document.Files, manifest.File{
			Path: change.Path, Owner: componentOwner, Class: componentClass,
			ContentMode: componentMode, BaselineSHA256: change.BaselineSHA256, LifecycleState: "current",
		})
	}
	document.SchemaVersion = manifest.SchemaVersionV2
	if options.SokuVersion != "" {
		document.SokuVersion = options.SokuVersion
	}
	document.Components = append(document.Components, manifest.Component{
		ID: componentID, CatalogVersion: catalogVersion, ConfigurationPath: options.ConfigPath,
	})
	sort.Slice(document.Files, func(i, j int) bool { return document.Files[i].Path < document.Files[j].Path })
	sort.Slice(document.Components, func(i, j int) bool { return document.Components[i].ID < document.Components[j].ID })
	if err := manifest.Validate(document); err != nil {
		return InitReport{}, failure(2, "manual.manifest.invalid", "construct manifest v2: %v", err)
	}
	if err := applyInitTransaction(options, report.Changes, document); err != nil {
		return InitReport{}, err
	}
	report.State = "applied"
	return report, nil
}

func installationAssets() (map[string][]byte, error) {
	result := map[string][]byte{}
	err := fs.WalkDir(componentAssets, "assets", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		content, err := componentAssets.ReadFile(name)
		if err != nil {
			return err
		}
		relative := strings.TrimPrefix(name, "assets/")
		var output string
		switch {
		case strings.HasPrefix(relative, "runner/"):
			output = "tools/manual-capture/" + strings.TrimPrefix(relative, "runner/")
		case strings.HasPrefix(relative, "schemas/"):
			output = "docs/manual/schema/" + strings.TrimPrefix(relative, "schemas/")
		case strings.HasPrefix(relative, "catalogs/"):
			output = "docs/manual/schema/" + strings.TrimPrefix(relative, "catalogs/")
		case strings.HasPrefix(relative, "examples/"):
			output = "docs/manual/" + strings.TrimPrefix(relative, "examples/")
		case strings.HasPrefix(relative, "templates/"):
			output = "docs/manual/" + strings.TrimPrefix(relative, "templates/")
		default:
			return fmt.Errorf("unexpected component asset %q", name)
		}
		if err := manifest.ValidatePath(output); err != nil {
			return err
		}
		result[output] = content
		return nil
	})
	return result, err
}

type initJournal struct {
	ID              string   `json:"id"`
	State           string   `json:"state"`
	ManifestExisted bool     `json:"manifest_existed"`
	Paths           []string `json:"paths"`
}

func applyInitTransaction(options InitOptions, changes []InitChange, document manifest.Document) error {
	id, err := initTransactionID()
	if err != nil {
		return err
	}
	directory := filepath.Join(options.Root, ".soku", "transactions", id)
	backup := filepath.Join(directory, "manifest-v1.json")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return failure(7, "manual.apply.rolled_back", "create transaction: %v", err)
	}
	manifestPath := filepath.Join(options.Root, filepath.FromSlash(manifest.ManifestPath))
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		_ = os.RemoveAll(directory)
		return failure(7, "manual.apply.rolled_back", "backup manifest: %v", err)
	}
	if err := os.WriteFile(backup, manifestData, 0o600); err != nil {
		_ = os.RemoveAll(directory)
		return failure(7, "manual.apply.rolled_back", "backup manifest: %v", err)
	}
	journal := initJournal{ID: id, State: "prepared", ManifestExisted: true, Paths: []string{}}
	for _, change := range changes {
		if change.Action != "unchanged" {
			journal.Paths = append(journal.Paths, change.Path)
		}
	}
	journalData, _ := json.MarshalIndent(journal, "", "  ")
	if err := os.WriteFile(filepath.Join(directory, "journal.json"), append(journalData, '\n'), 0o600); err != nil {
		_ = os.RemoveAll(directory)
		return failure(7, "manual.apply.rolled_back", "write transaction journal: %v", err)
	}
	rollback := func(applyErr error) error {
		for index := len(journal.Paths) - 1; index >= 0; index-- {
			target := filepath.Join(options.Root, filepath.FromSlash(journal.Paths[index]))
			_ = os.Remove(target)
			removeEmptyDirectories(filepath.Dir(target), options.Root)
		}
		_ = os.Remove(filepath.Join(options.Root, filepath.FromSlash(manifest.PendingPath)))
		restoreErr := os.WriteFile(manifestPath, manifestData, 0o600)
		cleanupErr := os.RemoveAll(directory)
		if restoreErr != nil || cleanupErr != nil {
			return failure(8, "manual.rollback.failed", "component apply failed and exact manifest restoration failed; preserve .soku/transactions/%s", id)
		}
		return failure(7, "manual.apply.rolled_back", "component apply failed and rollback restored manifest v1 and created files: %v", applyErr)
	}
	for _, change := range changes {
		if change.Action == "unchanged" {
			continue
		}
		if options.ApplyHook != nil {
			if hookErr := options.ApplyHook("before-write", change.Path); hookErr != nil {
				return rollback(hookErr)
			}
		}
		target := filepath.Join(options.Root, filepath.FromSlash(change.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return rollback(err)
		}
		temp := target + ".soku-tmp-" + id
		if err := os.WriteFile(temp, change.content, 0o644); err != nil {
			return rollback(err)
		}
		if err := os.Rename(temp, target); err != nil {
			_ = os.Remove(temp)
			return rollback(err)
		}
	}
	if options.ApplyHook != nil {
		if hookErr := options.ApplyHook("before-manifest", manifest.ManifestPath); hookErr != nil {
			return rollback(hookErr)
		}
	}
	if err := manifest.NewStore(options.Root).Write(document); err != nil {
		return rollback(err)
	}
	if err := os.RemoveAll(directory); err != nil {
		return failure(8, "manual.recovery.required", "component committed but transaction cleanup failed; preserve .soku/transactions/%s", id)
	}
	_ = os.Remove(filepath.Join(options.Root, ".soku", "transactions"))
	return nil
}

func initTransactionID() (string, error) {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return "docs-manual-" + time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(random), nil
}

func removeEmptyDirectories(directory, root string) {
	for directory != root && directory != filepath.Join(root, ".soku") {
		if err := os.Remove(directory); err != nil {
			return
		}
		directory = filepath.Dir(directory)
	}
}

// HumanInit renders the component migration and file plan.
func HumanInit(report InitReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Soku docs manual init: %s\nComponent: %s catalog v%s\nConfig: %s\nManifest: v%d -> v%d\n",
		report.State, report.ComponentID, report.CatalogVersion, report.ConfigurationPath,
		report.ManifestSchemaBefore, report.ManifestSchemaAfter)
	for _, change := range report.Changes {
		fmt.Fprintf(&builder, "  %s %s\n", change.Action, change.Path)
	}
	return builder.String()
}
