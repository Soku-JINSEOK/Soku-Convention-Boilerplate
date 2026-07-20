package initcmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

var (
	integrationSourcePattern = regexp.MustCompile(`^github:([A-Za-z0-9](?:[A-Za-z0-9-]{0,38}))/([A-Za-z0-9_.-]+)/([A-Za-z0-9._/-]+)$`)
	integrationIDPattern     = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	providerExecutable       = regexp.MustCompile(`(?i)\.(?:sh|bash|zsh|exe|dll|so|dylib|bat|cmd|ps1|jar)$`)
)

type IntegrationFetcher interface {
	FetchIntegration(context.Context, string, string) (ProviderBundle, bool, error)
}

func (client *SourceClient) FetchIntegration(ctx context.Context, source, ref string) (ProviderBundle, bool, error) {
	match := integrationSourcePattern.FindStringSubmatch(source)
	if match == nil || !lowerCommit(ref) {
		return ProviderBundle{}, false, fail(2, "integration.source.invalid", "provider source or ref is invalid")
	}
	if client.HTTP == nil {
		client.HTTP = NewSourceClient().HTTP
	}
	if client.APIBase == "" {
		client.APIBase = "https://api.github.com"
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", strings.TrimSuffix(client.APIBase, "/"), match[1], match[2], ref)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	client.authorize(request)
	response, err := client.HTTP.Do(request)
	if err != nil {
		return ProviderBundle{}, false, fail(6, "provider.fetch", "download provider archive: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return ProviderBundle{}, false, fail(6, "provider.fetch", "download provider archive: GitHub returned %s", response.Status)
	}
	if response.Request.URL.Scheme != "https" && !isLoopbackTestURL(response.Request.URL, client.APIBase) {
		return ProviderBundle{}, false, fail(6, "provider.fetch", "provider archive redirect target must use HTTPS")
	}
	archive, err := io.ReadAll(io.LimitReader(response.Body, maxArchiveBytes+1))
	if err != nil || len(archive) > maxArchiveBytes {
		return ProviderBundle{}, false, fail(6, "provider.fetch", "provider archive exceeds the bounded compressed size")
	}
	archiveFiles, err := extractArchive(archive)
	if err != nil {
		return ProviderBundle{}, false, err
	}
	bundlePath := strings.Trim(match[3], "/")
	bundleFiles := map[string][]byte{}
	for path, content := range archiveFiles {
		if strings.HasPrefix(path, bundlePath+"/") {
			bundleFiles[strings.TrimPrefix(path, bundlePath+"/")] = content
		}
	}
	metadata, ok := bundleFiles["provider-v1.json"]
	if !ok {
		return ProviderBundle{}, false, nil
	}
	bundle, err := DecodeProviderBundle(metadata, bundleFiles)
	if err != nil {
		return ProviderBundle{}, false, err
	}
	return bundle, true, nil
}

type ProviderBundle struct {
	SchemaVersion           int               `json:"schema_version"`
	ID                      string            `json:"id"`
	Source                  string            `json:"source"`
	Ref                     string            `json:"ref,omitempty"`
	ProviderAPIVersion      string            `json:"provider_api_version"`
	ProviderSchemaVersion   string            `json:"provider_schema_version"`
	CompatibleSoku          string            `json:"compatible_soku"`
	ConfigurationSchemaHash string            `json:"configuration_schema_hash"`
	ConfigurationHash       string            `json:"configuration_hash"`
	CompatibleProfiles      []string          `json:"compatible_profiles"`
	Outputs                 []ProviderOutput  `json:"outputs"`
	Files                   map[string][]byte `json:"-"`
}

type ProviderOutput struct {
	Template    string `json:"template"`
	Path        string `json:"path"`
	ContentMode string `json:"content_mode"`
}

type integrationPlan struct {
	Changes     []Change
	Integration manifest.Integration
}

func DecodeProviderBundle(data []byte, files map[string][]byte) (ProviderBundle, error) {
	var bundle ProviderBundle
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bundle); err != nil {
		return ProviderBundle{}, fail(5, "provider.incompatible", "decode provider bundle: %v", err)
	}
	bundle.Files = files
	if err := validateProviderBundle(bundle); err != nil {
		return ProviderBundle{}, err
	}
	return bundle, nil
}

func validateProviderBundle(bundle ProviderBundle) error {
	if bundle.SchemaVersion != 1 || bundle.ProviderAPIVersion != "1" || bundle.ProviderSchemaVersion != "1" || !integrationIDPattern.MatchString(bundle.ID) || !regexp.MustCompile(`^>=[0-9]+\.[0-9]+\.[0-9]+ <[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(bundle.CompatibleSoku) {
		return fail(5, "provider.incompatible", "provider bundle must use API/schema v1 and a portable id")
	}
	_, sourceID, sourceErr := parseIntegrationSource(bundle.Source)
	if sourceErr != nil || sourceID != bundle.ID {
		return fail(5, "provider.incompatible", "provider source and id are inconsistent")
	}
	if (bundle.Ref != "" && !lowerCommit(bundle.Ref)) || !isSHA256(bundle.ConfigurationSchemaHash) || !isSHA256(bundle.ConfigurationHash) || len(bundle.Outputs) == 0 {
		return fail(5, "provider.incompatible", "provider bundle has invalid immutable or schema metadata")
	}
	schema, schemaExists := bundle.Files["configuration.schema.json"]
	schemaHash := sha256.Sum256([]byte(normalizeText(schema)))
	if !schemaExists || hex.EncodeToString(schemaHash[:]) != bundle.ConfigurationSchemaHash {
		return fail(5, "provider.incompatible", "provider configuration schema is missing or does not match its hash")
	}
	profiles := append([]string(nil), bundle.CompatibleProfiles...)
	sort.Strings(profiles)
	if strings.Join(profiles, "\x00") != strings.Join(bundle.CompatibleProfiles, "\x00") {
		return fail(5, "provider.incompatible", "provider compatible profiles must be sorted")
	}
	seen := map[string]bool{}
	allowedFiles := map[string]bool{"provider-v1.json": true, "configuration.schema.json": true, "example-config.yml": true}
	for _, output := range bundle.Outputs {
		if output.ContentMode != "text" && output.ContentMode != "binary" {
			return fail(5, "provider.incompatible", "provider output %q has invalid content mode", output.Path)
		}
		if err := manifest.ValidatePath(output.Template); err != nil || providerExecutable.MatchString(output.Template) {
			return fail(5, "provider.incompatible", "provider template %q is unsafe or executable", output.Template)
		}
		if err := manifest.ValidatePath(output.Path); err != nil || providerExecutable.MatchString(output.Path) {
			return fail(5, "provider.incompatible", "provider output %q is unsafe or executable", output.Path)
		}
		folded := strings.ToLower(output.Path)
		if seen[folded] {
			return fail(4, "ownership.conflict", "provider output %q is repeated", output.Path)
		}
		seen[folded] = true
		if _, ok := bundle.Files[output.Template]; !ok {
			return fail(5, "provider.incompatible", "provider template %q is missing", output.Template)
		}
		allowedFiles[output.Template] = true
	}
	for path := range bundle.Files {
		if !allowedFiles[path] || providerExecutable.MatchString(path) {
			return fail(5, "provider.incompatible", "provider bundle contains undeclared or executable file %q", path)
		}
	}
	return nil
}

func planIntegration(ctx context.Context, source, ref, configPath, profile string, fetcher IntegrationFetcher) (integrationPlan, error) {
	portableSource, id, err := parseIntegrationSource(source)
	if err != nil {
		return integrationPlan{}, err
	}
	if !lowerCommit(ref) {
		return integrationPlan{}, fail(2, "integration.ref.invalid", "--integration-ref must be a lowercase 40-character SHA")
	}
	configurationHash, err := integrationConfigurationHash(configPath)
	if err != nil {
		return integrationPlan{}, err
	}
	requestPath := ".github/soku/integrations/" + id + ".json"
	requestContent, _ := json.MarshalIndent(struct {
		SchemaVersion     int    `json:"schema_version"`
		ID                string `json:"id"`
		Source            string `json:"source"`
		Ref               string `json:"ref"`
		ConfigurationHash string `json:"configuration_hash"`
	}{1, id, portableSource, ref, configurationHash}, "", "  ")
	requestContent = append(requestContent, '\n')
	requestHash, _ := manifest.HashContent(requestContent, "text")
	plan := integrationPlan{
		Changes:     []Change{{Path: requestPath, Action: "create", Owner: id, Class: "provider-managed", ContentMode: "text", BaselineSHA256: requestHash, Content: requestContent}},
		Integration: manifest.Integration{ID: id, Source: portableSource, Ref: ref, ProviderAPIVersion: "1", ProviderSchemaVersion: "1", ConfigurationHash: configurationHash, LifecycleState: "pending", ManagedFiles: []string{requestPath}},
	}
	if fetcher == nil {
		return plan, nil
	}
	bundle, available, err := fetcher.FetchIntegration(ctx, source, ref)
	if err != nil {
		return integrationPlan{}, err
	}
	if !available {
		return plan, nil
	}
	if err := validateProviderBundle(bundle); err != nil {
		return integrationPlan{}, err
	}
	if bundle.ID != id || bundle.Source != source || bundle.ConfigurationHash != configurationHash {
		return plan, nil
	}
	if !contains(bundle.CompatibleProfiles, profile) {
		return integrationPlan{}, fail(5, "provider.incompatible", "provider %q does not support profile %q", id, profile)
	}
	for _, output := range bundle.Outputs {
		content := append([]byte(nil), bundle.Files[output.Template]...)
		hash, hashErr := manifest.HashContent(content, output.ContentMode)
		if hashErr != nil {
			return integrationPlan{}, fail(5, "provider.incompatible", "provider output %q: %v", output.Path, hashErr)
		}
		plan.Changes = append(plan.Changes, Change{Path: output.Path, Action: "create", Owner: id, Class: "provider-managed", ContentMode: output.ContentMode, BaselineSHA256: hash, Content: content})
		plan.Integration.ManagedFiles = append(plan.Integration.ManagedFiles, output.Path)
	}
	sort.Strings(plan.Integration.ManagedFiles)
	plan.Integration.LifecycleState = "connected"
	return plan, nil
}

func parseIntegrationSource(source string) (string, string, error) {
	match := integrationSourcePattern.FindStringSubmatch(source)
	if match == nil || strings.Contains(match[3], "..") || strings.HasPrefix(match[3], "/") {
		return "", "", fail(2, "integration.source.invalid", "--integration-source must use github:<owner>/<repo>/<bundle-path>")
	}
	parts := strings.Split(strings.TrimSuffix(match[3], "/"), "/")
	id := strings.ToLower(parts[len(parts)-1])
	id = strings.TrimSuffix(id, filepath.Ext(id))
	id = strings.ReplaceAll(id, "_", "-")
	if !integrationIDPattern.MatchString(id) {
		return "", "", fail(2, "integration.source.invalid", "bundle path must end in a portable provider id")
	}
	return fmt.Sprintf("https://github.com/%s/%s/%s", match[1], match[2], match[3]), id, nil
}

func integrationConfigurationHash(path string) (string, error) {
	if path == "" {
		return "", fail(2, "integration.configuration.invalid", "--integration-config is required")
	}
	extension := strings.ToLower(filepath.Ext(path))
	if extension != ".yml" && extension != ".yaml" {
		return "", fail(2, "integration.configuration.invalid", "integration configuration must use a .yml or .yaml path")
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64*1024 {
		return "", fail(2, "integration.configuration.invalid", "integration configuration must be a regular file no larger than 64 KiB")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fail(2, "integration.configuration.invalid", "read integration configuration: %v", err)
	}
	if secretBearing(content) || regexp.MustCompile(`(?im)^\s*(?:token|password|secret|credential|api[_-]?key)\s*:`).Match(content) {
		return "", fail(2, "integration.configuration.secret", "integration configuration appears to contain a secret")
	}
	if len(content) == 0 || !utf8.Valid(content) {
		return "", fail(2, "integration.configuration.invalid", "integration configuration must be non-empty UTF-8 YAML")
	}
	normalized := []byte(normalizeText(content))
	sum := sha256.Sum256(normalized)
	return hex.EncodeToString(sum[:]), nil
}

func providerBaselineChanges(root string, document manifest.Document) ([]Change, error) {
	result := []Change{}
	for _, file := range document.Files {
		if file.Owner == "core" || file.Owner == "project" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fail(4, "path.conflict", "read provider path %q: %v", file.Path, err)
		}
		result = append(result, Change{Path: file.Path, Action: "unchanged", Owner: file.Owner, Class: file.Class, ContentMode: file.ContentMode, BaselineSHA256: file.BaselineSHA256, Content: content})
	}
	return result, nil
}

func validateChangeOwnership(changes []Change) error {
	seen := map[string]Change{}
	for _, change := range changes {
		folded := strings.ToLower(change.Path)
		if previous, exists := seen[folded]; exists {
			return fail(4, "ownership.conflict", "path %q is declared by both %s and %s", change.Path, previous.Owner, change.Owner)
		}
		seen[folded] = change
	}
	return nil
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}
