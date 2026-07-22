package lifecyclee2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
)

const (
	publicProviderSource = "github:Soku-JINSEOK/Soku-Convention-Boilerplate/soku/providers/ci-cd-control-plane-v1"
	publicProviderRef    = "0123456789abcdef0123456789abcdef01234567"
)

func TestPublicControlPlaneProviderConformance(t *testing.T) {
	repository := repositoryRoot(t)
	providerRoot := filepath.Join(
		repository, "soku", "providers", "ci-cd-control-plane-v1",
	)
	metadata := readAbsolute(t, filepath.Join(providerRoot, "provider-v1.json"))
	files := map[string][]byte{
		"provider-v1.json": metadata,
		"configuration.schema.json": readAbsolute(
			t, filepath.Join(providerRoot, "configuration.schema.json"),
		),
		"example-config.yml": readAbsolute(
			t, filepath.Join(providerRoot, "example-config.yml"),
		),
		"templates/control-plane-provider.yml": readAbsolute(
			t,
			filepath.Join(
				providerRoot, "templates", "control-plane-provider.yml",
			),
		),
	}
	bundle, err := initcmd.DecodeProviderBundle(metadata, files)
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	options := initOptions(root)
	options.Explicit.Profile = initcmd.ProfileStandard
	options.IntegrationSource = publicProviderSource
	options.IntegrationRef = publicProviderRef
	options.IntegrationConfigPath = filepath.Join(
		providerRoot, "example-config.yml",
	)
	options.IntegrationFetcher = providerFetcher{
		bundle: bundle, available: true,
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
		report.Integrations[0].LifecycleState != "connected" ||
		report.Integrations[0].Ref != publicProviderRef {
		t.Fatalf("public provider report = %#v", report.Integrations)
	}
	output := filepath.Join(
		root, ".github", "control-plane-provider.yml",
	)
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("public provider output: %v", err)
	}
	if string(readAbsolute(t, output)) !=
		string(files["templates/control-plane-provider.yml"]) {
		t.Fatal("public provider output is not a literal-byte copy")
	}
	assertClean(t, root)
}
