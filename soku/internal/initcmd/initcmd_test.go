package initcmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const testSource = "https://github.com/example/boilerplate"
const testCommit = "0123456789abcdef0123456789abcdef01234567"

type staticFetcher struct {
	snapshot SourceSnapshot
	err      error
}

func (f staticFetcher) Fetch(context.Context, string, string) (SourceSnapshot, error) {
	return f.snapshot, f.err
}

func TestLoadConfigSupportsBlockSequenceAndRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "soku.yml")
	writeTestFile(t, path, "schema_version: 1\nboilerplate_source: https://github.com/example/boilerplate\nboilerplate_release: v1.0.0\nstacks:\n  - go\n  - mysql\nverify: true\n")
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config.Stacks, []string{"go", "mysql"}) || !config.Verify {
		t.Fatalf("config=%#v", config)
	}
	writeTestFile(t, path, "schema_version: 1\nunknown: value\n")
	if _, err := LoadConfig(path); failureCode(err) != 2 {
		t.Fatalf("err=%v", err)
	}
}

func TestPublishedCatalogMatchesSchema(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiled, err := compiler.Compile("../../schema/catalog-core-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(mustRead(t, "../../catalog/core-v1.json")))
	if err != nil {
		t.Fatal(err)
	}
	if err := compiled.Validate(instance); err != nil {
		t.Fatal(err)
	}
	catalog := mustCatalog(t)
	catalog.Stacks[0].Files[0].Output = "README.md"
	unsafe, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCatalog(unsafe); failureCode(err) != 5 {
		t.Fatalf("unsafe output err=%v", err)
	}
}

func TestResolveConfigPrecedenceDetectionAndHash(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "package.json"), "{}\n")
	catalog := mustCatalog(t)
	file := Config{SchemaVersion: 1, BoilerplateSource: testSource, BoilerplateRelease: "v1.0.0", Stacks: []string{"python"}, ProjectName: "from-file"}
	resolved, err := ResolveConfig(root, file, Explicit{Stacks: []string{"go"}, StacksSet: true, ModulePath: "github.com/example/demo", ModulePathSet: true}, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resolved.Stacks, []string{"go"}) {
		t.Fatalf("stacks=%v", resolved.Stacks)
	}
	first, _ := configHash(resolved)
	resolved.Verify = !resolved.Verify
	second, _ := configHash(resolved)
	if first != second {
		t.Fatal("verify changed configuration hash")
	}
	detected, err := ResolveConfig(root, Config{SchemaVersion: 1, BoilerplateSource: testSource, BoilerplateRelease: "v1.0.0"}, Explicit{}, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(detected.Stacks, []string{"javascript-typescript-node"}) {
		t.Fatalf("detected=%v", detected.Stacks)
	}
}

func TestCatalogRenderingUsesExactTokensAndJavaPaths(t *testing.T) {
	snapshot := repositorySnapshot(t)
	catalog := mustCatalog(t)
	config := Config{SchemaVersion: 1, Profile: "standard", Stacks: []string{"gcp", "go", "java-spring", "javascript-typescript-node", "mysql", "postgresql", "python"}, ProjectName: "demo-project", ModulePath: "github.com/example/demo", JavaGroup: "io.example.demo", ServiceName: "demo-api"}
	changes, err := renderCatalog(snapshot, catalog, config)
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]Change{}
	for _, change := range changes {
		paths[change.Path] = change
		if bytes.Contains(change.Content, []byte("your-project-name")) || bytes.Contains(change.Content, []byte("your-service")) {
			t.Fatalf("unresolved token in %s", change.Path)
		}
	}
	for _, path := range []string{"db/mysql/schema.sql", "db/postgresql/schema.sql", "src/main/java/io/example/demo/profile/Application.java", ".github/workflows/ci.yml"} {
		if _, ok := paths[path]; !ok {
			t.Errorf("missing %s", path)
		}
	}
	if strings.Contains(string(paths[".github/workflows/ci.yml"].Content), "# javascript") || !strings.Contains(string(paths[".github/workflows/ci.yml"].Content), "java-spring:") {
		t.Fatal("CI was not selected deterministically")
	}
	workflow := string(paths[".github/workflows/ci.yml"].Content)
	for _, action := range []string{
		"actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7",
		"actions/setup-node@820762786026740c76f36085b0efc47a31fe5020 # v7",
		"actions/setup-python@ece7cb06caefa5fff74198d8649806c4678c61a1 # v6",
		"actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7",
		"actions/setup-java@03ad4de0992f5dab5e18fcb136590ce7c4a0ac95 # v5",
	} {
		if !strings.Contains(workflow, action) {
			t.Errorf("generated CI does not pin %s", action)
		}
	}
	if regexp.MustCompile(`uses:\s+[^\s]+@v\d+`).MatchString(workflow) {
		t.Fatal("generated CI contains a mutable major-version action reference")
	}
}

