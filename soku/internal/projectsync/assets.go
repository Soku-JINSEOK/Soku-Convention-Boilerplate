// Package projectsync implements the opt-in GitHub Project Sync core component.
package projectsync

import (
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	ComponentID    = "github-project-sync"
	CatalogVersion = "1"
	ConfigPath     = ".github/project-sync.yml"
	projectNumber  = "__PROJECT_NUMBER__"
)

//go:embed assets/project-sync.yml assets/project-sync-workflow.yml assets/github-project-sync.mjs assets/github-project-sync.test.mjs
var componentAssets embed.FS

// Asset describes one output owned by the component.
type Asset struct {
	Path        string
	Owner       string
	Class       string
	ContentMode string
	Content     []byte
}

// Assets returns the reviewed component outputs for a positive user-owned
// GitHub Project number. The configuration is the only output that contains
// the selected number; the runtime resolves the repository from the workflow
// GITHUB_REPOSITORY environment variable.
func Assets(number int) ([]Asset, error) {
	if number < 1 {
		return nil, fmt.Errorf("project number must be a positive integer")
	}
	config, err := componentAssets.ReadFile("assets/project-sync.yml")
	if err != nil {
		return nil, fmt.Errorf("read project sync configuration asset: %w", err)
	}
	config = []byte(strings.ReplaceAll(string(config), projectNumber, strconv.Itoa(number)))
	files := map[string]Asset{
		ConfigPath: {
			Path: ConfigPath, Owner: "project", Class: "project-owned", Content: config,
		},
		".github/workflows/project-sync.yml":   readAsset(".github/workflows/project-sync.yml", "assets/project-sync-workflow.yml", "core", "core-managed", "text"),
		"scripts/github-project-sync.mjs":      readAsset("scripts/github-project-sync.mjs", "assets/github-project-sync.mjs", "core", "core-managed", "text"),
		"scripts/github-project-sync.test.mjs": readAsset("scripts/github-project-sync.test.mjs", "assets/github-project-sync.test.mjs", "core", "core-managed", "text"),
	}
	result := make([]Asset, 0, len(files))
	for _, asset := range files {
		if asset.Content == nil {
			return nil, fmt.Errorf("component asset is missing")
		}
		result = append(result, asset)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

// ManagedAssets returns the component-managed outputs used by lifecycle
// diff/upgrade. Project-owned configuration is intentionally excluded.
func ManagedAssets() ([]Asset, error) {
	assets, err := Assets(1)
	if err != nil {
		return nil, err
	}
	result := assets[:0]
	for _, asset := range assets {
		if asset.Class != "project-owned" {
			result = append(result, asset)
		}
	}
	return result, nil
}

func readAsset(output, source, owner, class, mode string) Asset {
	content, err := componentAssets.ReadFile(source)
	if err != nil {
		return Asset{Path: output, Owner: owner, Class: class, ContentMode: mode}
	}
	return Asset{Path: output, Owner: owner, Class: class, ContentMode: mode, Content: content}
}
