// Package lifecyclee2e verifies the complete core lifecycle with hermetic releases.
package lifecyclee2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	lifecyclestatus "github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/status"
)

const (
	testSource   = "https://github.com/example/boilerplate"
	baseCommit   = "0123456789abcdef0123456789abcdef01234567"
	targetCommit = "89abcdef0123456789abcdef0123456789abcdef"
)

type releaseFetcher map[string]initcmd.SourceSnapshot

func (fetcher releaseFetcher) Fetch(_ context.Context, source, release string) (initcmd.SourceSnapshot, error) {
	snapshot, ok := fetcher[release]
	if !ok || snapshot.Source != source {
		return initcmd.SourceSnapshot{}, &initcmd.Failure{Code: 6, Key: "source.missing", Message: "synthetic release is unavailable"}
	}
	return snapshot, nil
}

func TestCoreLifecycleFixtures(t *testing.T) {
	base := repositorySnapshot(t, "v1.0.0", baseCommit)
	target := cloneSnapshot(base)
	target.Release = "v1.1.0"
	target.ResolvedCommit = targetCommit
	target.Files["templates/go/profile.go"] = append(target.Files["templates/go/profile.go"], []byte("\n// upgraded release\n")...)
	target.Files[".gitignore"] = append(target.Files[".gitignore"], []byte("\nrelease-only/\n")...)
	fetcher := releaseFetcher{"v1.0.0": base, "v1.1.0": target}

	t.Run("empty-single-stack", func(t *testing.T) {
		root := t.TempDir()
		uninitialized, err := lifecyclestatus.Inspect(root)
		if err != nil || uninitialized.Code != 3 || uninitialized.Report.State != "uninitialized" {
			t.Fatalf("uninitialized status = %#v, %v", uninitialized.Report, err)
		}
		initialize(t, root, base, []string{"go"})
		assertClean(t, root)
		assertNoPlaceholders(t, root)
	})

	t.Run("existing-repository", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, ".gitignore", "local-cache/\n")
		writeFile(t, root, ".editorconfig", "root = true\n\n[*.local]\nindent_size = 3\n")
		initialize(t, root, base, []string{"go"})
		gitignore := readFile(t, root, ".gitignore")
		editorconfig := readFile(t, root, ".editorconfig")
		if !bytes.Contains(gitignore, []byte("local-cache/")) || !bytes.Contains(editorconfig, []byte("[*.local]")) {
			t.Fatal("initialization discarded existing mergeable content")
		}
		assertClean(t, root)
	})

	t.Run("multi-stack", func(t *testing.T) {
		root := t.TempDir()
		initialize(t, root, base, append([]string(nil), initcmd.StackIDs...))
		assertClean(t, root)
		assertNoPlaceholders(t, root)
		for _, path := range []string{"package.json", "pyproject.toml", "go.mod", "pom.xml", "db/mysql/schema.sql", "db/postgresql/schema.sql", "cloudbuild.yaml", "buildspec.yml", "azure-pipelines.yml"} {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
				t.Errorf("multi-stack output %s is missing: %v", path, err)
			}
		}
	})

	t.Run("status-edit-diff-upgrade-status", func(t *testing.T) {
		root := t.TempDir()
		initialize(t, root, base, []string{"go"})
		gitignorePath := filepath.Join(root, ".gitignore")
		gitignore := readFile(t, root, ".gitignore")
		if err := os.WriteFile(gitignorePath, append(gitignore, []byte("local-only/\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
		drifted, err := lifecyclestatus.Inspect(root)
		if err != nil || drifted.Code != 3 || drifted.Report.State != "drifted" {
			t.Fatalf("drifted status = %#v, %v", drifted.Report, err)
		}
		beforeManifest := readFile(t, root, manifest.ManifestPath)
		diff, err := initcmd.RunTransition(context.Background(), initcmd.TransitionOptions{Root: root, TargetRelease: "v1.1.0"}, fetcher, false)
		if err != nil || !diff.HasChanges || diff.State != "changes" {
			t.Fatalf("diff = %#v, %v", diff, err)
		}
		if !bytes.Equal(beforeManifest, readFile(t, root, manifest.ManifestPath)) {
			t.Fatal("diff mutated the manifest")
		}
		upgrade, err := initcmd.RunTransition(context.Background(), initcmd.TransitionOptions{Root: root, TargetRelease: "v1.1.0", Yes: true, SokuVersion: "e2e"}, fetcher, true)
		if err != nil || upgrade.State != "applied" {
			t.Fatalf("upgrade = %#v, %v", upgrade, err)
		}
		if !bytes.Contains(readFile(t, root, ".gitignore"), []byte("local-only/")) {
			t.Fatal("upgrade discarded a compatible local customization")
		}
		assertClean(t, root)
		noOp, err := initcmd.RunTransition(context.Background(), initcmd.TransitionOptions{Root: root, TargetRelease: "v1.1.0", Yes: true}, fetcher, true)
		if err != nil || noOp.State != "no-op" || noOp.HasChanges {
			t.Fatalf("rerun = %#v, %v", noOp, err)
		}
	})
}

func TestFilesystemRiskMatrix(t *testing.T) {
	base := repositorySnapshot(t, "v1.0.0", baseCommit)

	t.Run("case-collision", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "PROFILE.GO", "project-owned\n")
		_, err := initcmd.Run(context.Background(), initOptions(root), releaseFetcher{"v1.0.0": base})
		if failureCode(err) != 4 {
			t.Fatalf("case collision error = %v", err)
		}
	})

	t.Run("canonical-line-endings", func(t *testing.T) {
		root := t.TempDir()
		initialize(t, root, base, []string{"go"})
		path := filepath.Join(root, "profile.go")
		content := readFile(t, root, "profile.go")
		if err := os.WriteFile(path, bytes.ReplaceAll(content, []byte("\n"), []byte("\r\n")), 0o644); err != nil {
			t.Fatal(err)
		}
		assertClean(t, root)
	})

	t.Run("symlink-boundary", func(t *testing.T) {
		root := t.TempDir()
		initialize(t, root, base, []string{"go"})
		outside := filepath.Join(t.TempDir(), "outside.go")
		if err := os.WriteFile(outside, []byte("outside\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		managed := filepath.Join(root, "profile.go")
		if err := os.Remove(managed); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, managed); err != nil {
			if runtime.GOOS == "windows" {
				t.Skipf("symlink creation is unavailable: %v", err)
			}
			t.Fatal(err)
		}
		result, err := lifecyclestatus.Inspect(root)
		if err != nil || result.Report.State != "drifted" || fileState(result, "profile.go") != "symlink-escape" {
			t.Fatalf("symlink status = %#v, %v", result.Report, err)
		}
	})

	t.Run("deletion-rollback-and-manifest-restore", func(t *testing.T) {
		root := t.TempDir()
		initialize(t, root, base, []string{"go"})
		before := projectTree(t, root)
		target := cloneSnapshot(base)
		target.Release = "v1.1.0"
		target.ResolvedCommit = targetCommit
		catalog := decodeCatalog(t, target)
		stack := catalogStack(t, &catalog, "go")
		stack.Files = stack.Files[:len(stack.Files)-1]
		target.Files[initcmd.CatalogPath] = marshalCatalog(t, catalog)
		_, err := initcmd.RunTransition(context.Background(), initcmd.TransitionOptions{
			Root: root, TargetRelease: "v1.1.0", Yes: true,
			ApplyHook: func(stage, _ string) error {
				if stage == "before-manifest" {
					return errors.New("injected e2e failure")
				}
				return nil
			},
		}, releaseFetcher{"v1.0.0": base, "v1.1.0": target}, true)
		if failureCode(err) != 7 {
			t.Fatalf("rollback error = %v", err)
		}
		if !equalTree(before, projectTree(t, root)) {
			t.Fatal("rollback did not restore files and manifest atomically")
		}
		assertClean(t, root)
	})
}

func initialize(t *testing.T, root string, snapshot initcmd.SourceSnapshot, stacks []string) {
	t.Helper()
	options := initOptions(root)
	options.Explicit.Stacks = append([]string(nil), stacks...)
	report, err := initcmd.Run(context.Background(), options, releaseFetcher{snapshot.Release: snapshot})
	if err != nil || report.State != "applied" {
		t.Fatalf("init = %#v, %v", report, err)
	}
}

func initOptions(root string) initcmd.Options {
	return initcmd.Options{
		Root: root,
		Explicit: initcmd.Explicit{
			Source: testSource, Release: "v1.0.0", Stacks: []string{"go"}, Profile: "standard",
			ProjectName: "lifecycle-project", ModulePath: "github.com/example/lifecycle-project",
			JavaGroup: "io.example.lifecycle", ServiceName: "lifecycle-service",
			SourceSet: true, ReleaseSet: true, StacksSet: true, ProfileSet: true,
			ProjectNameSet: true, ModulePathSet: true, JavaGroupSet: true, ServiceNameSet: true,
		},
		Yes: true, SokuVersion: "e2e",
	}
}

func repositorySnapshot(t *testing.T, release, commit string) initcmd.SourceSnapshot {
	t.Helper()
	root := repositoryRoot(t)
	catalogData := readAbsolute(t, filepath.Join(root, filepath.FromSlash(initcmd.CatalogPath)))
	catalog, err := initcmd.DecodeCatalog(catalogData)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{initcmd.CatalogPath: catalogData}
	files[initcmd.ProfileIndexPath] = readAbsolute(t, filepath.Join(root, filepath.FromSlash(initcmd.ProfileIndexPath)))
	files["AGENTS.md"] = readAbsolute(t, filepath.Join(root, "AGENTS.md"))
	files[".github/CODEOWNERS"] = readAbsolute(t, filepath.Join(root, ".github", "CODEOWNERS"))
	for _, declaration := range catalog.Files {
		files[declaration.Source] = readAbsolute(t, filepath.Join(root, filepath.FromSlash(declaration.Source)))
	}
	for _, stack := range catalog.Stacks {
		for _, declaration := range stack.Files {
			files[declaration.Source] = readAbsolute(t, filepath.Join(root, filepath.FromSlash(declaration.Source)))
		}
	}
	return initcmd.SourceSnapshot{Source: testSource, Release: release, ResolvedCommit: commit, Files: files}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve lifecycle test source")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func cloneSnapshot(snapshot initcmd.SourceSnapshot) initcmd.SourceSnapshot {
	clone := snapshot
	clone.Files = make(map[string][]byte, len(snapshot.Files))
	for path, content := range snapshot.Files {
		clone.Files[path] = append([]byte(nil), content...)
	}
	return clone
}

func decodeCatalog(t *testing.T, snapshot initcmd.SourceSnapshot) initcmd.Catalog {
	t.Helper()
	catalog, err := initcmd.DecodeCatalog(snapshot.Files[initcmd.CatalogPath])
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func catalogStack(t *testing.T, catalog *initcmd.Catalog, id string) *initcmd.Stack {
	t.Helper()
	for index := range catalog.Stacks {
		if catalog.Stacks[index].ID == id {
			return &catalog.Stacks[index]
		}
	}
	t.Fatalf("catalog stack %s is missing", id)
	return nil
}

func marshalCatalog(t *testing.T, catalog initcmd.Catalog) []byte {
	t.Helper()
	data, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertClean(t *testing.T, root string) {
	t.Helper()
	result, err := lifecyclestatus.Inspect(root)
	if err != nil || result.Code != 0 || result.Report.State != "clean" {
		t.Fatalf("status = %#v, %v", result.Report, err)
	}
}

func assertNoPlaceholders(t *testing.T, root string) {
	t.Helper()
	for path, content := range projectTree(t, root) {
		if strings.HasPrefix(path, ".soku/") {
			continue
		}
		for _, placeholder := range []string{"your-project-name", "github.com/your-org/your-repo", "com.example", "your-service"} {
			if strings.Contains(content, placeholder) {
				t.Errorf("generated path %s retains placeholder %q", path, placeholder)
			}
		}
	}
}

func projectTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[filepath.ToSlash(relative)] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func equalTree(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for path, content := range left {
		if right[path] != content {
			return false
		}
	}
	return true
}

func fileState(result lifecyclestatus.Result, path string) string {
	for _, file := range result.Report.Files {
		if file.Path == path {
			return file.State
		}
	}
	return ""
}

func failureCode(err error) int {
	var failure *initcmd.Failure
	if errors.As(err, &failure) {
		return failure.Code
	}
	return -1
}

func writeFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, root, relative string) []byte {
	t.Helper()
	return readAbsolute(t, filepath.Join(root, filepath.FromSlash(relative)))
}

func readAbsolute(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
