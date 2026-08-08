package initcmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	lifecyclestatus "github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/status"
)

func TestProjectSyncFreshInstallIsOptInAndTransactional(t *testing.T) {
	snapshot := repositorySnapshot(t)
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	root := t.TempDir()

	dryRun, err := Run(context.Background(), Options{
		Root: root, Explicit: explicit, ProjectSync: true, ProjectSyncProjectNumber: 17, DryRun: true,
	}, staticFetcher{snapshot: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if dryRun.State != "dry-run" || len(dryRun.Components) != 1 {
		t.Fatalf("dry-run = %#v", dryRun)
	}
	expectedPaths := map[string]bool{
		".github/project-sync.yml":             false,
		".github/workflows/project-sync.yml":   false,
		"scripts/github-project-sync.mjs":      false,
		"scripts/github-project-sync.test.mjs": false,
	}
	for _, change := range dryRun.Changes {
		if _, ok := expectedPaths[change.Path]; ok {
			expectedPaths[change.Path] = true
		}
	}
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("dry-run did not list Project Sync asset %q", path)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".soku")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run changed state: %v", err)
	}

	applied, err := Run(context.Background(), Options{
		Root: root, Explicit: explicit, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true, SokuVersion: "test",
	}, staticFetcher{snapshot: snapshot})
	if err != nil || applied.State != "applied" {
		t.Fatalf("apply = %#v, %v", applied, err)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if document.SchemaVersion != manifest.SchemaVersionV2 || !hasProjectSyncComponent(document) {
		t.Fatalf("manifest = %#v", document)
	}
	configData, err := os.ReadFile(filepath.Join(root, ".github/project-sync.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatal(err)
	}
	if config["repository"] != nil || !strings.Contains(string(configData), `"number": 17`) {
		t.Fatalf("generated configuration = %s", configData)
	}
	if strings.Contains(string(configData), "August 2026") || strings.Contains(string(configData), "Soku-JINSEOK") {
		t.Fatalf("generated configuration contains repository metadata: %s", configData)
	}
	workflow, err := os.ReadFile(filepath.Join(root, ".github/workflows/project-sync.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(workflow), "vars.PROJECT_SYNC_ENABLED == 'true'") || !strings.Contains(string(workflow), `mode="${DISPATCH_MODE:-${PROJECT_SYNC_MODE:-audit}}"`) {
		t.Fatalf("workflow is not guarded: %s", workflow)
	}

	customConfig := []byte(`{"schemaVersion":1,"project":{"owner":"@me","number":99}}
`)
	if err := os.WriteFile(filepath.Join(root, ".github/project-sync.yml"), customConfig, 0o644); err != nil {
		t.Fatal(err)
	}
	noOp, err := Run(context.Background(), Options{
		Root: root, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true,
	}, staticFetcher{err: errors.New("boilerplate fetch must not run for component-only install")})
	if err != nil || noOp.State != "no-op" {
		t.Fatalf("repeat install = %#v, %v", noOp, err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, ".github/project-sync.yml")); string(got) != string(customConfig) {
		t.Fatal("project-owned configuration was overwritten")
	}
}

func TestProjectSyncRejectsExistingFilesAndMissingNumber(t *testing.T) {
	snapshot := repositorySnapshot(t)
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "scripts/github-project-sync.mjs"), []byte("local\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, collisionErr := Run(context.Background(), Options{Root: root, Explicit: explicit, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true}, staticFetcher{snapshot: snapshot})
	if failureCode(collisionErr) != 4 {
		t.Fatalf("collision error = %v", collisionErr)
	}
	if _, err := os.Stat(filepath.Join(root, ".soku")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("collision changed state: %v", err)
	}
	if _, err := Run(context.Background(), Options{Root: t.TempDir(), ProjectSync: true, Yes: true}, staticFetcher{snapshot: snapshot}); failureCode(err) != 2 {
		t.Fatalf("missing number error = %v", err)
	}
}

func TestProjectSyncInstallsIntoInitializedRepositoryWithoutFetching(t *testing.T) {
	snapshot := repositorySnapshot(t)
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	root := t.TempDir()
	if _, err := Run(context.Background(), Options{Root: root, Explicit: explicit, Yes: true}, staticFetcher{snapshot: snapshot}); err != nil {
		t.Fatal(err)
	}

	dryRun, err := Run(context.Background(), Options{
		Root: root, ProjectSync: true, ProjectSyncProjectNumber: 17, DryRun: true,
	}, staticFetcher{err: errors.New("component-only dry-run must not fetch the boilerplate")})
	if err != nil || dryRun.State != "dry-run" || len(dryRun.Components) != 1 {
		t.Fatalf("component-only dry-run = %#v, %v", dryRun, err)
	}

	report, err := Run(context.Background(), Options{
		Root: root, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true,
	}, staticFetcher{err: errors.New("component-only installation must not fetch the boilerplate")})
	if err != nil || report.State != "applied" {
		t.Fatalf("component-only install = %#v, %v", report, err)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if document.SchemaVersion != manifest.SchemaVersionV2 || !hasProjectSyncComponent(document) {
		t.Fatalf("migrated manifest = %#v", document)
	}
}

func TestProjectSyncInstallationPreservesManifestV3(t *testing.T) {
	snapshot := repositorySnapshot(t)
	root := initializeOwnershipRelease(t, snapshot)
	path := ".prettierignore"
	content := []byte("project-owned formatting boundary\n")
	if err := os.WriteFile(filepath.Join(root, path), content, 0o644); err != nil {
		t.Fatal(err)
	}
	hash, _ := manifest.HashContent(content, "text")
	if _, err := HandoffOwnership(HandoffOptions{
		Root: root, Path: path, ExpectedSHA256: hash, Yes: true, SokuVersion: "v0.3.0",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), Options{
		Root: root, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true, SokuVersion: "v0.3.0",
	}, staticFetcher{err: errors.New("component-only installation must not fetch")}); err != nil {
		t.Fatal(err)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil || document.SchemaVersion != manifest.SchemaVersionV3 || !hasProjectSyncComponent(document) ||
		len(document.Selection.ProjectOwnedOverrides) != 1 || document.Selection.ProjectOwnedOverrides[0] != path {
		t.Fatalf("manifest = %#v, %v", document, err)
	}
	if got, err := os.ReadFile(filepath.Join(root, path)); err != nil || string(got) != string(content) {
		t.Fatalf("project-owned path changed: %v", err)
	}
}

func TestProjectSyncV1MigrationRollbackRestoresExactManifest(t *testing.T) {
	snapshot := repositorySnapshot(t)
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	root := t.TempDir()
	if _, err := Run(context.Background(), Options{Root: root, Explicit: explicit, Yes: true}, staticFetcher{snapshot: snapshot}); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, ".soku", "manifest.json")
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Run(context.Background(), Options{
		Root: root, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true,
		ApplyHook: func(stage, _ string) error {
			if stage == "before-manifest" {
				return errors.New("inject v1 migration failure")
			}
			return nil
		},
	}, staticFetcher{err: errors.New("component-only migration must not fetch the boilerplate")})
	if failureCode(err) != 7 {
		t.Fatalf("migration rollback error = %v", err)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil || string(after) != string(before) {
		t.Fatalf("v1 manifest was not restored exactly: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "project-sync.yml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("component output remained after migration rollback: %v", err)
	}
}

func TestProjectSyncRollbackLeavesNoPartialComponentState(t *testing.T) {
	snapshot := repositorySnapshot(t)
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	root := t.TempDir()
	_, err := Run(context.Background(), Options{
		Root: root, Explicit: explicit, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true,
		ApplyHook: func(stage, _ string) error {
			if stage == "before-manifest" {
				return errors.New("inject component manifest failure")
			}
			return nil
		},
	}, staticFetcher{snapshot: snapshot})
	if failureCode(err) != 7 {
		t.Fatalf("rollback error = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".github", "project-sync.yml")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("project-owned config remained after rollback: %v", statErr)
	}
	if _, loadErr := manifest.NewStore(root).Load(); !errors.Is(loadErr, manifest.ErrNotInitialized) {
		t.Fatalf("manifest after rollback = %v", loadErr)
	}
}

func TestProjectSyncIsPreservedByDiffAndStatus(t *testing.T) {
	snapshot := repositorySnapshot(t)
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	root := t.TempDir()
	if _, err := Run(context.Background(), Options{
		Root: root, Explicit: explicit, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true,
	}, staticFetcher{snapshot: snapshot}); err != nil {
		t.Fatal(err)
	}
	customConfig := []byte(`{"schemaVersion":1,"project":{"owner":"@me","number":99}}
`)
	configPath := filepath.Join(root, ".github", "project-sync.yml")
	if err := os.WriteFile(configPath, customConfig, 0o644); err != nil {
		t.Fatal(err)
	}

	diff, err := RunTransition(context.Background(), TransitionOptions{
		Root: root, TargetRelease: "v1.0.0",
	}, staticFetcher{snapshot: snapshot}, false)
	if err != nil || diff.State != "no-op" || len(diff.Components) != 1 {
		t.Fatalf("clean diff = %#v, %v", diff, err)
	}
	workflowPath := filepath.Join(root, ".github", "workflows", "project-sync.yml")
	workflow, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workflowPath, append(workflow, []byte("# local drift\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	drifted, err := RunTransition(context.Background(), TransitionOptions{
		Root: root, TargetRelease: "v1.0.0",
	}, staticFetcher{snapshot: snapshot}, false)
	if err != nil || !drifted.HasChanges || len(drifted.Components) != 1 {
		t.Fatalf("drifted diff = %#v, %v", drifted, err)
	}
	statusResult, err := lifecyclestatus.Inspect(root)
	if err != nil || statusResult.Report.Components[0].ID != "github-project-sync" {
		t.Fatalf("status = %#v, %v", statusResult, err)
	}
	if got, readErr := os.ReadFile(configPath); readErr != nil || string(got) != string(customConfig) {
		t.Fatalf("project-owned config changed: %v", readErr)
	}
}

func TestProjectSyncUpgradeRejectsUnsupportedCatalogMetadata(t *testing.T) {
	snapshot := repositorySnapshot(t)
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	root := t.TempDir()
	if _, err := Run(context.Background(), Options{
		Root: root, Explicit: explicit, ProjectSync: true, ProjectSyncProjectNumber: 17, Yes: true,
	}, staticFetcher{snapshot: snapshot}); err != nil {
		t.Fatal(err)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	document.Components[0].CatalogVersion = "99"
	if err := manifest.NewStore(root).Write(document); err != nil {
		t.Fatal(err)
	}
	_, err = RunTransition(context.Background(), TransitionOptions{
		Root: root, TargetRelease: "v1.0.0",
	}, staticFetcher{snapshot: snapshot}, false)
	if failureCode(err) != 5 {
		t.Fatalf("unsupported catalog error = %v", err)
	}
}
