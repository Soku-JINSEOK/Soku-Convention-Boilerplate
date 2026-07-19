// Package manifest defines and validates the portable Soku manifest v1 wire format.
package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	// SchemaVersion is the only manifest schema understood by this release.
	SchemaVersion = 1
	// ManifestPath is the repository-relative location of durable lifecycle state.
	ManifestPath = ".soku/manifest.json"
	// PendingPath is the repository-relative interrupted-write marker.
	PendingPath = ".soku/manifest.json.pending"
)

var (
	integrationIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	portableDrivePattern = regexp.MustCompile(`^[A-Za-z]:`)
)

// Document is the complete manifest-v1 wire record.
type Document struct {
	SchemaVersion int           `json:"schema_version"`
	SokuVersion   string        `json:"soku_version"`
	Boilerplate   Boilerplate   `json:"boilerplate"`
	Selection     Selection     `json:"selection"`
	Files         []File        `json:"files"`
	Integrations  []Integration `json:"integrations"`
}

// Boilerplate pins the immutable convention source last applied.
type Boilerplate struct {
	Source         string `json:"source"`
	Release        string `json:"release"`
	ResolvedCommit string `json:"resolved_commit"`
}

// Selection records portable inputs without retaining raw configuration.
type Selection struct {
	Profile           string   `json:"profile"`
	Stacks            []string `json:"stacks"`
	ProjectName       string   `json:"project_name,omitempty"`
	ModulePath        string   `json:"module_path,omitempty"`
	JavaGroup         string   `json:"java_group,omitempty"`
	ServiceName       string   `json:"service_name,omitempty"`
	ConfigurationHash string   `json:"configuration_hash"`
}

// File records ownership and the last applied baseline for one path.
type File struct {
	Path           string `json:"path"`
	Owner          string `json:"owner"`
	Class          string `json:"class"`
	ContentMode    string `json:"content_mode,omitempty"`
	BaselineSHA256 string `json:"baseline_sha256,omitempty"`
	LifecycleState string `json:"lifecycle_state"`
}

// Integration records one exact provider snapshot and its managed paths.
type Integration struct {
	ID                    string   `json:"id"`
	Source                string   `json:"source"`
	Ref                   string   `json:"ref"`
	ProviderAPIVersion    string   `json:"provider_api_version"`
	ProviderSchemaVersion string   `json:"provider_schema_version"`
	ConfigurationHash     string   `json:"configuration_hash"`
	LifecycleState        string   `json:"lifecycle_state"`
	ManagedFiles          []string `json:"managed_files"`
}

// UnsupportedSchemaError distinguishes a valid JSON record from a version this
// binary cannot interpret.
type UnsupportedSchemaError struct {
	Version int
}

func (e *UnsupportedSchemaError) Error() string {
	return fmt.Sprintf("manifest schema version %d is not supported", e.Version)
}

