package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	testCommit = "0123456789abcdef0123456789abcdef01234567"
	testHash   = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestPublishedFixturesMatchManifestContract(t *testing.T) {
	schemaData, err := os.ReadFile(filepath.Join("..", "..", "schema", "manifest-v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("unexpected schema dialect: %v", schema["$schema"])
	}
	compiler := jsonschema.NewCompiler()
	compiled, err := compiler.Compile(filepath.Join("..", "..", "schema", "manifest-v1.schema.json"))
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}

	valid, err := filepath.Glob(filepath.Join("..", "..", "testdata", "manifest-v1", "valid", "*.json"))
	if err != nil || len(valid) == 0 {
		t.Fatalf("valid fixtures: files=%v err=%v", valid, err)
	}
	for _, name := range valid {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Decode(data); err != nil {
			t.Errorf("valid fixture %s: %v", name, err)
		}
		instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if err := compiled.Validate(instance); err != nil {
			t.Errorf("valid fixture %s does not match schema: %v", name, err)
		}
	}

	invalid, err := filepath.Glob(filepath.Join("..", "..", "testdata", "manifest-v1", "invalid", "*.json"))
	if err != nil || len(invalid) == 0 {
		t.Fatalf("invalid fixtures: files=%v err=%v", invalid, err)
	}
	for _, name := range invalid {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Decode(data); err == nil {
			t.Errorf("invalid fixture %s was accepted", name)
		}
		instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if err := compiled.Validate(instance); err == nil {
			t.Errorf("invalid fixture %s matches schema", name)
		}
	}
}

func TestPublishedV2FixturesMatchManifestContract(t *testing.T) {
	schemaPath := filepath.Join("..", "..", "schema", "manifest-v2.schema.json")
	compiler := jsonschema.NewCompiler()
	compiled, err := compiler.Compile(schemaPath)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	for _, group := range []struct {
		name      string
		wantValid bool
	}{
		{name: "valid", wantValid: true},
		{name: "invalid", wantValid: false},
	} {
		files, err := filepath.Glob(filepath.Join("..", "..", "testdata", "manifest-v2", group.name, "*.json"))
		if err != nil || len(files) == 0 {
			t.Fatalf("%s fixtures: files=%v err=%v", group.name, files, err)
		}
		for _, name := range files {
			data, readErr := os.ReadFile(name)
			if readErr != nil {
				t.Fatal(readErr)
			}
			_, decodeErr := Decode(data)
			instance, unmarshalErr := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if unmarshalErr != nil {
				t.Fatal(unmarshalErr)
			}
			schemaErr := compiled.Validate(instance)
			if group.wantValid && (decodeErr != nil || schemaErr != nil) {
				t.Errorf("valid fixture %s: decode=%v schema=%v", name, decodeErr, schemaErr)
			}
			if !group.wantValid && decodeErr == nil {
				t.Errorf("invalid fixture %s was accepted: decode=%v schema=%v", name, decodeErr, schemaErr)
			}
		}
	}
}

func TestPublishedV3FixturesMatchManifestContract(t *testing.T) {
	schemaPath := filepath.Join("..", "..", "schema", "manifest-v3.schema.json")
	compiler := jsonschema.NewCompiler()
	compiled, err := compiler.Compile(schemaPath)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	for _, group := range []struct {
		name      string
		wantValid bool
	}{
		{name: "valid", wantValid: true},
		{name: "invalid", wantValid: false},
	} {
		files, err := filepath.Glob(filepath.Join("..", "..", "testdata", "manifest-v3", group.name, "*.json"))
		if err != nil || len(files) == 0 {
			t.Fatalf("%s fixtures: files=%v err=%v", group.name, files, err)
		}
		for _, name := range files {
			data, readErr := os.ReadFile(name)
			if readErr != nil {
				t.Fatal(readErr)
			}
			_, decodeErr := Decode(data)
			instance, unmarshalErr := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if unmarshalErr != nil {
				t.Fatal(unmarshalErr)
			}
			schemaErr := compiled.Validate(instance)
			if group.wantValid && (decodeErr != nil || schemaErr != nil) {
				t.Errorf("valid fixture %s: decode=%v schema=%v", name, decodeErr, schemaErr)
			}
			if !group.wantValid && decodeErr == nil {
				t.Errorf("invalid fixture %s was accepted: schema=%v", name, schemaErr)
			}
		}
	}
}

