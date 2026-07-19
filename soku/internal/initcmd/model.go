// Package initcmd implements the bounded, transactional soku init workflow.
package initcmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const (
	CatalogPath      = "soku/catalog/core-v1.json"
	ProfileIndexPath = "soku/catalog/index-v2.json"
	ProfileBootstrap = "bootstrap"
	ProfileStandard  = "standard"
	ProfileScaled    = "scaled"
)

var StackIDs = []string{"aws", "azure", "gcp", "go", "java-spring", "javascript-typescript-node", "mysql", "postgresql", "python"}

type Failure struct {
	Code    int
	Key     string
	Message string
	Cause   error
	Data    any
}

func (e *Failure) Error() string { return e.Message }
func (e *Failure) Unwrap() error { return e.Cause }

func fail(code int, key, format string, args ...any) *Failure {
	return &Failure{Code: code, Key: key, Message: fmt.Sprintf(format, args...)}
}

type Config struct {
	SchemaVersion      int      `json:"schema_version"`
	BoilerplateSource  string   `json:"boilerplate_source"`
	BoilerplateRelease string   `json:"boilerplate_release"`
	Stacks             []string `json:"stacks"`
	Profile            string   `json:"profile"`
	ProjectName        string   `json:"project_name,omitempty"`
	ModulePath         string   `json:"module_path,omitempty"`
	JavaGroup          string   `json:"java_group,omitempty"`
	ServiceName        string   `json:"service_name,omitempty"`
	Verify             bool     `json:"verify"`
}

type Explicit struct {
	Source, Release, Profile, ProjectName, ModulePath, JavaGroup, ServiceName                                            string
	Stacks                                                                                                               []string
	Verify                                                                                                               bool
	SourceSet, ReleaseSet, ProfileSet, StacksSet, ProjectNameSet, ModulePathSet, JavaGroupSet, ServiceNameSet, VerifySet bool
}

type Catalog struct {
	SchemaVersion int           `json:"schema_version"`
	Profile       string        `json:"profile"`
	Files         []CatalogFile `json:"files"`
	Stacks        []Stack       `json:"stacks"`
}

type Stack struct {
	ID      string        `json:"id"`
	Markers []string      `json:"markers"`
	Files   []CatalogFile `json:"files"`
}

type CatalogFile struct {
	Source       string   `json:"source"`
	Output       string   `json:"output"`
	Owner        string   `json:"owner"`
	Class        string   `json:"class"`
	ContentMode  string   `json:"content_mode"`
	Strategy     string   `json:"strategy"`
	Placeholders []string `json:"placeholders"`
}

type ProfileIndex struct {
	SchemaVersion  int            `json:"schema_version"`
	DefaultProfile string         `json:"default_profile"`
	Profiles       []Profile      `json:"profiles"`
	Layers         []ProfileLayer `json:"layers"`
}

type Profile struct {
	ID     string   `json:"id"`
	Layers []string `json:"layers"`
}

type ProfileLayer struct {
	ID             string        `json:"id"`
	SharedOutputs  []string      `json:"shared_outputs"`
	StackFileLimit int           `json:"stack_file_limit"`
	Files          []CatalogFile `json:"files"`
}

type SourceSnapshot struct {
	Source         string            `json:"source"`
	Release        string            `json:"release"`
	ResolvedCommit string            `json:"resolved_commit"`
	Files          map[string][]byte `json:"-"`
}

type Change struct {
	Path           string `json:"path"`
	Action         string `json:"action"`
	Owner          string `json:"owner"`
	Class          string `json:"class"`
	ContentMode    string `json:"content_mode"`
	BaselineSHA256 string `json:"baseline_sha256"`
	Content        []byte `json:"-"`
}

type Verification struct {
	Stack   string   `json:"stack"`
	Command []string `json:"command"`
	Status  string   `json:"status"`
}

type Recovery struct {
	Required      bool     `json:"required"`
	TransactionID string   `json:"transaction_id,omitempty"`
	Instructions  []string `json:"instructions"`
}

type Report struct {
	State             string                 `json:"state"`
	Source            string                 `json:"source"`
	Release           string                 `json:"release"`
	ResolvedCommit    string                 `json:"resolved_commit"`
	Profile           string                 `json:"profile"`
	Stacks            []string               `json:"stacks"`
	SelectionHash     string                 `json:"selection_hash"`
	ConfigurationHash string                 `json:"configuration_hash"`
	Changes           []Change               `json:"changes"`
	Verification      []Verification         `json:"verification"`
	Recovery          Recovery               `json:"recovery"`
	Integrations      []manifest.Integration `json:"integrations"`
}

type Options struct {
	Root                  string
	ConfigPath            string
	Explicit              Explicit
	DryRun                bool
	Yes                   bool
	Interactive           bool
	Confirm               func(Report) (bool, error)
	SokuVersion           string
	IntegrationSource     string
	IntegrationRef        string
	IntegrationConfigPath string
	IntegrationFetcher    IntegrationFetcher
}

func canonicalHash(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeStacks(stacks []string) ([]string, error) {
	known := make(map[string]bool, len(StackIDs))
	for _, id := range StackIDs {
		known[id] = true
	}
	seen := map[string]bool{}
	for _, id := range stacks {
		id = strings.TrimSpace(id)
		if !known[id] {
			return nil, fail(2, "selection.invalid", "unsupported stack %q", id)
		}
		if seen[id] {
			return nil, fail(2, "selection.invalid", "stack %q is repeated", id)
		}
		seen[id] = true
	}
	result := append([]string(nil), stacks...)
	sort.Strings(result)
	return result, nil
}