// Decode parses exactly one manifest and rejects undeclared fields.
func Decode(data []byte) (Document, error) {
	var header struct {
		SchemaVersion *int `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return Document{}, fmt.Errorf("decode manifest: %w", err)
	}
	if header.SchemaVersion == nil || *header.SchemaVersion <= 0 {
		return Document{}, errors.New("schema_version must be a positive integer")
	}
	if *header.SchemaVersion != SchemaVersion {
		return Document{}, &UnsupportedSchemaError{Version: *header.SchemaVersion}
	}

	var document Document
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return Document{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Document{}, errors.New("decode manifest: multiple JSON values")
	}
	if err := Validate(document); err != nil {
		return Document{}, err
	}
	return document, nil
}

// MarshalCanonical validates and serializes a deterministically ordered record.
func MarshalCanonical(document Document) ([]byte, error) {
	document.Selection.Stacks = append([]string{}, document.Selection.Stacks...)
	document.Files = append([]File(nil), document.Files...)
	document.Integrations = append([]Integration(nil), document.Integrations...)
	sort.Strings(document.Selection.Stacks)
	sort.Slice(document.Files, func(i, j int) bool { return document.Files[i].Path < document.Files[j].Path })
	sort.Slice(document.Integrations, func(i, j int) bool { return document.Integrations[i].ID < document.Integrations[j].ID })
	for index := range document.Integrations {
		document.Integrations[index].ManagedFiles = append([]string(nil), document.Integrations[index].ManagedFiles...)
		sort.Strings(document.Integrations[index].ManagedFiles)
	}
	if err := Validate(document); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}
	return append(data, '\n'), nil
}

// Validate enforces the semantic rules that JSON Schema cannot express alone.
func Validate(document Document) error {
	if document.SchemaVersion != SchemaVersion {
		return &UnsupportedSchemaError{Version: document.SchemaVersion}
	}
	if strings.TrimSpace(document.SokuVersion) == "" {
		return errors.New("soku_version is required")
	}
	if err := validatePortableSource("boilerplate.source", document.Boilerplate.Source); err != nil {
		return err
	}
	if strings.TrimSpace(document.Boilerplate.Release) == "" {
		return errors.New("boilerplate.release is required")
	}
	if !isLowerSHA(document.Boilerplate.ResolvedCommit) {
		return errors.New("boilerplate.resolved_commit must be a lowercase 40-character commit")
	}
	if strings.TrimSpace(document.Selection.Profile) == "" {
		return errors.New("selection.profile is required")
	}
	if err := validateSortedUnique("selection.stacks", document.Selection.Stacks); err != nil {
		return err
	}
	if !isSHA256(document.Selection.ConfigurationHash) {
		return errors.New("selection.configuration_hash must be a lowercase SHA-256")
	}
	for _, input := range []struct {
		name     string
		value    string
		required bool
	}{
		{"project_name", document.Selection.ProjectName, selectionUses(document.Selection.Stacks, "javascript-typescript-node", "python")},
		{"module_path", document.Selection.ModulePath, selectionUses(document.Selection.Stacks, "go")},
		{"java_group", document.Selection.JavaGroup, selectionUses(document.Selection.Stacks, "java-spring")},
		{"service_name", document.Selection.ServiceName, selectionUses(document.Selection.Stacks, "java-spring", "gcp")},
	} {
		if input.required && input.value == "" {
			return fmt.Errorf("selection.%s is required by the selected stacks", input.name)
		}
		if !input.required && input.value != "" {
			return fmt.Errorf("selection.%s is not used by the selected stacks", input.name)
		}
		if strings.ContainsAny(input.value, "\r\n\x00") || strings.HasPrefix(input.value, "/") || portableDrivePattern.MatchString(input.value) {
			return fmt.Errorf("selection.%s is not portable", input.name)
		}
	}
	configurationHash, err := HashSelection(document.Selection)
	if err != nil {
		return err
	}
	if configurationHash != document.Selection.ConfigurationHash {
		return errors.New("selection.configuration_hash does not match the canonical stored selection")
	}

	integrations := make(map[string]Integration, len(document.Integrations))
	previousID := ""
	for _, integration := range document.Integrations {
		if !integrationIDPattern.MatchString(integration.ID) {
			return fmt.Errorf("integration id %q is not portable", integration.ID)
		}
		if integration.ID <= previousID {
			return errors.New("integrations must be sorted by unique id")
		}
		previousID = integration.ID
		if err := validatePortableSource("integration.source", integration.Source); err != nil {
			return err
		}
		if !isLowerSHA(integration.Ref) {
			return fmt.Errorf("integration %q ref must be a lowercase 40-character commit", integration.ID)
		}
		if integration.ProviderAPIVersion == "" || integration.ProviderSchemaVersion == "" {
			return fmt.Errorf("integration %q provider versions are required", integration.ID)
		}
		if !isSHA256(integration.ConfigurationHash) {
			return fmt.Errorf("integration %q configuration_hash must be a lowercase SHA-256", integration.ID)
		}
		if !oneOf(integration.LifecycleState, "pending", "connected", "drifted", "incompatible") {
			return fmt.Errorf("integration %q has invalid lifecycle_state %q", integration.ID, integration.LifecycleState)
		}
		if err := validateSortedUnique("integration.managed_files", integration.ManagedFiles); err != nil {
			return fmt.Errorf("integration %q: %w", integration.ID, err)
		}
		integrations[integration.ID] = integration
	}

	paths := make(map[string]File, len(document.Files))
	previousPath := ""
	for _, file := range document.Files {
		if err := ValidatePath(file.Path); err != nil {
			return err
		}
		folded := strings.ToLower(file.Path)
		if _, exists := paths[folded]; exists {
			return fmt.Errorf("file path %q has a case-insensitive collision", file.Path)
		}
		if file.Path <= previousPath {
			return errors.New("files must be sorted by unique path")
		}
		previousPath = file.Path
		if !oneOf(file.Class, "core-managed", "provider-managed", "mergeable", "project-owned") {
			return fmt.Errorf("file %q has invalid class %q", file.Path, file.Class)
		}
		if !oneOf(file.LifecycleState, "current", "obsolete", "unmanaged-expected") {
			return fmt.Errorf("file %q has invalid lifecycle_state %q", file.Path, file.LifecycleState)
		}
		if file.Class == "project-owned" {
			if file.Owner != "project" {
				return fmt.Errorf("project-owned file %q must use owner project", file.Path)
			}
			if file.LifecycleState != "unmanaged-expected" {
				return fmt.Errorf("project-owned file %q must use lifecycle_state unmanaged-expected", file.Path)
			}
			if file.ContentMode != "" || file.BaselineSHA256 != "" {
				return fmt.Errorf("project-owned file %q must not contain a baseline", file.Path)
			}
		} else {
			if file.LifecycleState == "unmanaged-expected" {
				return fmt.Errorf("managed file %q cannot use lifecycle_state unmanaged-expected", file.Path)
			}
			if !oneOf(file.ContentMode, "text", "binary") || !isSHA256(file.BaselineSHA256) {
				return fmt.Errorf("managed file %q requires content_mode and baseline_sha256", file.Path)
			}
			if file.Owner == "" || file.Owner == "project" {
				return fmt.Errorf("managed file %q has invalid owner %q", file.Path, file.Owner)
			}
			if file.Class == "core-managed" && file.Owner != "core" {
				return fmt.Errorf("core-managed file %q must use owner core", file.Path)
			}
			if file.Class == "provider-managed" {
				if _, exists := integrations[file.Owner]; !exists {
					return fmt.Errorf("provider-managed file %q references unknown owner %q", file.Path, file.Owner)
				}
			}
			if file.Class == "mergeable" && file.Owner != "core" {
				if _, exists := integrations[file.Owner]; !exists {
					return fmt.Errorf("mergeable file %q references unknown owner %q", file.Path, file.Owner)
				}
			}
		}
		paths[folded] = file
	}

	referenced := make(map[string]string)
	for _, integration := range document.Integrations {
		for _, managedPath := range integration.ManagedFiles {
			if err := ValidatePath(managedPath); err != nil {
				return fmt.Errorf("integration %q: %w", integration.ID, err)
			}
			file, exists := paths[strings.ToLower(managedPath)]
			if !exists || file.Path != managedPath || file.Owner != integration.ID ||
				!oneOf(file.Class, "provider-managed", "mergeable") {
				return fmt.Errorf("integration %q has inconsistent managed file %q", integration.ID, managedPath)
			}
			if owner, exists := referenced[strings.ToLower(managedPath)]; exists {
				return fmt.Errorf("managed file %q is referenced by both %q and %q", managedPath, owner, integration.ID)
			}
			referenced[strings.ToLower(managedPath)] = integration.ID
		}
	}
	for _, file := range document.Files {
		if file.Owner != "core" && file.Owner != "project" {
			if referenced[strings.ToLower(file.Path)] != file.Owner {
				return fmt.Errorf("provider-owned file %q is not referenced by integration %q", file.Path, file.Owner)
			}
		}
	}
	return nil
}

func selectionUses(stacks []string, candidates ...string) bool {
	for _, stack := range stacks {
		for _, candidate := range candidates {
			if stack == candidate {
				return true
			}
		}
	}
	return false
}

// HashSelection returns the canonical hash of the portable rendering inputs.
// The stored hash deliberately excludes itself so it can be reproduced from a
// manifest without retaining raw configuration or machine-local values.
func HashSelection(selection Selection) (string, error) {
	data, err := json.Marshal(struct {
		Profile     string   `json:"profile"`
		Stacks      []string `json:"stacks"`
		ProjectName string   `json:"project_name,omitempty"`
		ModulePath  string   `json:"module_path,omitempty"`
		JavaGroup   string   `json:"java_group,omitempty"`
		ServiceName string   `json:"service_name,omitempty"`
	}{selection.Profile, selection.Stacks, selection.ProjectName, selection.ModulePath, selection.JavaGroup, selection.ServiceName})
	if err != nil {
		return "", fmt.Errorf("encode canonical selection: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// ValidatePath accepts only canonical repository-relative POSIX paths.
func ValidatePath(value string) error {
	if value == "" || strings.Contains(value, "\\") || strings.HasPrefix(value, "/") || path.IsAbs(value) {
		return fmt.Errorf("file path %q is not a canonical relative POSIX path", value)
	}
	if path.Clean(value) != value || value == "." {
		return fmt.Errorf("file path %q contains traversal or non-canonical components", value)
	}
	for _, component := range strings.Split(value, "/") {
		lower := strings.ToLower(component)
		if component == ".." || lower == ".git" || lower == ".soku" {
			return fmt.Errorf("file path %q enters reserved state", value)
		}
		if strings.ContainsAny(component, "<>:\"|?*") || strings.HasSuffix(component, ".") || strings.HasSuffix(component, " ") {
			return fmt.Errorf("file path %q contains a Windows-incompatible component", value)
		}
		base := strings.SplitN(lower, ".", 2)[0]
		if oneOf(base, "con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9") {
			return fmt.Errorf("file path %q contains a reserved Windows name", value)
		}
		for _, character := range component {
			if character < 32 {
				return fmt.Errorf("file path %q contains a control character", value)
			}
		}
	}
	return nil
}

// HashContent calculates the canonical baseline for the declared mode.
func HashContent(content []byte, mode string) (string, error) {
	switch mode {
	case "text":
		if !utf8.Valid(content) {
			return "", errors.New("text content is not valid UTF-8")
		}
		content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
		content = bytes.ReplaceAll(content, []byte("\r"), []byte("\n"))
	case "binary":
	default:
		return "", fmt.Errorf("unsupported content mode %q", mode)
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func validatePortableSource(field, value string) error {
	if value == "" || strings.ContainsAny(value, `\@?#`) {
		return fmt.Errorf("%s is not a portable source", field)
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("%s must be an absolute HTTPS URL", field)
	}
	if parsed.User != nil {
		return fmt.Errorf("%s must not contain credentials", field)
	}
	return nil
}

func validateSortedUnique(field string, values []string) error {
	previous := ""
	for _, value := range values {
		if value == "" || value <= previous {
			return fmt.Errorf("%s must be sorted and unique", field)
		}
		previous = value
	}
	return nil
}

func oneOf(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func isLowerSHA(value string) bool {
	return len(value) == 40 && value == strings.ToLower(value) && isHex(value)
}

func isSHA256(value string) bool {
	return len(value) == 64 && value == strings.ToLower(value) && isHex(value)
}

func isHex(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil
}
