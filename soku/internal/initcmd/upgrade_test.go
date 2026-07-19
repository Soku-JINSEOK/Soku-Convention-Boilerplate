package initcmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const targetCommit = "89abcdef0123456789abcdef0123456789abcdef"

type releaseFetcher map[string]SourceSnapshot

func (fetcher releaseFetcher) Fetch(_ context.Context, source, release string) (SourceSnapshot, error) {
	snapshot, ok := fetcher[release]
	if !ok {
		return SourceSnapshot{}, fail(6, "source.missing", "missing synthetic release %s", release)
	}
	if snapshot.Source != source {
		return SourceSnapshot{}, fail(6, "source.invalid", "unexpected source")
	}
	return snapshot, nil
}

func TestTransitionForwardMergeDeletionDryRunApplyAndNoOp(t *testing.T) {
	base := repositorySnapshot(t)
	target := cloneSnapshot(base)
	target.Release = "v1.1.0"
	target.ResolvedCommit = targetCommit
	catalog := mustCatalog(t)
	goStack := stackByID(t, &catalog, "go")
	updated := goStack.Files[0]
	removed := goStack.Files[len(goStack.Files)-1]
	target.Files[updated.Source] = append(append([]byte(nil), target.Files[updated.Source]...), []byte("\n// target release\n")...)
	goStack.Files = goStack.Files[:len(goStack.Files)-1]
	target.Files[CatalogPath] = marshalCatalog(t, catalog)

	root := initializeRelease(t, base)
	gitignorePath := filepath.Join(root, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gitignorePath, append(data, []byte("local-only/\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	before := readTree(t, root)
	fetcher := releaseFetcher{"v1.0.0": base, "v1.1.0": target}

	diff, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0"}, fetcher, false)
	if err != nil {
		t.Fatal(err)
	}
	assertAction(t, diff.Changes, updated.Output, "updated")
	assertAction(t, diff.Changes, removed.Output, "removed")
	assertAction(t, diff.Changes, ".gitignore", "merged")
	if !reflect.DeepEqual(before, readTree(t, root)) {
		t.Fatal("diff changed the repository")
	}

	dryRun, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0", DryRun: true}, fetcher, true)
	if err != nil || dryRun.State != "dry-run" {
		t.Fatalf("report=%#v err=%v", dryRun, err)
	}
	if !reflect.DeepEqual(before, readTree(t, root)) {
		t.Fatal("dry-run changed the repository")
	}

	applied, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0", Yes: true, SokuVersion: "test"}, fetcher, true)
	if err != nil || applied.State != "applied" {
		t.Fatalf("report=%#v err=%v", applied, err)
	}
	merged, _ := os.ReadFile(gitignorePath)
	if !strings.Contains(string(merged), "local-only/") {
		t.Fatal("shared-file customization was lost")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(removed.Output))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("removed output still exists: %v", err)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil || document.Boilerplate.Release != "v1.1.0" || document.Boilerplate.ResolvedCommit != targetCommit {
		t.Fatalf("manifest=%#v err=%v", document.Boilerplate, err)
	}

	noOp, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0", Yes: true}, fetcher, true)
	if err != nil || noOp.State != "no-op" || noOp.HasChanges {
		t.Fatalf("report=%#v err=%v", noOp, err)
	}
	assertAction(t, noOp.Changes, updated.Output, "unchanged")
}

func TestTransitionAddsNewlyDeclaredOutput(t *testing.T) {
	target := repositorySnapshot(t)
	target.Release = "v1.1.0"
	target.ResolvedCommit = targetCommit
	base := cloneSnapshot(target)
	base.Release = "v1.0.0"
	base.ResolvedCommit = testCommit
	catalog := mustCatalog(t)
	goStack := stackByID(t, &catalog, "go")
	added := goStack.Files[len(goStack.Files)-1]
	goStack.Files = goStack.Files[:len(goStack.Files)-1]
	base.Files[CatalogPath] = marshalCatalog(t, catalog)
	root := initializeRelease(t, base)
	report, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0"}, releaseFetcher{"v1.0.0": base, "v1.1.0": target}, false)
	if err != nil {
		t.Fatal(err)
	}
	assertAction(t, report.Changes, added.Output, "added")
}

func TestTransitionRejectsDriftDowngradeMovedTagAndProjectCollision(t *testing.T) {
	base := repositorySnapshot(t)
	target := cloneSnapshot(base)
	target.Release = "v1.1.0"
	target.ResolvedCommit = targetCommit
	catalog := mustCatalog(t)
	declaration := stackByID(t, &catalog, "go").Files[0]
	target.Files[declaration.Source] = append(target.Files[declaration.Source], []byte("\n// changed\n")...)

	t.Run("drift", func(t *testing.T) {
		root := initializeRelease(t, base)
		writeTestFile(t, filepath.Join(root, filepath.FromSlash(declaration.Output)), "local replacement\n")
		before := readTree(t, root)
		report, diffErr := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0"}, releaseFetcher{"v1.0.0": base, "v1.1.0": target}, false)
		if diffErr != nil {
			t.Fatal(diffErr)
		}
		assertAction(t, report.Changes, declaration.Output, "locally-modified")
		_, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0", Yes: true}, releaseFetcher{"v1.0.0": base, "v1.1.0": target}, true)
		if failureCode(err) != 4 || !reflect.DeepEqual(before, readTree(t, root)) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("missing-managed-file", func(t *testing.T) {
		root := initializeRelease(t, base)
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(declaration.Output))); err != nil {
			t.Fatal(err)
		}
		report, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0"}, releaseFetcher{"v1.0.0": base, "v1.1.0": target}, false)
		if err != nil {
			t.Fatal(err)
		}
		assertAction(t, report.Changes, declaration.Output, "conflict")
	})
	t.Run("downgrade", func(t *testing.T) {
		root := initializeRelease(t, base)
		_, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v0.9.0"}, releaseFetcher{"v1.0.0": base}, false)
		if failureCode(err) != 5 {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("source-change", func(t *testing.T) {
		root := initializeRelease(t, base)
		configPath := filepath.Join(root, "upgrade.yml")
		writeTestFile(t, configPath, "schema_version: 1\nboilerplate_source: https://github.com/example/other\n")
		_, err := RunTransition(context.Background(), TransitionOptions{Root: root, ConfigPath: configPath, TargetRelease: "v1.1.0"}, releaseFetcher{"v1.0.0": base, "v1.1.0": target}, false)
		if failureCode(err) != 5 {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("moved-tag", func(t *testing.T) {
		root := initializeRelease(t, base)
		moved := cloneSnapshot(base)
		moved.ResolvedCommit = targetCommit
		_, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0"}, releaseFetcher{"v1.0.0": moved, "v1.1.0": target}, false)
		if failureCode(err) != 5 {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("project-collision", func(t *testing.T) {
		trimmed := cloneSnapshot(base)
		trimmedCatalog := mustCatalog(t)
		stack := stackByID(t, &trimmedCatalog, "go")
		added := stack.Files[len(stack.Files)-1]
		stack.Files = stack.Files[:len(stack.Files)-1]
		trimmed.Files[CatalogPath] = marshalCatalog(t, trimmedCatalog)
		root := initializeRelease(t, trimmed)
		writeTestFile(t, filepath.Join(root, filepath.FromSlash(added.Output)), "project-owned\n")
		before := readTree(t, root)
		_, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0"}, releaseFetcher{"v1.0.0": trimmed, "v1.1.0": target}, false)
		if failureCode(err) != 4 || !reflect.DeepEqual(before, readTree(t, root)) {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestTransitionCancellationAndRollbackRestoreManifestAndDeletion(t *testing.T) {
	base := repositorySnapshot(t)
	target := cloneSnapshot(base)
	target.Release = "v1.1.0"
	target.ResolvedCommit = targetCommit
	catalog := mustCatalog(t)
	stack := stackByID(t, &catalog, "go")
	updated := stack.Files[0]
	removed := stack.Files[len(stack.Files)-1]
	target.Files[updated.Source] = append(target.Files[updated.Source], []byte("\n// changed\n")...)
	stack.Files = stack.Files[:len(stack.Files)-1]
	target.Files[CatalogPath] = marshalCatalog(t, catalog)
	fetcher := releaseFetcher{"v1.0.0": base, "v1.1.0": target}

	root := initializeRelease(t, base)
	before := readTree(t, root)
	cancelled, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0", Interactive: true, Confirm: func(TransitionReport) (bool, error) { return false, nil }}, fetcher, true)
	if err != nil || cancelled.State != "cancelled" || !reflect.DeepEqual(before, readTree(t, root)) {
		t.Fatalf("report=%#v err=%v", cancelled, err)
	}

	_, err = RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.1.0", Yes: true, ApplyHook: func(stage, path string) error {
		if stage == "before-manifest" {
			return errors.New("injected failure")
		}
		return nil
	}}, fetcher, true)
	if failureCode(err) != 7 {
		t.Fatalf("err=%v", err)
	}
	if !reflect.DeepEqual(before, readTree(t, root)) {
		t.Fatal("rollback did not restore the complete previous state")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(removed.Output))); err != nil {
		t.Fatalf("deleted file was not restored: %v", err)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil || document.Boilerplate.Release != "v1.0.0" {
		t.Fatalf("manifest not restored: %#v err=%v", document.Boilerplate, err)
	}
}

func TestStructuralThreeWayMergesPreserveIndependentLocalEntries(t *testing.T) {
	gitignore, err := mergeGitignoreThreeWay(
		[]byte("base-only/\nkept/\n"),
		[]byte("base-only/\nkept/\nlocal-only/\n"),
		[]byte("kept/\ntarget-only/\n"),
	)
	if err != nil {
		t.Fatal(err)
	}
	text := string(gitignore)
	if strings.Contains(text, "base-only/") || !strings.Contains(text, "local-only/") || !strings.Contains(text, "target-only/") {
		t.Fatalf("merged .gitignore=%q", text)
	}

	base := []byte("root = true\n\n[*]\nindent_size = 2\n")
	current := []byte("root = true\n\n[*]\nindent_size = 2\n\n[*.md]\ntrim_trailing_whitespace = false\n")
	target := []byte("root = true\n\n[*]\nindent_size = 4\ninsert_final_newline = true\n")
	editorconfig, err := mergeEditorconfigThreeWay(base, current, target)
	if err != nil {
		t.Fatal(err)
	}
	text = string(editorconfig)
	for _, expected := range []string{"indent_size = 4", "insert_final_newline = true", "[*.md]", "trim_trailing_whitespace = false"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("merged .editorconfig lacks %q:\n%s", expected, text)
		}
	}
	_, err = mergeEditorconfigThreeWay(base, []byte("root = true\n\n[*]\nindent_size = 8\n"), target)
	if failureCode(err) != 4 {
		t.Fatalf("conflicting merge err=%v", err)
	}
}

func initializeRelease(t *testing.T, snapshot SourceSnapshot) string {
	t.Helper()
	root := t.TempDir()
	_, err := Run(context.Background(), Options{Root: root, Explicit: Explicit{Source: snapshot.Source, Release: snapshot.Release, Stacks: []string{"go"}, ModulePath: "github.com/example/project", SourceSet: true, ReleaseSet: true, StacksSet: true, ModulePathSet: true}, Yes: true, SokuVersion: "test"}, staticFetcher{snapshot: snapshot})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func cloneSnapshot(snapshot SourceSnapshot) SourceSnapshot {
	clone := snapshot
	clone.Files = make(map[string][]byte, len(snapshot.Files))
	for path, content := range snapshot.Files {
		clone.Files[path] = append([]byte(nil), content...)
	}
	return clone
}

func stackByID(t *testing.T, catalog *Catalog, id string) *Stack {
	t.Helper()
	for index := range catalog.Stacks {
		if catalog.Stacks[index].ID == id {
			return &catalog.Stacks[index]
		}
	}
	t.Fatalf("stack %s is missing", id)
	return nil
}

func marshalCatalog(t *testing.T, catalog Catalog) []byte {
	t.Helper()
	data, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertAction(t *testing.T, changes []Change, path, action string) {
	t.Helper()
	for _, change := range changes {
		if change.Path == path {
			if change.Action != action {
				t.Fatalf("%s action=%s want=%s", path, change.Action, action)
			}
			return
		}
	}
	t.Fatalf("change for %s is missing", path)
}
