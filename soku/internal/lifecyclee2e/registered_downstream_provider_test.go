package lifecyclee2e

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
)

var registeredProviderIDs = []string{
	"cutvi-control-plane-v1",
	"archviz-control-plane-v1",
	"report-hub-control-plane-v1",
	"soku-pr-site-control-plane-v1",
}

func loadRegisteredBundle(t *testing.T, providerID string) initcmd.ProviderBundle {
	t.Helper()
	root := filepath.Join(repositoryRoot(t), "soku", "providers", providerID)
	metadata := readAbsolute(t, filepath.Join(root, "provider-v1.json"))
	bundle, err := initcmd.DecodeProviderBundle(metadata, map[string][]byte{
		"provider-v1.json": metadata,
		"configuration.schema.json": readAbsolute(
			t, filepath.Join(root, "configuration.schema.json"),
		),
		"example-config.yml": readAbsolute(
			t, filepath.Join(root, "example-config.yml"),
		),
		"templates/control-plane-provider.yml": readAbsolute(
			t, filepath.Join(root, "templates", "control-plane-provider.yml"),
		),
	})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestRegisteredDownstreamProvidersConnect(t *testing.T) {
	repository := repositoryRoot(t)
	for _, providerID := range registeredProviderIDs {
		t.Run(providerID, func(t *testing.T) {
			root := t.TempDir()
			options := initOptions(root)
			options.Explicit.Profile = initcmd.ProfileStandard
			options.IntegrationSource = "github:Soku-JINSEOK/" +
				"Soku-Convention-Boilerplate/soku/providers/" + providerID
			options.IntegrationRef = publicProviderRef
			options.IntegrationConfigPath = filepath.Join(
				repository, "soku", "providers", providerID, "example-config.yml",
			)
			options.IntegrationFetcher = providerFetcher{
				bundle: loadRegisteredBundle(t, providerID), available: true,
			}
			report, err := initcmd.Run(
				context.Background(),
				options,
				releaseFetcher{
					"v1.0.0": repositorySnapshot(t, "v1.0.0", baseCommit),
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Integrations) != 1 ||
				report.Integrations[0].LifecycleState != "connected" {
				t.Fatalf("integration report = %#v", report.Integrations)
			}
			assertClean(t, root)
		})
	}
}

func TestRegisteredProviderRejectsAnotherProjectConfiguration(t *testing.T) {
	repository := repositoryRoot(t)
	options := initOptions(t.TempDir())
	options.Explicit.Profile = initcmd.ProfileStandard
	options.IntegrationSource = "github:Soku-JINSEOK/" +
		"Soku-Convention-Boilerplate/soku/providers/cutvi-control-plane-v1"
	options.IntegrationRef = publicProviderRef
	options.IntegrationConfigPath = filepath.Join(
		repository,
		"soku/providers/archviz-control-plane-v1/example-config.yml",
	)
	options.IntegrationFetcher = providerFetcher{
		bundle: loadRegisteredBundle(t, "cutvi-control-plane-v1"), available: true,
	}
	report, err := initcmd.Run(
		context.Background(),
		options,
		releaseFetcher{"v1.0.0": repositorySnapshot(t, "v1.0.0", baseCommit)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Integrations) != 1 ||
		report.Integrations[0].LifecycleState == "connected" {
		t.Fatalf("cross-project configuration connected: %#v", report.Integrations)
	}
}
