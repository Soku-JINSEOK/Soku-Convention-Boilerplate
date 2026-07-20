package lifecyclee2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/initcmd"
	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const (
	networkProviderSource = "github:Soku-JINSEOK/Soku-Convention-Boilerplate/soku/providers/ai-collaboration"
	networkProviderCommit = "a81f7c91b0c9c8faa5ba2988fde29e9d17972a83"
)

type networkIntegrationFetcher struct {
	client *initcmd.SourceClient
	bundle initcmd.ProviderBundle
}

func (fetcher *networkIntegrationFetcher) FetchIntegration(ctx context.Context, source, ref string) (initcmd.ProviderBundle, bool, error) {
	bundle, available, err := fetcher.client.FetchIntegration(ctx, source, ref)
	if err == nil && available {
		fetcher.bundle = bundle
	}
	return bundle, available, err
}

func TestProviderNetworkConformance(t *testing.T) {
	if os.Getenv("SOKU_PROVIDER_NETWORK_CONFORMANCE") != "1" {
		t.Skip("set SOKU_PROVIDER_NETWORK_CONFORMANCE=1 to run the pinned HTTPS conformance test")
	}

	configPath := filepath.Join(t.TempDir(), "provider.yml")
	configuration := []byte("review_mode: advisory\nlanguages:\n  - en\n  - ko\n  - ja\n")
	if err := os.WriteFile(configPath, configuration, 0o600); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	options := initOptions(root)
	options.IntegrationSource = networkProviderSource
	options.IntegrationRef = networkProviderCommit
	options.IntegrationConfigPath = configPath
	fetcher := &networkIntegrationFetcher{client: initcmd.NewSourceClient()}
	options.IntegrationFetcher = fetcher

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	report, err := initcmd.Run(ctx, options, releaseFetcher{"v1.0.0": repositorySnapshot(t, "v1.0.0", baseCommit)})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Integrations) != 1 || report.Integrations[0].LifecycleState != "connected" {
		t.Fatalf("network provider report = %#v", report.Integrations)
	}
	if fetcher.bundle.Ref == nil || *fetcher.bundle.Ref == networkProviderCommit {
		t.Fatalf("fixture legacy ref = %#v, want a mismatching legacy value", fetcher.bundle.Ref)
	}
	if report.Integrations[0].Ref != networkProviderCommit {
		t.Fatalf("report ref = %q, want fetched commit", report.Integrations[0].Ref)
	}

	document, err := manifest.NewStore(root).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Integrations) != 1 || document.Integrations[0].Ref != networkProviderCommit {
		t.Fatalf("manifest integration = %#v", document.Integrations)
	}
	request := readFile(t, root, ".github/soku/integrations/ai-collaboration.json")
	var artifact map[string]any
	if err := json.Unmarshal(request, &artifact); err != nil {
		t.Fatal(err)
	}
	if len(artifact) != 5 || artifact["ref"] != networkProviderCommit {
		t.Fatalf("request artifact = %#v", artifact)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "ai-collaboration.yml")); err != nil {
		t.Fatalf("connected output: %v", err)
	}
	assertClean(t, root)
}
