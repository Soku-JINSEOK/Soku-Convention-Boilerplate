package lifecyclee2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const (
	providerRequestSource = "github:example/provider/ai-collaboration"
	providerCommit        = "abcdef0123456789abcdef0123456789abcdef01"
)

type providerFetcher struct {
	bundle    initcmd.ProviderBundle
	available bool
}

func (fetcher providerFetcher) FetchIntegration(context.Context, string, string) (initcmd.ProviderBundle, bool, error) {
	return fetcher.bundle, fetcher.available, nil
}

func TestProviderConformanceProfileMatrix(t *testing.T) {
	snapshot := repositorySnapshot(t, "v1.0.0", baseCommit)
	configPath, configurationHash := providerConfiguration(t)
	bundle := conformanceProvider(configurationHash)

	for _, profile := range []string{initcmd.ProfileBootstrap, initcmd.ProfileStandard, initcmd.ProfileScaled} {
		profile := profile
		t.Run(profile, func(t *testing.T) {
			root := t.TempDir()
			options := initOptions(root)
			options.Explicit.Profile = profile
			options.IntegrationSource = providerRequestSource
			options.IntegrationRef = providerCommit
			options.IntegrationConfigPath = configPath
			options.IntegrationFetcher = providerFetcher{bundle: bundle, available: true}

			report, err := initcmd.Run(context.Background(), options, releaseFetcher{"v1.0.0": snapshot})
			if err != nil || len(report.Integrations) != 1 || report.Integrations[0].LifecycleState != "connected" {
				t.Fatalf("profile %s provider report = %#v, %v", profile, report, err)
			}
			if _, err := os.Stat(filepath.Join(root, ".github", "ai-collaboration.yml")); err != nil {
				t.Fatalf("profile %s provider output: %v", profile, err)
			}
			assertClean(t, root)
		})
	}
}

func TestProviderConformanceLifecycleTransitions(t *testing.T) {
	base := repositorySnapshot(t, "v1.0.0", baseCommit)
	target := cloneSnapshot(base)
	target.Release = "v1.1.0"
	target.ResolvedCommit = targetCommit
	configPath, configurationHash := providerConfiguration(t)
	bundle := conformanceProvider(configurationHash)
	fetcher := releaseFetcher{"v1.0.0": base, "v1.1.0": target}

	t.Run("pending-to-connected", func(t *testing.T) {
		root := t.TempDir()
		options := integrationOptions(root, configPath, providerFetcher{available: false})
		report, err := initcmd.Run(context.Background(), options, fetcher)
		if err != nil || report.Integrations[0].LifecycleState != "pending" {
			t.Fatalf("pending init = %#v, %v", report, err)
		}

		transition, err := initcmd.RunTransition(context.Background(), initcmd.TransitionOptions{
			Root: root, TargetRelease: "v1.0.0", Yes: true, SokuVersion: "e2e",
			IntegrationSource: providerRequestSource, IntegrationRef: providerCommit,
			IntegrationConfigPath: configPath,
			IntegrationFetcher:    providerFetcher{bundle: bundle, available: true},
		}, fetcher, true)
		if err != nil || transition.State != "applied" || transition.Integrations[0].LifecycleState != "connected" {
			t.Fatalf("connect transition = %#v, %v", transition, err)
		}
		assertClean(t, root)
	})

	t.Run("release-profile-provider-upgrade", func(t *testing.T) {
		root := t.TempDir()
		initialize(t, root, base, []string{"go"})
		transition, err := initcmd.RunTransition(context.Background(), initcmd.TransitionOptions{
			Root: root, TargetRelease: "v1.1.0", TargetProfile: initcmd.ProfileScaled,
			Yes: true, SokuVersion: "e2e", IntegrationSource: providerRequestSource,
			IntegrationRef: providerCommit, IntegrationConfigPath: configPath,
			IntegrationFetcher: providerFetcher{bundle: bundle, available: true},
		}, fetcher, true)
		if err != nil || transition.State != "applied" || transition.TargetProfile != initcmd.ProfileScaled || transition.Integrations[0].LifecycleState != "connected" {
			t.Fatalf("combined upgrade = %#v, %v", transition, err)
		}
		assertClean(t, root)
	})
}

