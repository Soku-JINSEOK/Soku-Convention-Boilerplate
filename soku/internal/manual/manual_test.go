package manual

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	manualTestCommit = "0123456789abcdef0123456789abcdef01234567"
)

func TestPublishedExampleBuildsDeterministicPlan(t *testing.T) {
	first, err := BuildPlan(".", "assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildPlan(".", "assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("plans differ:\n%#v\n%#v", first, second)
	}
	if first.Authenticity != "runtime-authentic" || len(first.Captures) != 1 ||
		first.Captures[0].OutputPath != "docs/manual/captures/dashboard-ready.png" {
		t.Fatalf("unexpected plan: %#v", first)
	}
}

func TestAllPublishedProviderExamplesAreValid(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("assets", "examples", "*.yml"))
	if err != nil || len(files) != 3 {
		t.Fatalf("examples=%v err=%v", files, err)
	}
	for _, name := range files {
		portable := filepath.ToSlash(name)
		if _, err := LoadConfig(".", portable); err != nil {
			t.Errorf("%s: %v", portable, err)
		}
	}
}

func TestPublishedCaptureSchemasCompileAndValidateExample(t *testing.T) {
	configurationSchema := filepath.Join("..", "..", "schema", "manual-capture-v1.schema.json")
	reportSchema := filepath.Join("..", "..", "schema", "manual-capture-report-v1.schema.json")
	compiler := jsonschema.NewCompiler()
	compiled, err := compiler.Compile(configurationSchema)
	if err != nil {
		t.Fatalf("compile configuration schema: %v", err)
	}
	if _, err := compiler.Compile(reportSchema); err != nil {
		t.Fatalf("compile report schema: %v", err)
	}
	loaded, err := LoadConfig(".", "assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(loaded.Config)
	if err != nil {
		t.Fatal(err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := compiled.Validate(instance); err != nil {
		t.Fatalf("example does not match configuration schema: %v", err)
	}
}

func TestComponentCatalogMatchesEmbeddedCoreOutputs(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "components", "docs-manual", "component-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		SchemaVersion      int      `json:"schema_version"`
		ID                 string   `json:"id"`
		CatalogVersion     string   `json:"catalog_version"`
		ManifestSchema     int      `json:"manifest_schema"`
		CoreManagedOutputs []string `json:"core_managed_outputs"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.SchemaVersion != 1 || catalog.ID != componentID ||
		catalog.CatalogVersion != catalogVersion || catalog.ManifestSchema != manifest.SchemaVersionV2 {
		t.Fatalf("unexpected component catalog metadata: %#v", catalog)
	}
	assets, err := installationAssets()
	if err != nil {
		t.Fatal(err)
	}
	outputs := make([]string, 0, len(assets))
	for output := range assets {
		outputs = append(outputs, output)
	}
	sort.Strings(outputs)
	if !reflect.DeepEqual(outputs, catalog.CoreManagedOutputs) {
		t.Fatalf("catalog output mismatch:\nembedded=%v\ncatalog=%v", outputs, catalog.CoreManagedOutputs)
	}
}

func TestPublishedProviderProfileMatchesInstalledCatalog(t *testing.T) {
	published, err := os.ReadFile(filepath.Join("..", "..", "components", "docs-manual", "egress-profiles-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	embedded, err := componentAssets.ReadFile("assets/catalogs/provider-egress-profiles-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(published, embedded) {
		t.Fatal("published and installed provider egress profiles differ")
	}
}

func TestStrictConfigRejectsUnknownDuplicateSecretAndUnsafeInputs(t *testing.T) {
	valid, err := os.ReadFile("assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(string) string
	}{
		{name: "unknown", mutate: func(value string) string { return value + "\nunknown: true\n" }},
		{name: "duplicate", mutate: func(value string) string {
			return strings.Replace(value, "schema_version: 1", "schema_version: 1\nschema_version: 1", 1)
		}},
		{name: "secret", mutate: func(value string) string { return value + "\napi_key: example\n" }},
		{name: "unsafe path", mutate: func(value string) string {
			return strings.Replace(value, "static_directory: dist", "static_directory: ../dist", 1)
		}},
		{name: "hosted mode", mutate: func(value string) string { return strings.Replace(value, "mode: local-manual", "mode: hosted", 1) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "capture.yml")
			if err := os.WriteFile(path, []byte(test.mutate(string(valid))), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadConfig(root, "capture.yml"); err == nil {
				t.Fatal("unsafe configuration was accepted")
			}
		})
	}
}

func TestDoctorReportsUnownedPlannedOutput(t *testing.T) {
	root := t.TempDir()
	config, err := os.ReadFile("assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "docs", "manual", "capture.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "docs", "manual", "generated-index.md")
	if err := os.WriteFile(output, []byte("user-authored\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := Doctor(context.Background(), root, "docs/manual/capture.yml", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range report.Checks {
		if check.ID == "output:docs/manual/generated-index.md" && check.Status == "fail" {
			return
		}
	}
	t.Fatalf("doctor did not report output collision: %#v", report.Checks)
}

func TestManualInitDryRunAndV1ToV2Apply(t *testing.T) {
	root := t.TempDir()
	writeV1Manifest(t, root)
	before, err := os.ReadFile(filepath.Join(root, ".soku", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	dryRun, err := Init(InitOptions{
		Root: root, ConfigPath: "docs/manual/capture.yml", DryRun: true, SokuVersion: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if dryRun.State != "dry-run" || dryRun.ManifestSchemaBefore != 1 || dryRun.ManifestSchemaAfter != 2 || len(dryRun.Changes) == 0 {
		t.Fatalf("unexpected dry-run: %#v", dryRun)
	}
	afterDryRun, _ := os.ReadFile(filepath.Join(root, ".soku", "manifest.json"))
	if !bytes.Equal(before, afterDryRun) {
		t.Fatal("dry-run changed the manifest")
	}
	for _, change := range dryRun.Changes {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(change.Path))); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("dry-run wrote %s", change.Path)
		}
	}

	applied, err := Init(InitOptions{
		Root: root, ConfigPath: "docs/manual/capture.yml", Yes: true, SokuVersion: "test-v2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if applied.State != "applied" {
		t.Fatalf("state=%s", applied.State)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if document.SchemaVersion != 2 || document.SokuVersion != "test-v2" ||
		len(document.Components) != 1 || document.Components[0].ID != "docs-manual" ||
		document.Components[0].ConfigurationPath != "docs/manual/capture.yml" {
		t.Fatalf("unexpected manifest: %#v", document)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "manual", "capture.yml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("project-owned capture.yml was generated")
	}
	noOp, err := Init(InitOptions{
		Root: root, ConfigPath: "docs/manual/capture.yml", DryRun: true, SokuVersion: "test-v2",
	})
	if err != nil || noOp.State != "no-op" {
		t.Fatalf("rerun=%#v err=%v", noOp, err)
	}
}

func TestManualInitCollisionAndRollbackRestoreExactV1(t *testing.T) {
	t.Run("collision", func(t *testing.T) {
		root := t.TempDir()
		writeV1Manifest(t, root)
		target := filepath.Join(root, "tools", "manual-capture", "package.json")
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Init(InitOptions{Root: root, ConfigPath: "docs/manual/capture.yml", Yes: true}); err == nil {
			t.Fatal("collision was accepted")
		}
	})

	t.Run("rollback", func(t *testing.T) {
		root := t.TempDir()
		writeV1Manifest(t, root)
		manifestPath := filepath.Join(root, ".soku", "manifest.json")
		before, _ := os.ReadFile(manifestPath)
		_, err := Init(InitOptions{
			Root: root, ConfigPath: "docs/manual/capture.yml", Yes: true,
			ApplyHook: func(stage, path string) error {
				if stage == "before-manifest" {
					return errors.New("injected failure")
				}
				return nil
			},
		})
		var manualError *Error
		if !errors.As(err, &manualError) || manualError.Code != 7 {
			t.Fatalf("error=%T %v", err, err)
		}
		after, _ := os.ReadFile(manifestPath)
		if !bytes.Equal(before, after) {
			t.Fatal("rollback did not restore the exact v1 manifest bytes")
		}
		assets, assetErr := installationAssets()
		if assetErr != nil {
			t.Fatal(assetErr)
		}
		for output := range assets {
			if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(output))); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("rollback left component output %s", output)
			}
		}
	})
}

func TestManualInitPreservesManifestV3(t *testing.T) {
	root := t.TempDir()
	projectPath := "policy.txt"
	if err := os.WriteFile(filepath.Join(root, projectPath), []byte("project owned\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	selection := manifest.Selection{
		Profile: "standard", Stacks: []string{}, ProjectOwnedOverrides: []string{projectPath},
	}
	selection.ConfigurationHash, _ = manifest.HashSelection(selection)
	document := manifest.Document{
		SchemaVersion: manifest.SchemaVersionV3,
		SokuVersion:   "v0.3.0",
		Boilerplate: manifest.Boilerplate{
			Source: "https://github.com/example/boilerplate", Release: "v1.0.0", ResolvedCommit: manualTestCommit,
		},
		Selection: selection,
		Files: []manifest.File{{
			Path: projectPath, Owner: "project", Class: "project-owned", LifecycleState: "unmanaged-expected",
		}},
		Integrations: []manifest.Integration{},
	}
	if err := manifest.NewStore(root).Write(document); err != nil {
		t.Fatal(err)
	}
	report, err := Init(InitOptions{
		Root: root, ConfigPath: "docs/manual/capture.yml", Yes: true, SokuVersion: "v0.3.0",
	})
	if err != nil || report.ManifestSchemaBefore != 3 || report.ManifestSchemaAfter != 3 {
		t.Fatalf("report = %#v, %v", report, err)
	}
	applied, err := manifest.NewStore(root).Load()
	if err != nil || applied.SchemaVersion != manifest.SchemaVersionV3 ||
		len(applied.Selection.ProjectOwnedOverrides) != 1 || applied.Selection.ProjectOwnedOverrides[0] != projectPath {
		t.Fatalf("manifest = %#v, %v", applied, err)
	}
	if got, err := os.ReadFile(filepath.Join(root, projectPath)); err != nil || string(got) != "project owned\n" {
		t.Fatalf("project-owned path changed: %v", err)
	}
}

func writeV1Manifest(t *testing.T, root string) {
	t.Helper()
	selection := manifest.Selection{Profile: "standard", Stacks: []string{}}
	selection.ConfigurationHash, _ = manifest.HashSelection(selection)
	document := manifest.Document{
		SchemaVersion: 1,
		SokuVersion:   "test-v1",
		Boilerplate: manifest.Boilerplate{
			Source: "https://github.com/example/boilerplate", Release: "v1.0.0",
			ResolvedCommit: manualTestCommit,
		},
		Selection: selection, Files: []manifest.File{}, Integrations: []manifest.Integration{},
	}
	if err := manifest.NewStore(root).Write(document); err != nil {
		t.Fatal(err)
	}
	// Re-encode through JSON to make exact restoration assertions independent
	// from the Store's temporary file implementation.
	data, err := os.ReadFile(filepath.Join(root, ".soku", "manifest.json"))
	if err != nil || !json.Valid(data) {
		t.Fatalf("manifest setup failed: %v", err)
	}
}