func TestManifestSelectionAndPinnedSnapshotReproduceDesiredTree(t *testing.T) {
	snapshot := repositorySnapshot(t)
	catalog := mustCatalog(t)
	config := Config{SchemaVersion: 1, Profile: "standard", Stacks: []string{"gcp", "go", "java-spring", "javascript-typescript-node", "python"}, ProjectName: "demo-project", ModulePath: "github.com/example/demo", JavaGroup: "io.example.demo", ServiceName: "demo-api"}
	desired, err := renderCatalog(snapshot, catalog, config)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := configHash(config)
	if err != nil {
		t.Fatal(err)
	}
	document, err := buildManifest("test", snapshot, config, hash, desired)
	if err != nil {
		t.Fatal(err)
	}
	reproduced, err := renderCatalog(snapshot, catalog, configFromSelection(document.Selection))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(desired, reproduced) {
		t.Fatal("manifest selection and pinned source did not reproduce the desired tree")
	}
}

func TestMergePreservesExistingValuesAndOrder(t *testing.T) {
	editor := []byte("root = true\n\n[*]\nindent_size = 8\n")
	desired := []byte("root = true\n\n[*]\ncharset = utf-8\nindent_size = 2\n\n[*.go]\nindent_style = tab\n")
	merged, err := mergeEditorconfig(editor, desired)
	if err != nil {
		t.Fatal(err)
	}
	text := string(merged)
	if !strings.Contains(text, "indent_size = 8") || strings.Contains(text, "indent_size = 2") || strings.Index(text, "charset = utf-8") > strings.Index(text, "[*.go]") {
		t.Fatalf("merged:\n%s", text)
	}
	ignored, err := mergeGitignore([]byte("custom/\nnode_modules/\n"), []byte("node_modules/\ndist/\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(ignored), "custom/\nnode_modules/\n") || strings.Count(string(ignored), "dist/") != 1 || !strings.Contains(string(ignored), gitignoreBegin) {
		t.Fatalf("merged:\n%s", ignored)
	}
	if _, err := mergeEditorconfig([]byte("[*]\na=1\n[*]\nb=2\n"), desired); failureCode(err) != 4 {
		t.Fatalf("err=%v", err)
	}
}

func TestRunDryRunApplyRerunAndConflict(t *testing.T) {
	snapshot := repositorySnapshot(t)
	fetcher := staticFetcher{snapshot: snapshot}
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"javascript-typescript-node"}, ProjectName: "demo", SourceSet: true, ReleaseSet: true, StacksSet: true, ProjectNameSet: true}
	root := t.TempDir()
	report, err := Run(context.Background(), Options{Root: root, Explicit: explicit, DryRun: true, SokuVersion: "v1.0.0"}, fetcher)
	if err != nil {
		t.Fatal(err)
	}
	if report.State != "dry-run" || len(report.Changes) == 0 {
		t.Fatalf("report=%#v", report)
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("dry-run wrote %v", entries)
	}
	report, err = Run(context.Background(), Options{Root: root, Explicit: explicit, Yes: true, SokuVersion: "v1.0.0"}, fetcher)
	if err != nil {
		t.Fatal(err)
	}
	if report.State != "applied" {
		t.Fatalf("state=%s", report.State)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if document.Selection.ProjectName != "demo" || document.Selection.ModulePath != "" || document.Selection.JavaGroup != "" || document.Selection.ServiceName != "" {
		t.Fatalf("stored selection=%#v", document.Selection)
	}
	recomputed, _ := manifest.HashSelection(document.Selection)
	if recomputed != document.Selection.ConfigurationHash {
		t.Fatalf("configuration hash=%s want %s", document.Selection.ConfigurationHash, recomputed)
	}
	report, err = Run(context.Background(), Options{Root: root, Explicit: explicit, Yes: true, SokuVersion: "v1.0.0"}, fetcher)
	if err != nil || report.State != "no-op" {
		t.Fatalf("report=%#v err=%v", report, err)
	}
	writeTestFile(t, filepath.Join(root, "package.json"), "drift\n")
	if _, err := Run(context.Background(), Options{Root: root, Explicit: explicit, Yes: true}, fetcher); failureCode(err) != 4 {
		t.Fatalf("err=%v", err)
	}
	conflictRoot := t.TempDir()
	writeTestFile(t, filepath.Join(conflictRoot, "package.json"), string(snapshot.Files["templates/javascript-typescript-node/package.json"]))
	before := readTree(t, conflictRoot)
	if _, err := Run(context.Background(), Options{Root: conflictRoot, Explicit: explicit, Yes: true}, fetcher); failureCode(err) != 4 {
		t.Fatalf("err=%v", err)
	}
	if !reflect.DeepEqual(before, readTree(t, conflictRoot)) {
		t.Fatal("conflict changed target")
	}
}