func TestProviderConformanceRejectsIncompatibleState(t *testing.T) {
	snapshot := repositorySnapshot(t, "v1.0.0", baseCommit)
	configPath, configurationHash := providerConfiguration(t)
	bundle := conformanceProvider(configurationHash)

	t.Run("unsupported-profile", func(t *testing.T) {
		bundle := bundle
		bundle.CompatibleProfiles = []string{initcmd.ProfileBootstrap}
		root := t.TempDir()
		options := integrationOptions(root, configPath, providerFetcher{bundle: bundle, available: true})
		if _, err := initcmd.Run(context.Background(), options, releaseFetcher{"v1.0.0": snapshot}); failureCode(err) != 5 {
			t.Fatalf("unsupported profile error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, ".soku")); !os.IsNotExist(err) {
			t.Fatal("unsupported provider wrote lifecycle state")
		}
		if len(projectTree(t, root)) != 0 {
			t.Fatal("unsupported provider wrote project output")
		}
	})

	t.Run("ownership-conflict", func(t *testing.T) {
		bundle := bundle
		bundle.Outputs[0].Path = "profile.go"
		root := t.TempDir()
		options := integrationOptions(root, configPath, providerFetcher{bundle: bundle, available: true})
		if _, err := initcmd.Run(context.Background(), options, releaseFetcher{"v1.0.0": snapshot}); failureCode(err) != 4 {
			t.Fatalf("ownership error = %v", err)
		}
		if len(projectTree(t, root)) != 0 {
			t.Fatal("ownership conflict wrote lifecycle state or project output")
		}
	})

	t.Run("unsupported-manifest", func(t *testing.T) {
		root := t.TempDir()
		initialize(t, root, snapshot, []string{"go"})
		manifestPath := filepath.Join(root, filepath.FromSlash(manifest.ManifestPath))
		var document map[string]any
		if err := json.Unmarshal(readFile(t, root, manifest.ManifestPath), &document); err != nil {
			t.Fatal(err)
		}
		document["schema_version"] = 999
		data, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(manifestPath, append(data, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err = initcmd.RunTransition(context.Background(), initcmd.TransitionOptions{Root: root, TargetRelease: "v1.0.0"}, releaseFetcher{"v1.0.0": snapshot}, false)
		if failureCode(err) != 5 {
			t.Fatalf("manifest compatibility error = %v", err)
		}
	})
}

func integrationOptions(root, configPath string, fetcher providerFetcher) initcmd.Options {
	options := initOptions(root)
	options.IntegrationSource = providerRequestSource
	options.IntegrationRef = providerCommit
	options.IntegrationConfigPath = configPath
	options.IntegrationFetcher = fetcher
	return options
}

func providerConfiguration(t *testing.T) (string, string) {
	t.Helper()
	content := []byte("review_mode: advisory\n")
	path := filepath.Join(t.TempDir(), "provider.yml")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	return path, hex.EncodeToString(sum[:])
}

func conformanceProvider(configurationHash string) initcmd.ProviderBundle {
	return initcmd.ProviderBundle{
		SchemaVersion: 1, ID: "ai-collaboration", Source: providerRequestSource,
		ProviderAPIVersion: "1", ProviderSchemaVersion: "1",
		CompatibleSoku:          ">=0.1.0 <2.0.0",
		ConfigurationSchemaHash: "ca3d163bab055381827226140568f3bef7eaac187cebd76878e0b63e9e442356",
		ConfigurationHash:       configurationHash,
		CompatibleProfiles:      []string{initcmd.ProfileBootstrap, initcmd.ProfileScaled, initcmd.ProfileStandard},
		Outputs: []initcmd.ProviderOutput{{
			Template: "templates/ai.yml", Path: ".github/ai-collaboration.yml", ContentMode: "text",
		}},
		Files: map[string][]byte{
			"configuration.schema.json": []byte("{}\n"),
			"templates/ai.yml":          []byte("mode: advisory\n"),
		},
	}
}
