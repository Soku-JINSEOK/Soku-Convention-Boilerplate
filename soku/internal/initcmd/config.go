package initcmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

var (
	releasePattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	projectPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]{0,212}[a-z0-9])?$`)
	servicePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	modulePattern  = regexp.MustCompile(`^[A-Za-z0-9._~/-]+$`)
	javaPattern    = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)+$`)
)

func LoadConfig(path string) (Config, error) {
	if path == "" {
		return Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fail(2, "configuration.invalid", "read configuration: %v", err)
	}
	config := Config{}
	seen := map[string]bool{}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for index := 0; index < len(lines); index++ {
		lineNumber := index + 1
		line := strings.TrimSpace(lines[index])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.TrimLeft(lines[index], " \t") != lines[index] {
			return Config{}, fail(2, "configuration.invalid", "configuration line %d is nested; only a flat mapping is supported", lineNumber)
		}
		key, raw, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) == "" {
			return Config{}, fail(2, "configuration.invalid", "configuration line %d is not a flat mapping", lineNumber)
		}
		key, raw = strings.TrimSpace(key), strings.TrimSpace(raw)
		if seen[key] {
			return Config{}, fail(2, "configuration.invalid", "configuration field %q is repeated", key)
		}
		seen[key] = true
		value, err := yamlScalar(raw)
		if err != nil {
			return Config{}, fail(2, "configuration.invalid", "configuration field %q: %v", key, err)
		}
		switch key {
		case "schema_version":
			config.SchemaVersion, err = strconv.Atoi(value)
		case "boilerplate_source":
			config.BoilerplateSource = value
		case "boilerplate_release":
			config.BoilerplateRelease = value
		case "profile":
			config.Profile = value
		case "project_name":
			config.ProjectName = value
		case "module_path":
			config.ModulePath = value
		case "java_group":
			config.JavaGroup = value
		case "service_name":
			config.ServiceName = value
		case "verify":
			config.Verify, err = strconv.ParseBool(value)
		case "stacks":
			if value == "" {
				for index+1 < len(lines) {
					next := strings.TrimSpace(lines[index+1])
					if !strings.HasPrefix(next, "-") {
						break
					}
					item, scalarErr := yamlScalar(strings.TrimSpace(strings.TrimPrefix(next, "-")))
					if scalarErr != nil || item == "" {
						return Config{}, fail(2, "configuration.invalid", "configuration field %q contains an invalid item", key)
					}
					config.Stacks = append(config.Stacks, item)
					index++
				}
			} else {
				config.Stacks, err = parseList(value)
			}
		default:
			return Config{}, fail(2, "configuration.invalid", "unknown configuration field %q", key)
		}
		if err != nil {
			return Config{}, fail(2, "configuration.invalid", "configuration field %q is invalid", key)
		}
	}
	if config.SchemaVersion != 1 {
		return Config{}, fail(2, "configuration.invalid", "schema_version must be 1")
	}
	return config, nil
}

func yamlScalar(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if strings.HasPrefix(raw, "[") {
		return raw, nil
	}
	if raw[0] == '\'' || raw[0] == '"' {
		if len(raw) < 2 || raw[len(raw)-1] != raw[0] {
			return "", errors.New("unterminated quoted scalar")
		}
		return raw[1 : len(raw)-1], nil
	}
	if strings.Contains(raw, " #") {
		raw = strings.SplitN(raw, " #", 2)[0]
	}
	return strings.TrimSpace(raw), nil
}

func parseList(raw string) ([]string, error) {
	if raw == "" {
		return []string{}, nil
	}
	if !strings.HasPrefix(raw, "[") || !strings.HasSuffix(raw, "]") {
		return nil, errors.New("must be an inline YAML sequence")
	}
	body := strings.TrimSpace(raw[1 : len(raw)-1])
	if body == "" {
		return []string{}, nil
	}
	parts := strings.Split(body, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		value, err := yamlScalar(strings.TrimSpace(part))
		if err != nil || value == "" {
			return nil, errors.New("contains an invalid item")
		}
		result = append(result, value)
	}
	return result, nil
}