func TestSelectionHashIsReproducibleAndDetectsTampering(t *testing.T) {
	selection := Selection{Profile: "standard", Stacks: []string{"go"}, ModulePath: "github.com/example/project"}
	selection.ConfigurationHash, _ = HashSelection(selection)
	document := validDocument()
	document.Selection = selection
	if err := Validate(document); err != nil {
		t.Fatal(err)
	}
	document.Selection.ModulePath = "github.com/example/other"
	if err := Validate(document); err == nil {
		t.Fatal("selection changed without updating configuration_hash")
	}
}

func TestLegacySelectionHashRemainsByteStable(t *testing.T) {
	selection := Selection{
		Profile: "team", Stacks: []string{"go", "postgresql"}, ModulePath: "github.com/example/project",
	}
	got, err := HashSelection(selection)
	if err != nil {
		t.Fatal(err)
	}
	const want = "2f5a3d95e02474fb266185f2d0720d709da5be41cafc453ba2c1d3e419e2ca79"
	if got != want {
		t.Fatalf("legacy selection hash = %s, want %s", got, want)
	}
}

func TestMarshalCanonicalSortsAndIsDeterministic(t *testing.T) {
	document := validDocument()
	document.Selection.Stacks = []string{"z", "a"}
	sort.Strings(document.Selection.Stacks)
	document.Selection.ConfigurationHash, _ = HashSelection(document.Selection)
	document.Files = []File{
		managedFile("z.txt", "core", "core-managed"),
		managedFile("a.txt", "core", "core-managed"),
	}
	document.Integrations = []Integration{}
	first, err := MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || !bytes.HasSuffix(first, []byte("\n")) {
		t.Fatalf("serialization is not deterministic: %q / %q", first, second)
	}
	if strings.Index(string(first), `"path": "a.txt"`) > strings.Index(string(first), `"path": "z.txt"`) {
		t.Fatalf("files were not sorted: %s", first)
	}
	decoded, err := Decode(first)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Selection.Stacks, []string{"a", "z"}) {
		t.Fatalf("stacks = %v", decoded.Selection.Stacks)
	}
}

func TestHashContentCanonicalization(t *testing.T) {
	lf, err := HashContent([]byte("\ufeffline 1  \nline 2\n"), "text")
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range [][]byte{
		[]byte("\ufeffline 1  \r\nline 2\r\n"),
		[]byte("\ufeffline 1  \rline 2\r"),
	} {
		got, err := HashContent(input, "text")
		if err != nil || got != lf {
			t.Fatalf("text hash = %q, %v; want %q", got, err, lf)
		}
	}
	withoutFinalNewline, _ := HashContent([]byte("\ufeffline 1  \nline 2"), "text")
	if withoutFinalNewline == lf {
		t.Fatal("final newline was changed")
	}
	if _, err := HashContent([]byte{0xff}, "text"); err == nil {
		t.Fatal("invalid UTF-8 was accepted")
	}
	binary, _ := HashContent([]byte{'a', '\r', '\n'}, "binary")
	binaryLF, _ := HashContent([]byte{'a', '\n'}, "binary")
	if binary == binaryLF {
		t.Fatal("binary bytes were normalized")
	}
}