func TestRunCancelIsZeroWrite(t *testing.T) {
	root := t.TempDir()
	explicit := Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"mysql"}, SourceSet: true, ReleaseSet: true, StacksSet: true}
	report, err := Run(context.Background(), Options{Root: root, Explicit: explicit, Interactive: true, Confirm: func(Report) (bool, error) { return false, nil }}, staticFetcher{snapshot: repositorySnapshot(t)})
	if err != nil || report.State != "cancelled" {
		t.Fatalf("report=%#v err=%v", report, err)
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("cancel wrote %v", entries)
	}
}

func TestTransactionRollbackAndRecoveryEvidence(t *testing.T) {
	root := t.TempDir()
	change := Change{Path: "nested/file.txt", Action: "create", Owner: "core", Class: "core-managed", ContentMode: "text", Content: []byte("new\n")}
	change.BaselineSHA256, _ = manifest.HashContent(change.Content, "text")
	document := testManifest(t, change)
	_, err := applyTransaction(root, []Change{change}, document, func(stage, path string) error {
		if stage == "before-manifest" {
			return errors.New("injected")
		}
		return nil
	})
	if failureCode(err) != 7 {
		t.Fatalf("err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "nested")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("rollback left directory: %v", err)
	}
	root = t.TempDir()
	id, err := applyTransaction(root, []Change{change}, document, func(stage, path string) error {
		if stage == "before-write" || stage == "before-rollback" {
			return errors.New("injected")
		}
		return nil
	})
	if failureCode(err) != 8 {
		t.Fatalf("err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".soku", "transactions", id, "journal.json")); err != nil {
		t.Fatalf("missing recovery journal: %v", err)
	}
}

func TestArchiveSecurityValidation(t *testing.T) {
	valid := makeArchive(t, []tarItem{{name: "root/soku/catalog/core-v1.json", body: "{}"}})
	if _, err := extractArchive(valid); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		items []tarItem
	}{{"traversal", []tarItem{{name: "root/../evil", body: "x"}}}, {"symlink", []tarItem{{name: "root/link", kind: tar.TypeSymlink, link: "../../evil"}, {name: "root/soku/catalog/core-v1.json", body: "{}"}}}, {"case collision", []tarItem{{name: "root/A", body: "1"}, {name: "root/a", body: "2"}, {name: "root/soku/catalog/core-v1.json", body: "{}"}}}, {"windows reserved", []tarItem{{name: "root/CON.txt", body: "x"}, {name: "root/soku/catalog/core-v1.json", body: "{}"}}}, {"secret", []tarItem{{name: "root/token.txt", body: "api_key=abcdefghijklmnop123456"}, {name: "root/soku/catalog/core-v1.json", body: "{}"}}}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := extractArchive(makeArchive(t, test.items)); failureCode(err) != 6 {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestArchiveAcceptsPAXGlobalHeader(t *testing.T) {
	archive := makeArchive(t, []tarItem{
		{kind: tar.TypeXGlobalHeader, pax: map[string]string{"comment": "github"}},
		{name: "root/soku/catalog/core-v1.json", body: "{}"},
	})
	if _, err := extractArchive(archive); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveAllowsManifestSecretFixtures(t *testing.T) {
	archive := makeArchive(t, []tarItem{
		{name: "root/soku/testdata/manifest-v1/invalid/raw-configuration.json", body: `{"password":"must-not-be-stored"}`},
		{name: "root/soku/catalog/core-v1.json", body: "{}"},
	})
	if _, err := extractArchive(archive); err != nil {
		t.Fatal(err)
	}
}

func TestSourceAndPathValidation(t *testing.T) {
	for _, value := range []string{"http://github.com/o/r", "https://user@github.com/o/r", "https://gitlab.com/o/r", "https://github.com/o/r?token=x", "https://github.com/o/r#x"} {
		if _, _, err := parseGitHubSource(value); failureCode(err) != 2 {
			t.Errorf("%s err=%v", value, err)
		}
	}
	if _, _, err := parseGitHubSource(testSource); err != nil {
		t.Fatal(err)
	}
	if err := validateArchivePath("dir/aux.txt"); failureCode(err) != 6 {
		t.Fatalf("err=%v", err)
	}
}

func TestSourceClientResolvesAnnotatedTagToCommit(t *testing.T) {
	archive := makeArchive(t, []tarItem{{name: "root/soku/catalog/core-v1.json", body: "{}"}})
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body []byte
		switch request.URL.Path {
		case "/repos/example/boilerplate/git/ref/tags/v1.0.0":
			body = []byte(`{"object":{"type":"tag","sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","url":"https://api.test/tag-object"}}`)
		case "/tag-object":
			body = []byte(`{"object":{"type":"commit","sha":"` + testCommit + `","url":""}}`)
		case "/repos/example/boilerplate/tarball/" + testCommit:
			body = archive
		default:
			return &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not Found", Body: io.NopCloser(strings.NewReader("")), Request: request}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(body)), Request: request, Header: make(http.Header)}, nil
	})
	client := &SourceClient{HTTP: &http.Client{Transport: transport}, APIBase: "https://api.test"}
	snapshot, err := client.Fetch(context.Background(), testSource, "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.ResolvedCommit != testCommit {
		t.Fatalf("snapshot=%#v", snapshot)
	}
}

func TestVerifyUsesOnlyBuiltInArgv(t *testing.T) {
	root := t.TempDir()
	change := Change{Path: "go.mod", Content: []byte("module example.com/test\n\ngo 1.26\n")}
	var commands [][]string
	mysql := Change{Path: "db/mysql/schema.sql", Content: []byte("CREATE TABLE users (id BIGINT);\n")}
	results, err := verifyPlan(context.Background(), root, []Change{change, mysql}, []string{"go", "mysql"}, func(_ context.Context, _ string, argv []string) error {
		commands = append(commands, append([]string(nil), argv...))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(commands, [][]string{{"go", "test", "./..."}}) || len(results) != 2 {
		t.Fatalf("commands=%v results=%v", commands, results)
	}
}

type tarItem struct {
	name, body, link string
	kind             byte
	pax              map[string]string
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func makeArchive(t *testing.T, items []tarItem) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	writer := tar.NewWriter(gzipWriter)
	for _, item := range items {
		kind := item.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		header := &tar.Header{Name: item.name, Typeflag: kind, Mode: 0o644, Size: int64(len(item.body)), Linkname: item.link}
		if kind == tar.TypeXGlobalHeader {
			header = &tar.Header{Typeflag: kind, PAXRecords: item.pax}
		}
		if kind == tar.TypeSymlink {
			header.Size = 0
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Size > 0 {
			if _, err := writer.Write([]byte(item.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
func repositorySnapshot(t *testing.T) SourceSnapshot {
	t.Helper()
	catalog := mustCatalog(t)
	files := map[string][]byte{CatalogPath: mustRead(t, "../../catalog/core-v1.json"), ".editorconfig": mustRead(t, "../../../.editorconfig"), ".gitignore": mustRead(t, "../../../.gitignore")}
	for _, file := range catalog.Files {
		files[file.Source] = mustRead(t, "../../../"+file.Source)
	}
	for _, stack := range catalog.Stacks {
		for _, file := range stack.Files {
			files[file.Source] = mustRead(t, "../../../"+file.Source)
		}
	}
	return SourceSnapshot{Source: testSource, Release: "v1.0.0", ResolvedCommit: testCommit, Files: files}
}
func mustCatalog(t *testing.T) Catalog {
	t.Helper()
	catalog, err := DecodeCatalog(mustRead(t, "../../catalog/core-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
func failureCode(err error) int {
	var failure *Failure
	if errors.As(err, &failure) {
		return failure.Code
	}
	return -1
}
func readTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		if entry.IsDir() {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		result[filepath.ToSlash(relative)] = string(data)
		return nil
	})
	return result
}
func testManifest(t *testing.T, change Change) manifest.Document {
	t.Helper()
	selection := manifest.Selection{Profile: "standard", Stacks: []string{"go"}, ModulePath: "github.com/example/project"}
	selection.ConfigurationHash, _ = manifest.HashSelection(selection)
	document := manifest.Document{SchemaVersion: 1, SokuVersion: "test", Boilerplate: manifest.Boilerplate{Source: testSource, Release: "v1.0.0", ResolvedCommit: testCommit}, Selection: selection, Files: []manifest.File{{Path: change.Path, Owner: "core", Class: "core-managed", ContentMode: "text", BaselineSHA256: change.BaselineSHA256, LifecycleState: "current"}}, Integrations: []manifest.Integration{}}
	if err := manifest.Validate(document); err != nil {
		t.Fatal(err)
	}
	return document
}
