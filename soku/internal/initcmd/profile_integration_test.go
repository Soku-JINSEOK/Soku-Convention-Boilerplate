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

	providerSchema, err := compiler.Compile("../../schema/provider-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var providerObject map[string]any
	if err := json.Unmarshal(providerData, &providerObject); err != nil {
		t.Fatal(err)
	}
	delete(providerObject, "ref")
	if err := providerSchema.Validate(providerObject); err != nil {
		t.Fatalf("provider without legacy ref: %v", err)
	}
	providerObject["ref"] = "main"
	if err := providerSchema.Validate(providerObject); err == nil {
		t.Fatal("provider schema accepted a malformed legacy ref")
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
		if strings.Contains(string(request), "review_mode") {
			t.Fatal("raw provider configuration was persisted")
		}
		var artifact map[string]any
		if err := json.Unmarshal(request, &artifact); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"schema_version", "id", "source", "ref", "configuration_hash"} {
			if _, ok := artifact[field]; !ok {
				t.Fatalf("pending artifact is missing %q", field)
			}
		}
		if len(artifact) != 5 {
			t.Fatalf("pending artifact contains unexpected fields: %v", artifact)
		}
	})

	t.Run("connected", func(t *testing.T) {
		root := t.TempDir()
		options := integrationInitOptions(root, configPath)
		options.IntegrationFetcher = staticIntegrationFetcher{bundle: bundle, available: true}
		report, err := Run(context.Background(), options, staticFetcher{snapshot: snapshot})
		if err != nil || report.Integrations[0].LifecycleState != "connected" {
			t.Fatalf("connected init = %#v, %v", report, err)
		}
		if !strings.Contains(string(readProjectPath(t, root, ".github/ai-collaboration.yml")), "advisory") {
			t.Fatal("connected provider output is missing")
		}
	})

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

func TestProviderLegacyRefDoesNotSelectConnectionState(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "ai.yml")
	writeTestFile(t, configPath, "review_mode: advisory\n")
	configurationHash, err := integrationConfigurationHash(configPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name string
		ref  string
	}{
		{name: "omitted"},
		{name: "matching", ref: providerRef},
		{name: "mismatching", ref: strings.Repeat("b", 40)},
	} {
		t.Run(test.name, func(t *testing.T) {
			bundle := validProviderBundle(configurationHash)
			bundle.Ref = test.ref
			plan, err := planIntegration(context.Background(), providerSource, providerRef, configPath, ProfileStandard, staticIntegrationFetcher{bundle: bundle, available: true})
			if err != nil || plan.Integration.LifecycleState != "connected" {
				t.Fatalf("legacy ref %q plan = %#v, %v", test.ref, plan, err)
			}
			if plan.Integration.Ref != providerRef {
				t.Fatalf("authoritative ref = %q", plan.Integration.Ref)
			}
		})
	}

	bundle := validProviderBundle(configurationHash)
	bundle.Ref = "main"
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeProviderBundle(data, bundle.Files); failureCode(err) != 5 {
		t.Fatalf("malformed legacy ref error = %v", err)
	}
	root := t.TempDir()
	options := integrationInitOptions(root, configPath)
	options.IntegrationFetcher = staticIntegrationFetcher{bundle: bundle, available: true}
	if _, err := Run(context.Background(), options, staticFetcher{snapshot: indexedSnapshot(t)}); failureCode(err) != 5 {
		t.Fatalf("malformed legacy ref lifecycle error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".soku")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("malformed legacy ref wrote lifecycle state")
	}
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
	return ProviderBundle{SchemaVersion: 1, ID: "ai-collaboration", Source: providerSource, Ref: providerRef, ProviderAPIVersion: "1", ProviderSchemaVersion: "1", CompatibleSoku: ">=0.1.0 <2.0.0", ConfigurationSchemaHash: "ca3d163bab055381827226140568f3bef7eaac187cebd76878e0b63e9e442356", ConfigurationHash: configurationHash, CompatibleProfiles: []string{ProfileBootstrap, ProfileScaled, ProfileStandard}, Outputs: []ProviderOutput{{Template: "templates/ai.yml", Path: ".github/ai-collaboration.yml", ContentMode: "text"}}, Files: map[string][]byte{"configuration.schema.json": []byte("{}\n"), "templates/ai.yml": []byte("mode: advisory\n")}}
}

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