func TestValidationRejectsUnsafePortableState(t *testing.T) {
	paths := []string{"", "/absolute", "../escape", "a/../escape", `a\\..\\escape`, ".git/config", ".SOKU/state", "a/CON.txt", "a/trailing. ", "a:b"}
	for _, value := range paths {
		if err := ValidatePath(value); err == nil {
			t.Errorf("path %q was accepted", value)
		}
	}
	if err := ValidatePath(".soku/ci-cd/caller.yml"); err == nil {
		t.Error("reserved CI/CD caller path was accepted by the ordinary path validator")
	}
	if err := ValidateManagedPath(".soku/ci-cd/caller.yml"); err != nil {
		t.Fatalf("managed CI/CD caller path was rejected: %v", err)
	}
	if err := ValidateManagedPath(".soku/other-state.json"); err == nil {
		t.Error("unreserved .soku path was accepted as a managed path")
	}
	if err := ValidateComponentPath(".soku/ci-cd-decision.yml"); err != nil {
		t.Fatalf("CI/CD component configuration path was rejected: %v", err)
	}
	if err := ValidateComponentPath(".soku/other-config.yml"); err == nil {
		t.Error("unreserved .soku path was accepted as a component configuration path")
	}

	tests := []struct {
		name   string
		mutate func(*Document)
	}{
		{name: "credential source", mutate: func(document *Document) { document.Boilerplate.Source = "https://token@example.com/repo" }},
		{name: "credential query", mutate: func(document *Document) { document.Boilerplate.Source = "https://example.com/repo?token=value" }},
		{name: "non-https source", mutate: func(document *Document) { document.Boilerplate.Source = "http://example.com/repo" }},
		{name: "fragment source", mutate: func(document *Document) { document.Boilerplate.Source = "https://example.com/repo#main" }},
		{name: "absolute source", mutate: func(document *Document) { document.Boilerplate.Source = "file:///tmp/repo" }},
		{name: "invalid integration id", mutate: func(document *Document) {
			document.Integrations = []Integration{validIntegration("provider:id", nil)}
		}},
		{name: "case collision", mutate: func(document *Document) {
			document.Files = []File{managedFile("A.txt", "core", "core-managed"), managedFile("a.txt", "core", "core-managed")}
		}},
		{name: "unknown provider owner", mutate: func(document *Document) {
			document.Files = []File{managedFile("provider.txt", "missing", "provider-managed")}
		}},
		{name: "unreferenced provider file", mutate: func(document *Document) {
			document.Integrations = []Integration{validIntegration("provider", nil)}
			document.Files = []File{managedFile("provider.txt", "provider", "provider-managed")}
		}},
		{name: "duplicate integration reference", mutate: func(document *Document) {
			document.Integrations = []Integration{validIntegration("a", []string{"shared.txt"}), validIntegration("b", []string{"shared.txt"})}
			document.Files = []File{managedFile("shared.txt", "a", "provider-managed")}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := validDocument()
			test.mutate(&document)
			if err := Validate(document); err == nil {
				t.Fatal("unsafe manifest was accepted")
			}
		})
	}

	data, err := MarshalCanonical(validDocument())
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["configuration"] = map[string]any{"password": "secret"}
	data, err = json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(data); err == nil {
		t.Fatal("raw configuration field was accepted")
	}
}

func TestDecodeDistinguishesMalformedAndUnsupportedSchema(t *testing.T) {
	if _, err := Decode([]byte(`{"files":[]}`)); err == nil {
		t.Fatal("missing schema_version was accepted")
	} else {
		var unsupported *UnsupportedSchemaError
		if errors.As(err, &unsupported) {
			t.Fatalf("missing schema_version was treated as compatibility: %v", err)
		}
	}
	if _, err := Decode([]byte(`{"schema_version":4}`)); err == nil {
		t.Fatal("unsupported schema was accepted")
	} else {
		var unsupported *UnsupportedSchemaError
		if !errors.As(err, &unsupported) || unsupported.Version != 4 {
			t.Fatalf("error = %T %v", err, err)
		}
	}
}

func TestV1AndV2ComponentBoundary(t *testing.T) {
	document := validDocument()
	document.Components = []Component{{
		ID: "docs-manual", CatalogVersion: "1", ConfigurationPath: "docs/manual/capture.yml",
	}}
	if err := Validate(document); err == nil {
		t.Fatal("manifest v1 accepted component metadata")
	}
	document.SchemaVersion = SchemaVersionV2
	if err := Validate(document); err != nil {
		t.Fatalf("manifest v2 rejected component metadata: %v", err)
	}
	data, err := MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(data)
	if err != nil || !reflect.DeepEqual(decoded.Components, document.Components) {
		t.Fatalf("v2 round trip: document=%#v err=%v", decoded, err)
	}
}

func validDocument() Document {
	document := Document{
		SchemaVersion: SchemaVersion,
		SokuVersion:   "v0.2.0",
		Boilerplate: Boilerplate{
			Source: "https://github.com/example/boilerplate", Release: "v1.0.0", ResolvedCommit: testCommit,
		},
		Selection: Selection{Profile: "team", Stacks: []string{}},
		Files:     []File{}, Integrations: []Integration{},
	}
	document.Selection.ConfigurationHash, _ = HashSelection(document.Selection)
	return document
}

func managedFile(name, owner, class string) File {
	return File{Path: name, Owner: owner, Class: class, ContentMode: "text", BaselineSHA256: testHash, LifecycleState: "current"}
}

func validIntegration(id string, files []string) Integration {
	return Integration{
		ID: id, Source: "https://github.com/example/provider", Ref: testCommit,
		ProviderAPIVersion: "1", ProviderSchemaVersion: "1", ConfigurationHash: testHash,
		LifecycleState: "connected", ManagedFiles: files,
	}
}
