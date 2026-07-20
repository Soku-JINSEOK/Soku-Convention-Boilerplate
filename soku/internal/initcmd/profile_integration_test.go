package initcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	providerSource = "github:example/provider/ai-collaboration"
	providerRef    = "abcdef0123456789abcdef0123456789abcdef01"
)

type staticIntegrationFetcher struct {
	bundle    ProviderBundle
	available bool
	err       error
}

func TestPublishedProfileAndProviderContracts(t *testing.T) {
	if _, err := DecodeProfileIndex(mustRead(t, "../../catalog/index-v2.json")); err != nil {
		t.Fatal(err)
	}
	providerData := mustRead(t, "../../providers/ai-collaboration/provider-v1.json")
	files := map[string][]byte{
		"configuration.schema.json":      mustRead(t, "../../providers/ai-collaboration/configuration.schema.json"),
		"templates/ai-collaboration.yml": mustRead(t, "../../providers/ai-collaboration/templates/ai-collaboration.yml"),
	}
	bundle, err := DecodeProviderBundle(providerData, files)
	if err != nil {
		t.Fatal(err)
	}
	crlfFiles := map[string][]byte{}
	for path, content := range files {
		canonical := []byte(normalizeText(content))
		crlfFiles[path] = bytes.ReplaceAll(canonical, []byte("\n"), []byte("\r\n"))
	}
	if _, err := DecodeProviderBundle(providerData, crlfFiles); err != nil {
		t.Fatalf("CRLF provider bundle: %v", err)
	}
	configurationHash, err := integrationConfigurationHash("../../providers/ai-collaboration/example-config.yml")
	if err != nil || configurationHash != bundle.ConfigurationHash {
		t.Fatalf("example configuration hash = %q, %v", configurationHash, err)
	}
	compiler := jsonschema.NewCompiler()
	for _, contract := range []struct{ schema, instance string }{
		{"../../schema/catalog-index-v2.schema.json", "../../catalog/index-v2.json"},
		{"../../schema/provider-v1.schema.json", "../../providers/ai-collaboration/provider-v1.json"},
	} {
		compiled, err := compiler.Compile(contract.schema)
		if err != nil {
			t.Fatal(err)
		}
		instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(mustRead(t, contract.instance)))
		if err != nil {
			t.Fatal(err)
		}
		if err := compiled.Validate(instance); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProviderBundleLegacyRefCompatibility(t *testing.T) {
	configurationHash := strings.Repeat("a", 64)
	files := validProviderBundle(configurationHash).Files
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile("../../schema/provider-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		mutate  func(map[string]any)
		wantErr bool
	}{
		{name: "omitted", mutate: func(object map[string]any) { delete(object, "ref") }},
		{name: "well-formed", mutate: func(object map[string]any) { object["ref"] = providerRef }},
		{name: "malformed", mutate: func(object map[string]any) { object["ref"] = "" }, wantErr: true},
		{name: "unknown-field", mutate: func(object map[string]any) { object["unknown"] = true }, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(validProviderBundle(configurationHash))
			if err != nil {
				t.Fatal(err)
			}
			var object map[string]any
			if err := json.Unmarshal(data, &object); err != nil {
				t.Fatal(err)
			}
			test.mutate(object)
			data, err = json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			schemaErr := schema.Validate(instance)
			if test.wantErr && schemaErr == nil {
				t.Fatal("provider schema accepted an invalid legacy shape")
			}
			if !test.wantErr && schemaErr != nil {
				t.Fatal(schemaErr)
			}
			_, err = DecodeProviderBundle(data, files)
			if test.wantErr && failureCode(err) != 5 {
				t.Fatalf("error = %v, want compatibility failure", err)
			}
			if !test.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func (fetcher staticIntegrationFetcher) FetchIntegration(context.Context, string, string) (ProviderBundle, bool, error) {
	return fetcher.bundle, fetcher.available, fetcher.err
}

func TestProfileCompositionAndLegacyFallback(t *testing.T) {
	snapshot := indexedSnapshot(t)
	catalog := mustCatalog(t)
	counts := map[string]int{}
	for _, profile := range []string{ProfileBootstrap, ProfileStandard, ProfileScaled} {
		config := Config{SchemaVersion: 1, Profile: profile, Stacks: []string{"go"}, ModulePath: "github.com/example/project"}
		changes, err := renderProfileCatalog(snapshot, catalog, config)
		if err != nil {
			t.Fatalf("profile %s: %v", profile, err)
		}
		counts[profile] = len(changes)
		if profile == ProfileScaled {
			assertChangePath(t, changes, "AGENTS.md")
			assertChangePath(t, changes, ".github/CODEOWNERS")
		}
	}
	if counts[ProfileBootstrap] >= counts[ProfileStandard] || counts[ProfileStandard] >= counts[ProfileScaled] {
		t.Fatalf("profile counts are not linearly composed: %v", counts)
	}
	legacy := repositorySnapshot(t)
	if _, err := renderProfileCatalog(legacy, catalog, Config{Profile: ProfileBootstrap, Stacks: []string{"go"}, ModulePath: "github.com/example/project"}); failureCode(err) != 5 {
		t.Fatalf("legacy bootstrap error = %v", err)
	}
	if _, err := renderProfileCatalog(legacy, catalog, Config{Profile: ProfileStandard, Stacks: []string{"go"}, ModulePath: "github.com/example/project"}); err != nil {
		t.Fatalf("legacy standard = %v", err)
	}
}

func TestProfileTransitionUsesOuterTransaction(t *testing.T) {
	snapshot := indexedSnapshot(t)
	root := initializeRelease(t, snapshot)
	report, err := RunTransition(context.Background(), TransitionOptions{Root: root, TargetRelease: "v1.0.0", TargetProfile: ProfileScaled, Yes: true, SokuVersion: "test"}, staticFetcher{snapshot: snapshot}, true)
	if err != nil || report.State != "applied" {
		t.Fatalf("profile upgrade = %#v, %v", report, err)
	}
	document, err := manifest.NewStore(root).Load()
	if err != nil || document.Selection.Profile != ProfileScaled {
		t.Fatalf("manifest profile = %q, %v", document.Selection.Profile, err)
	}
	for _, path := range []string{"AGENTS.md", ".github/CODEOWNERS"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("scaled output %s: %v", path, err)
		}
	}
}

func TestIntegrationPendingConnectedAndRollback(t *testing.T) {
	snapshot := indexedSnapshot(t)
	configPath := filepath.Join(t.TempDir(), "ai.yml")
	writeTestFile(t, configPath, "review_mode: advisory\nlanguages: [en, ko, ja]\n")
	configurationHash, err := integrationConfigurationHash(configPath)
	if err != nil {
		t.Fatal(err)
	}
	bundle := validProviderBundle(configurationHash)

	t.Run("pending", func(t *testing.T) {
		root := t.TempDir()
		options := integrationInitOptions(root, configPath)
		report, err := Run(context.Background(), options, staticFetcher{snapshot: snapshot})
		if err != nil || report.Integrations[0].LifecycleState != "pending" {
			t.Fatalf("pending init = %#v, %v", report, err)
		}
		document, err := manifest.NewStore(root).Load()
		if err != nil || document.Integrations[0].LifecycleState != "pending" {
			t.Fatalf("pending manifest = %#v, %v", document.Integrations, err)
		}
		request := readProjectPath(t, root, document.Integrations[0].ManagedFiles[0])
		assertProviderRequest(t, request, configurationHash)
	})

	for _, test := range []struct {
		name      string
		legacyRef *string
	}{
		{name: "connected-without-legacy-ref"},
		{name: "connected-with-matching-legacy-ref", legacyRef: pointerTo(providerRef)},
		{name: "connected-with-mismatching-legacy-ref", legacyRef: pointerTo(strings.Repeat("1", 40))},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			bundle := bundle
			bundle.Ref = test.legacyRef
			options := integrationInitOptions(root, configPath)
			options.IntegrationFetcher = staticIntegrationFetcher{bundle: bundle, available: true}
			report, err := Run(context.Background(), options, staticFetcher{snapshot: snapshot})
			if err != nil || report.Integrations[0].LifecycleState != "connected" {
				t.Fatalf("connected init = %#v, %v", report, err)
			}
			document, err := manifest.NewStore(root).Load()
			if err != nil || document.Integrations[0].Ref != providerRef {
				t.Fatalf("authoritative manifest ref = %q, %v", document.Integrations[0].Ref, err)
			}
			request := readProjectPath(t, root, ".github/soku/integrations/ai-collaboration.json")
			assertProviderRequest(t, request, configurationHash)
			if !strings.Contains(string(readProjectPath(t, root, ".github/ai-collaboration.yml")), "advisory") {
				t.Fatal("connected provider output is missing")
			}
		})
	}

	for _, test := range []struct {
		name   string
		mutate func(*ProviderBundle)
	}{
		{name: "configuration-hash-mismatch", mutate: func(bundle *ProviderBundle) { bundle.ConfigurationHash = strings.Repeat("b", 64) }},
		{name: "source-mismatch", mutate: func(bundle *ProviderBundle) { bundle.Source = "github:other/provider/ai-collaboration" }},
	} {
		t.Run(test.name+"-remains-pending", func(t *testing.T) {
			root := t.TempDir()
			bundle := bundle
			test.mutate(&bundle)
			options := integrationInitOptions(root, configPath)
			options.IntegrationFetcher = staticIntegrationFetcher{bundle: bundle, available: true}
			report, err := Run(context.Background(), options, staticFetcher{snapshot: snapshot})
			if err != nil || report.Integrations[0].LifecycleState != "pending" {
				t.Fatalf("pending mismatch = %#v, %v", report, err)
			}
		})
	}

	t.Run("pending-to-connected-and-rollback", func(t *testing.T) {
		root := t.TempDir()
		if _, err := Run(context.Background(), integrationInitOptions(root, configPath), staticFetcher{snapshot: snapshot}); err != nil {
			t.Fatal(err)
		}
		before := readTree(t, root)
		options := TransitionOptions{Root: root, TargetRelease: "v1.0.0", Yes: true, IntegrationSource: providerSource, IntegrationRef: providerRef, IntegrationConfigPath: configPath, IntegrationFetcher: staticIntegrationFetcher{bundle: bundle, available: true}, ApplyHook: func(stage, _ string) error {
			if stage == "before-manifest" {
				return errors.New("provider rollback injection")
			}
			return nil
		}}
		if _, err := RunTransition(context.Background(), options, staticFetcher{snapshot: snapshot}, true); failureCode(err) != 7 {
			t.Fatalf("provider rollback error = %v", err)
		}
		if !mapsEqual(before, readTree(t, root)) {
			t.Fatal("provider rollback changed the previous lifecycle state")
		}
		options.ApplyHook = nil
		report, err := RunTransition(context.Background(), options, staticFetcher{snapshot: snapshot}, true)
		if err != nil || report.State != "applied" {
			t.Fatalf("provider connection = %#v, %v", report, err)
		}
		document, err := manifest.NewStore(root).Load()
		if err != nil || document.Integrations[0].LifecycleState != "connected" {
			t.Fatalf("connected manifest = %#v, %v", document.Integrations, err)
		}
	})
}

func TestProviderRejectsExecutableEscapeSecretAndOwnershipConflict(t *testing.T) {
	configurationHash := strings.Repeat("a", 64)
	bundle := validProviderBundle(configurationHash)
	bundleData, _ := json.Marshal(bundle)
	var object map[string]any
	_ = json.Unmarshal(bundleData, &object)
	object["script"] = "run.sh"
	malicious, _ := json.Marshal(object)
	if _, err := DecodeProviderBundle(malicious, bundle.Files); failureCode(err) != 5 {
		t.Fatalf("script field error = %v", err)
	}
	bundle.Outputs[0].Path = "../escape.yml"
	if err := validateProviderBundle(bundle); failureCode(err) != 5 {
		t.Fatalf("escape error = %v", err)
	}
	bundle = validProviderBundle(configurationHash)
	bundle.Outputs[0].Path = "tools/provider.sh"
	if err := validateProviderBundle(bundle); failureCode(err) != 5 {
		t.Fatalf("executable error = %v", err)
	}

	secretConfig := filepath.Join(t.TempDir(), "secret.yml")
	writeTestFile(t, secretConfig, "token: should-not-persist\n")
	if _, err := integrationConfigurationHash(secretConfig); failureCode(err) != 2 {
		t.Fatalf("secret config error = %v", err)
	}

	snapshot := indexedSnapshot(t)
	configPath := filepath.Join(t.TempDir(), "ai.yml")
	writeTestFile(t, configPath, "review_mode: advisory\n")
	hash, _ := integrationConfigurationHash(configPath)
	bundle = validProviderBundle(hash)
	bundle.CompatibleProfiles = []string{ProfileBootstrap}
	if _, err := planIntegration(context.Background(), providerSource, providerRef, configPath, ProfileStandard, staticIntegrationFetcher{bundle: bundle, available: true}); failureCode(err) != 5 {
		t.Fatalf("unsupported profile error = %v", err)
	}
	bundle = validProviderBundle(hash)
	bundle.Outputs[0].Path = "profile.go"
	root := t.TempDir()
	options := integrationInitOptions(root, configPath)
	options.IntegrationFetcher = staticIntegrationFetcher{bundle: bundle, available: true}
	if _, err := Run(context.Background(), options, staticFetcher{snapshot: snapshot}); failureCode(err) != 4 {
		t.Fatalf("ownership conflict error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".soku")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("ownership conflict wrote lifecycle state")
	}
}

func indexedSnapshot(t *testing.T) SourceSnapshot {
	t.Helper()
	snapshot := repositorySnapshot(t)
	snapshot.Files[ProfileIndexPath] = mustRead(t, "../../catalog/index-v2.json")
	snapshot.Files["AGENTS.md"] = mustRead(t, "../../../AGENTS.md")
	snapshot.Files[".github/CODEOWNERS"] = mustRead(t, "../../../.github/CODEOWNERS")
	return snapshot
}

func integrationInitOptions(root, configPath string) Options {
	return Options{Root: root, Explicit: Explicit{Source: testSource, Release: "v1.0.0", Stacks: []string{"go"}, Profile: ProfileStandard, ModulePath: "github.com/example/project", SourceSet: true, ReleaseSet: true, StacksSet: true, ProfileSet: true, ModulePathSet: true}, Yes: true, SokuVersion: "test", IntegrationSource: providerSource, IntegrationRef: providerRef, IntegrationConfigPath: configPath}
}

func validProviderBundle(configurationHash string) ProviderBundle {
	return ProviderBundle{SchemaVersion: 1, ID: "ai-collaboration", Source: providerSource, ProviderAPIVersion: "1", ProviderSchemaVersion: "1", CompatibleSoku: ">=0.1.0 <2.0.0", ConfigurationSchemaHash: "ca3d163bab055381827226140568f3bef7eaac187cebd76878e0b63e9e442356", ConfigurationHash: configurationHash, CompatibleProfiles: []string{ProfileBootstrap, ProfileScaled, ProfileStandard}, Outputs: []ProviderOutput{{Template: "templates/ai.yml", Path: ".github/ai-collaboration.yml", ContentMode: "text"}}, Files: map[string][]byte{"configuration.schema.json": []byte("{}\n"), "templates/ai.yml": []byte("mode: advisory\n")}}
}

func assertProviderRequest(t *testing.T, request []byte, configurationHash string) {
	t.Helper()
	var artifact map[string]any
	if err := json.Unmarshal(request, &artifact); err != nil {
		t.Fatal(err)
	}
	expected := map[string]any{
		"schema_version":     float64(1),
		"id":                 "ai-collaboration",
		"source":             "https://github.com/example/provider/ai-collaboration",
		"ref":                providerRef,
		"configuration_hash": configurationHash,
	}
	if len(artifact) != len(expected) {
		t.Fatalf("pending artifact keys = %v", artifact)
	}
	for key, value := range expected {
		if artifact[key] != value {
			t.Fatalf("pending artifact %s = %#v, want %#v", key, artifact[key], value)
		}
	}
	serialized := strings.ToLower(string(request))
	for _, forbidden := range []string{"review_mode", "languages", "advisory", "secret", "token", "password"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("pending artifact persisted forbidden payload %q", forbidden)
		}
	}
}

func pointerTo(value string) *string { return &value }

func assertChangePath(t *testing.T, changes []Change, path string) {
	t.Helper()
	for _, change := range changes {
		if change.Path == path {
			return
		}
	}
	t.Fatalf("change %s is missing", path)
}

func readProjectPath(t *testing.T, root, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mapsEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
