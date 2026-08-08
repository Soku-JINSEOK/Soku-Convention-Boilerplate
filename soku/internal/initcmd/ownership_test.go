package initcmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	lifecyclestatus "github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/status"
)

const ownershipTestHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestOwnershipHandoffDryRunApplyStatusAndFutureUpgrade(t *testing.T) {
	base := repositorySnapshot(t)
	root := initializeOwnershipRelease(t, base)
	path := ".prettierignore"
	fullPath := filepath.Join(root, path)
	fixture := mustRead(t, "../../testdata/ownership-handoff/report-hub.prettierignore")
	if err := os.WriteFile(fullPath, fixture, 0o640); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(fullPath, 0o640); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := manifest.HashContent(fixture, "text")
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, filepath.FromSlash(manifest.ManifestPath))
	manifestBefore := mustRead(t, manifestPath)
	infoBefore, err := os.Stat(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	dryRun, err := HandoffOwnership(HandoffOptions{
		Root: root, Path: path, ExpectedSHA256: expected, DryRun: true, SokuVersion: "v0.3.0",
	})
	if err != nil || dryRun.State != "dry-run" || dryRun.ManifestSchemaBefore != 1 || dryRun.ManifestSchemaAfter != 3 {
		t.Fatalf("dry run = %#v, %v", dryRun, err)
	}
	if !bytes.Equal(manifestBefore, mustRead(t, manifestPath)) {
		t.Fatal("dry-run changed manifest bytes")
	}
	if _, err := os.Stat(filepath.Join(root, ".soku", "transactions")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run created transaction state: %v", err)
	}

	applied, err := HandoffOwnership(HandoffOptions{
		Root: root, Path: path, ExpectedSHA256: expected, Yes: true, SokuVersion: "v0.3.0",
	})
	if err != nil || applied.State != "applied" {
		t.Fatalf("apply = %#v, %v", applied, err)
	}
	assertFileBytesAndMode(t, fullPath, fixture, infoBefore.Mode())
	document, err := manifest.NewStore(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if document.SchemaVersion != manifest.SchemaVersionV3 ||
		len(document.Selection.ProjectOwnedOverrides) != 1 || document.Selection.ProjectOwnedOverrides[0] != path {
		t.Fatalf("manifest selection = %#v", document.Selection)
	}
	file := manifestFile(t, document, path)
	if file.Owner != "project" || file.Class != "project-owned" || file.LifecycleState != "unmanaged-expected" || file.ContentMode != "" || file.BaselineSHA256 != "" {
		t.Fatalf("handoff file = %#v", file)
	}
	status, err := lifecyclestatus.Inspect(root)
	if err != nil || status.Code != 0 || status.Report.State != "clean" || status.Report.Counts.UnmanagedExpected != 1 {
		t.Fatalf("status = %#v, %v", status, err)
	}

	sameRelease, err := RunTransition(context.Background(), TransitionOptions{
		Root: root, TargetRelease: base.Release,
	}, releaseFetcher{base.Release: base}, false)
	if err != nil || sameRelease.State != "no-op" || sameRelease.HasChanges {
		t.Fatalf("same-release diff = %#v, %v", sameRelease, err)
	}

	target := cloneSnapshot(base)
	target.Release = "v1.1.0"
	target.ResolvedCommit = targetCommit
	catalog := mustCatalog(t)
	for _, declaration := range catalog.Files {
		if declaration.Output == path {
			target.Files[declaration.Source] = []byte("future core replacement\n")
		}
	}
	future, err := RunTransition(context.Background(), TransitionOptions{
		Root: root, TargetRelease: target.Release, Yes: true, SokuVersion: "v0.3.0",
	}, releaseFetcher{base.Release: base, target.Release: target}, true)
	if err != nil || future.State != "applied" {
		t.Fatalf("future upgrade = %#v, %v", future, err)
	}
	assertFileBytesAndMode(t, fullPath, fixture, infoBefore.Mode())
	document, err = manifest.NewStore(root).Load()
	if err != nil || document.SchemaVersion != manifest.SchemaVersionV3 || document.Boilerplate.Release != target.Release {
		t.Fatalf("future manifest = %#v, %v", document, err)
	}
	if got := manifestFile(t, document, path); got.Class != "project-owned" {
		t.Fatalf("future handoff file = %#v", got)
	}
}

func TestOwnershipHandoffRejectsInvalidAndIneligibleInputs(t *testing.T) {
	snapshot := repositorySnapshot(t)
	tests := []struct {
		name   string
		path   string
		hash   string
		mutate func(*testing.T, string)
	}{
		{name: "non canonical path", path: "../escape", hash: ownershipTestHash},
		{name: "case mismatch", path: ".PRETTIERIGNORE", hash: ownershipTestHash},
		{name: "invalid hash", path: ".prettierignore", hash: "ABC"},
		{name: "clean core path", path: ".prettierignore", hash: ownershipTestHash},
		{name: "stale hash", path: ".prettierignore", hash: ownershipTestHash, mutate: func(t *testing.T, root string) {
			writeTestFile(t, filepath.Join(root, ".prettierignore"), "changed\n")
		}},
		{name: "obsolete path", path: ".prettierignore", hash: ownershipTestHash, mutate: func(t *testing.T, root string) {
			mutateOwnershipManifest(t, root, func(document *manifest.Document) {
				manifestFilePointer(t, document, ".prettierignore").LifecycleState = "obsolete"
			})
		}},
		{name: "mergeable path", path: ".gitignore", hash: ownershipTestHash},
		{name: "provider managed path", path: ".prettierignore", hash: ownershipTestHash, mutate: func(t *testing.T, root string) {
			mutateOwnershipManifest(t, root, func(document *manifest.Document) {
				file := manifestFilePointer(t, document, ".prettierignore")
				file.Owner = "test-provider"
				file.Class = "provider-managed"
				document.Integrations = []manifest.Integration{{
					ID: "test-provider", Source: "https://github.com/example/provider",
					Ref: testCommit, ProviderAPIVersion: "v1", ProviderSchemaVersion: "v1",
					ConfigurationHash: ownershipTestHash, LifecycleState: "connected",
					ManagedFiles: []string{".prettierignore"},
				}}
			})
		}},
		{name: "already project owned", path: ".prettierignore", hash: ownershipTestHash, mutate: func(t *testing.T, root string) {
			mutateOwnershipManifest(t, root, func(document *manifest.Document) {
				file := manifestFilePointer(t, document, ".prettierignore")
				*file = manifest.File{Path: ".prettierignore", Owner: "project", Class: "project-owned", LifecycleState: "unmanaged-expected"}
				document.SchemaVersion = manifest.SchemaVersionV3
				document.Selection.ProjectOwnedOverrides = []string{".prettierignore"}
				document.Selection.ConfigurationHash, _ = manifest.HashSelection(document.Selection)
			})
		}},
		{name: "missing path", path: ".prettierignore", hash: ownershipTestHash, mutate: func(t *testing.T, root string) {
			if err := os.Remove(filepath.Join(root, ".prettierignore")); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "non regular path", path: ".prettierignore", hash: ownershipTestHash, mutate: func(t *testing.T, root string) {
			if err := os.Remove(filepath.Join(root, ".prettierignore")); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(root, ".prettierignore"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "symlink path", path: ".prettierignore", hash: ownershipTestHash, mutate: func(t *testing.T, root string) {
			if runtime.GOOS == "windows" {
				t.Skip("symlink creation requires platform privileges")
			}
			target := filepath.Join(root, "target.txt")
			writeTestFile(t, target, "changed\n")
			if err := os.Remove(filepath.Join(root, ".prettierignore")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("target.txt", filepath.Join(root, ".prettierignore")); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := initializeOwnershipRelease(t, snapshot)
			if test.mutate != nil {
				test.mutate(t, root)
			}
			before := readTree(t, root)
			_, err := HandoffOwnership(HandoffOptions{
				Root: root, Path: test.path, ExpectedSHA256: test.hash, DryRun: true,
			})
			if err == nil {
				t.Fatal("ineligible handoff was accepted")
			}
			if after := readTree(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("rejected handoff changed repository")
			}
		})
	}
}

func mutateOwnershipManifest(t *testing.T, root string, mutate func(*manifest.Document)) {
	t.Helper()
	store := manifest.NewStore(root)
	document, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	mutate(&document)
	if err := store.Write(document); err != nil {
		t.Fatal(err)
	}
}

func manifestFilePointer(t *testing.T, document *manifest.Document, path string) *manifest.File {
	t.Helper()
	for index := range document.Files {
		if document.Files[index].Path == path {
			return &document.Files[index]
		}
	}
	t.Fatalf("manifest file %s is missing", path)
	return nil
}

func TestOwnershipHandoffRejectsStalePlanAndRollsBackManifest(t *testing.T) {
	snapshot := repositorySnapshot(t)
	root := initializeOwnershipRelease(t, snapshot)
	path := ".prettierignore"
	fullPath := filepath.Join(root, path)
	writeTestFile(t, fullPath, "approved change\n")
	expected, _ := manifest.HashContent([]byte("approved change\n"), "text")
	manifestPath := filepath.Join(root, filepath.FromSlash(manifest.ManifestPath))
	manifestBefore := mustRead(t, manifestPath)

	_, err := HandoffOwnership(HandoffOptions{
		Root: root, Path: path, ExpectedSHA256: expected, Yes: true,
		ApplyHook: func(stage, _ string) error {
			if stage == "before-manifest" {
				return os.WriteFile(fullPath, []byte("changed after approval\n"), 0o644)
			}
			return nil
		},
	})
	if failureCode(err) != 7 {
		t.Fatalf("stale apply error = %v", err)
	}
	if !bytes.Equal(manifestBefore, mustRead(t, manifestPath)) {
		t.Fatal("stale apply did not restore exact manifest bytes")
	}
	if _, err := os.Stat(filepath.Join(root, ".soku", "transactions")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rollback left transaction state: %v", err)
	}
}

func manifestFile(t *testing.T, document manifest.Document, path string) manifest.File {
	t.Helper()
	for _, file := range document.Files {
		if file.Path == path {
			return file
		}
	}
	t.Fatalf("manifest file %s is missing", path)
	return manifest.File{}
}

func initializeOwnershipRelease(t *testing.T, snapshot SourceSnapshot) string {
	t.Helper()
	root := t.TempDir()
	_, err := Run(context.Background(), Options{
		Root: root,
		Explicit: Explicit{
			Source: snapshot.Source, Release: snapshot.Release,
			Stacks: []string{"javascript-typescript-node"}, ProjectName: "ownership-fixture",
			SourceSet: true, ReleaseSet: true, StacksSet: true, ProjectNameSet: true,
		},
		Yes: true, SokuVersion: "test",
	}, staticFetcher{snapshot: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func assertFileBytesAndMode(t *testing.T, path string, content []byte, mode os.FileMode) {
	t.Helper()
	current, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(current, content) {
		t.Fatalf("file bytes changed: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode() != mode {
		t.Fatalf("file mode = %v, want %v", info.Mode(), mode)
	}
}
