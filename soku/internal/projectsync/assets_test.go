package projectsync

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetsArePortableAndProjectSpecific(t *testing.T) {
	assets, err := Assets(17)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 4 {
		t.Fatalf("asset count = %d, want 4", len(assets))
	}
	for _, asset := range assets {
		if asset.Content == nil {
			t.Fatalf("asset %q is empty", asset.Path)
		}
		if asset.Path == ConfigPath {
			if asset.Class != "project-owned" || asset.Owner != "project" || asset.ContentMode != "" {
				t.Fatalf("config ownership = %#v", asset)
			}
			if !bytes.Contains(asset.Content, []byte(`"number": 17`)) {
				t.Fatalf("config does not contain selected Project number: %s", asset.Content)
			}
			if bytes.Contains(asset.Content, []byte("repository")) || bytes.Contains(asset.Content, []byte("dependency-tracking")) {
				t.Fatalf("config contains repository-specific metadata: %s", asset.Content)
			}
		} else if asset.Owner != "core" || asset.Class != "core-managed" || asset.ContentMode != "text" {
			t.Fatalf("managed asset ownership = %#v", asset)
		}
		if bytes.Contains(asset.Content, []byte("August 2026")) || bytes.Contains(asset.Content, []byte("Soku-JINSEOK/Soku-Convention-Boilerplate")) {
			t.Fatalf("asset contains boilerplate-specific metadata: %s", asset.Content)
		}
	}
}

func TestAssetsRejectNonPositiveProjectNumber(t *testing.T) {
	if _, err := Assets(0); err == nil {
		t.Fatal("zero Project number was accepted")
	}
}

func TestCatalogAndWorkflowDeclareTheReviewedComponentContract(t *testing.T) {
	catalogData, err := os.ReadFile(filepath.Join("..", "..", "components", "github-project-sync", "component-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		ID                 string   `json:"id"`
		CatalogVersion     string   `json:"catalog_version"`
		ManifestSchema     int      `json:"manifest_schema"`
		CoreManagedOutputs []string `json:"core_managed_outputs"`
		ProjectOwned       []string `json:"project_owned_outputs"`
	}
	if err := json.Unmarshal(catalogData, &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.ID != ComponentID || catalog.CatalogVersion != CatalogVersion || catalog.ManifestSchema != 2 {
		t.Fatalf("catalog identity = %#v", catalog)
	}
	if len(catalog.CoreManagedOutputs) != 3 || len(catalog.ProjectOwned) != 1 || catalog.ProjectOwned[0] != ConfigPath {
		t.Fatalf("catalog outputs = %#v", catalog)
	}

	assets, err := Assets(17)
	if err != nil {
		t.Fatal(err)
	}
	var workflow string
	for _, asset := range assets {
		if asset.Path == ".github/workflows/project-sync.yml" {
			workflow = string(asset.Content)
		}
	}
	for _, required := range []string{
		"vars.PROJECT_SYNC_ENABLED == 'true'",
		"PROJECT_SYNC_MODE:-audit",
		"--repo \"$GITHUB_REPOSITORY\"",
		"--config .github/project-sync.yml",
		"secrets.PROJECT_SYNC_TOKEN",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("workflow is missing %q", required)
		}
	}
	if strings.Contains(workflow, "--project-number 2") || strings.Contains(workflow, "Soku-JINSEOK/Convention") {
		t.Fatal("workflow contains repository-specific Project metadata")
	}
}