func ResolveConfig(root string, file Config, cli Explicit, catalog Catalog) (Config, error) {
	resolved := file
	if resolved.SchemaVersion == 0 {
		resolved.SchemaVersion = 1
	}
	if cli.SourceSet {
		resolved.BoilerplateSource = cli.Source
	}
	if cli.ReleaseSet {
		resolved.BoilerplateRelease = cli.Release
	}
	if cli.StacksSet {
		resolved.Stacks = append([]string(nil), cli.Stacks...)
	}
	if cli.ProfileSet {
		resolved.Profile = cli.Profile
	}
	if cli.ProjectNameSet {
		resolved.ProjectName = cli.ProjectName
	}
	if cli.ModulePathSet {
		resolved.ModulePath = cli.ModulePath
	}
	if cli.JavaGroupSet {
		resolved.JavaGroup = cli.JavaGroup
	}
	if cli.ServiceNameSet {
		resolved.ServiceName = cli.ServiceName
	}
	if cli.VerifySet {
		resolved.Verify = cli.Verify
	}
	if resolved.Profile == "" {
		resolved.Profile = ProfileStandard
	}
	if resolved.Profile != ProfileStandard {
		return Config{}, fail(2, "selection.invalid", "profile must be standard")
	}
	if resolved.BoilerplateSource == "" || resolved.BoilerplateRelease == "" {
		return Config{}, fail(2, "selection.invalid", "--boilerplate-source and --boilerplate-release are required")
	}
	if !releasePattern.MatchString(resolved.BoilerplateRelease) {
		return Config{}, fail(2, "selection.invalid", "boilerplate_release must be an exact vMAJOR.MINOR.PATCH without prerelease")
	}
	if !cli.StacksSet && len(file.Stacks) == 0 {
		resolved.Stacks = detectStacks(root, catalog)
		if len(resolved.Stacks) == 0 {
			return Config{}, fail(2, "selection.missing", "no supported stack detected; select at least one with --stack")
		}
	}
	stacks, err := normalizeStacks(resolved.Stacks)
	if err != nil {
		return Config{}, err
	}
	resolved.Stacks = stacks
	base := strings.ToLower(filepath.Base(root))
	base = strings.ReplaceAll(base, "_", "-")
	if resolved.ProjectName == "" {
		resolved.ProjectName = base
	}
	if resolved.ServiceName == "" {
		resolved.ServiceName = strings.Trim(resolved.ProjectName, ".")
	}
	if contains(resolved.Stacks, "go") && resolved.ModulePath == "" {
		resolved.ModulePath = detectFirstLine(filepath.Join(root, "go.mod"), "module ")
	}
	if contains(resolved.Stacks, "java-spring") && resolved.JavaGroup == "" {
		resolved.JavaGroup = detectXMLValue(filepath.Join(root, "pom.xml"), "groupId")
	}
	if err := validatePlaceholders(resolved); err != nil {
		return Config{}, err
	}
	return resolved, nil
}

func detectStacks(root string, catalog Catalog) []string {
	var detected []string
	for _, stack := range catalog.Stacks {
		for _, marker := range stack.Markers {
			if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(marker))); err == nil && !info.IsDir() {
				detected = append(detected, stack.ID)
				break
			}
		}
	}
	result, _ := normalizeStacks(detected)
	return result
}

func validatePlaceholders(config Config) error {
	if (contains(config.Stacks, "javascript-typescript-node") || contains(config.Stacks, "python")) && !projectPattern.MatchString(config.ProjectName) {
		return fail(2, "configuration.invalid", "project_name is invalid for the selected stack")
	}
	if contains(config.Stacks, "go") && (config.ModulePath == "" || !modulePattern.MatchString(config.ModulePath) || strings.Contains(config.ModulePath, "..")) {
		return fail(2, "configuration.invalid", "module_path is required and must be a valid Go module path")
	}
	if contains(config.Stacks, "java-spring") && !javaPattern.MatchString(config.JavaGroup) {
		return fail(2, "configuration.invalid", "java_group is required and must be a dotted lowercase Java package")
	}
	if (contains(config.Stacks, "java-spring") || contains(config.Stacks, "gcp")) && !servicePattern.MatchString(config.ServiceName) {
		return fail(2, "configuration.invalid", "service_name is invalid for the selected stack")
	}
	return nil
}

func detectFirstLine(path, prefix string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), prefix))
		}
	}
	return ""
}
func detectXMLValue(path, name string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	start, end := "<"+name+">", "</"+name+">"
	text := string(data)
	if name == "groupId" {
		if parentEnd := strings.Index(text, "</parent>"); parentEnd >= 0 {
			text = text[parentEnd+len("</parent>"):]
		}
	}
	a := strings.Index(text, start)
	if a < 0 {
		return ""
	}
	text = text[a+len(start):]
	b := strings.Index(text, end)
	if b < 0 {
		return ""
	}
	return strings.TrimSpace(text[:b])
}
func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
func ensureNoState(root string) error {
	entries, err := os.ReadDir(filepath.Join(root, ".soku", "transactions"))
	if err == nil && len(entries) > 0 {
		return fail(8, "recovery.required", "an unfinished transaction exists; preserve .soku/transactions and run soku status")
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fail(8, "recovery.required", "cannot inspect transaction state: %v", err)
	}
	return nil
}
func configHash(config Config) (string, error) {
	selection := selectionFromConfig(config)
	return manifest.HashSelection(selection)
}

func selectionFromConfig(config Config) manifest.Selection {
	return manifest.Selection{
		Profile:     config.Profile,
		Stacks:      append([]string(nil), config.Stacks...),
		ProjectName: used(config, "project_name", config.ProjectName),
		ModulePath:  used(config, "module_path", config.ModulePath),
		JavaGroup:   used(config, "java_group", config.JavaGroup),
		ServiceName: used(config, "service_name", config.ServiceName),
	}
}

func configFromSelection(selection manifest.Selection) Config {
	return Config{
		SchemaVersion: 1,
		Profile:       selection.Profile,
		Stacks:        append([]string(nil), selection.Stacks...),
		ProjectName:   selection.ProjectName,
		ModulePath:    selection.ModulePath,
		JavaGroup:     selection.JavaGroup,
		ServiceName:   selection.ServiceName,
	}
}
func used(c Config, key, value string) string {
	for _, id := range c.Stacks {
		switch key {
		case "project_name":
			if id == "javascript-typescript-node" || id == "python" {
				return value
			}
		case "module_path":
			if id == "go" {
				return value
			}
		case "java_group":
			if id == "java-spring" {
				return value
			}
		case "service_name":
			if id == "java-spring" || id == "gcp" {
				return value
			}
		}
	}
	return ""
}

var _ = fmt.Sprintf
