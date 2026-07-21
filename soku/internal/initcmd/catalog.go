package initcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

var unresolvedTokenPattern = regexp.MustCompile(`\{\{[a-z_]+\}\}`)

var catalogOutputs = map[string][]string{
	"shared": {".editorconfig", ".gitignore", ".github/workflows/ci.yml"},
	"aws":    {"buildspec.yml"}, "azure": {"azure-pipelines.yml"}, "gcp": {"Dockerfile", "cloudbuild.yaml"},
	"go":                         {".golangci.yml", "Makefile", "go.mod", "profile.go", "profile_test.go"},
	"java-spring":                {"pom.xml", "src/main/java/{{java_group_path}}/profile/Application.java", "src/main/java/{{java_group_path}}/profile/ProfileMapper.java", "src/main/java/{{java_group_path}}/profile/User.java", "src/main/java/{{java_group_path}}/profile/UserResponse.java", "src/main/resources/application.yml", "src/test/java/{{java_group_path}}/profile/ProfileMapperTest.java"},
	"javascript-typescript-node": {".prettierignore", "eslint.config.mjs", "package-lock.json", "package.json", "prettier.config.cjs", "src/profile.ts", "test/profile.test.ts", "tsconfig.json", "vitest.config.ts"},
	"mysql":                      {"db/mysql/schema.sql"}, "postgresql": {"db/postgresql/schema.sql"},
	"python": {"pyproject.toml", "requirements-lock.txt", "src/user_profile.py", "tests/test_user_profile.py"},
}

func DecodeCatalog(data []byte) (Catalog, error) {
	var catalog Catalog
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return Catalog{}, fail(5, "catalog.incompatible", "decode core catalog: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Catalog{}, fail(5, "catalog.incompatible", "core catalog contains multiple JSON values")
	}
	if catalog.SchemaVersion != 1 || catalog.Profile != ProfileStandard {
		return Catalog{}, fail(5, "catalog.incompatible", "unsupported core catalog version or profile")
	}
	if len(catalog.Stacks) != len(StackIDs) {
		return Catalog{}, fail(5, "catalog.incompatible", "core catalog must define exactly nine stacks")
	}
	known := map[string]bool{}
	for _, id := range StackIDs {
		known[id] = true
	}
	seenStacks, seenOutputs := map[string]bool{}, map[string]string{}
	if len(catalog.Files) != 3 {
		return Catalog{}, fail(5, "catalog.incompatible", "core catalog must define exactly three shared files")
	}
	if err := validateDeclarations("shared", catalog.Files, seenOutputs); err != nil {
		return Catalog{}, err
	}
	for _, stack := range catalog.Stacks {
		if !known[stack.ID] || seenStacks[stack.ID] || len(stack.Markers) == 0 || len(stack.Files) == 0 {
			return Catalog{}, fail(5, "catalog.incompatible", "core catalog contains an invalid or repeated stack")
		}
		seenStacks[stack.ID] = true
		for _, marker := range stack.Markers {
			if err := validateOutputPath(marker, false); err != nil {
				return Catalog{}, fail(5, "catalog.incompatible", "stack %s marker: %v", stack.ID, err)
			}
		}
		if err := validateDeclarations(stack.ID, stack.Files, seenOutputs); err != nil {
			return Catalog{}, err
		}
	}
	return catalog, nil
}

func validateDeclarations(scope string, files []CatalogFile, seenOutputs map[string]string) error {
	allowed := map[string]bool{}
	for _, output := range catalogOutputs[scope] {
		allowed[output] = true
	}
	if (scope == "shared" && len(files) != len(allowed)) || len(files) > len(allowed) {
		return fail(5, "catalog.incompatible", "%s declarations exceed the bounded standard output set", scope)
	}
	for _, file := range files {
		if file.Owner != "core" || (file.Class != "core-managed" && file.Class != "mergeable") || (file.ContentMode != "text" && file.ContentMode != "binary") || !contains([]string{"render", "gitignore-merge", "editorconfig-merge"}, file.Strategy) {
			return fail(5, "catalog.incompatible", "%s contains an invalid file declaration", scope)
		}
		if err := validateOutputPath(file.Source, false); err != nil {
			return fail(5, "catalog.incompatible", "invalid source path %q", file.Source)
		}
		if err := validateOutputPath(file.Output, true); err != nil {
			return fail(5, "catalog.incompatible", "invalid output path %q", file.Output)
		}
		if !allowed[file.Output] {
			return fail(5, "catalog.incompatible", "%s output %q is outside the standard profile", scope, file.Output)
		}
		if scope != "shared" && (file.Class != "core-managed" || file.Strategy != "render" || !strings.HasPrefix(file.Source, "templates/"+templateDirectory(scope)+"/")) {
			return fail(5, "catalog.incompatible", "%s declaration %q has an invalid ownership or source boundary", scope, file.Output)
		}
		if scope == "shared" {
			expectedStrategy := map[string]string{".editorconfig": "editorconfig-merge", ".gitignore": "gitignore-merge", ".github/workflows/ci.yml": "render"}
			expectedSource := map[string]string{".editorconfig": ".editorconfig", ".gitignore": ".gitignore", ".github/workflows/ci.yml": "templates/_shared/ci/downstream-ci.yml"}
			expectedClass := map[string]string{".editorconfig": "mergeable", ".gitignore": "mergeable", ".github/workflows/ci.yml": "core-managed"}
			if file.Strategy != expectedStrategy[file.Output] || file.Source != expectedSource[file.Output] || file.Class != expectedClass[file.Output] {
				return fail(5, "catalog.incompatible", "shared output %q has an invalid strategy, source, or class", file.Output)
			}
		}
		folded := strings.ToLower(file.Output)
		if owner, exists := seenOutputs[folded]; exists {
			return fail(5, "catalog.incompatible", "output %q is declared by both %s and %s", file.Output, owner, scope)
		}
		seenOutputs[folded] = scope
		for _, placeholder := range file.Placeholders {
			if !contains([]string{"project_name", "module_path", "java_group", "service_name"}, placeholder) {
				return fail(5, "catalog.incompatible", "unsupported placeholder %q", placeholder)
			}
		}
	}
	return nil
}

func templateDirectory(stack string) string {
	if stack == "gcp" {
		return "gcloud"
	}
	return stack
}

func validateOutputPath(value string, allowJavaToken bool) error {
	check := value
	if allowJavaToken {
		check = strings.ReplaceAll(check, "{{java_group_path}}", "com/example")
	}
	return manifest.ValidatePath(check)
}

func renderCatalog(snapshot SourceSnapshot, catalog Catalog, config Config) ([]Change, error) {
	values := map[string]string{"project_name": config.ProjectName, "module_path": config.ModulePath, "java_group": config.JavaGroup, "java_group_path": strings.ReplaceAll(config.JavaGroup, ".", "/"), "service_name": config.ServiceName}
	var changes []Change
	folded := map[string]string{}
	for _, selected := range config.Stacks {
		var found *Stack
		for index := range catalog.Stacks {
			if catalog.Stacks[index].ID == selected {
				found = &catalog.Stacks[index]
				break
			}
		}
		if found == nil {
			return nil, fail(5, "catalog.incompatible", "selected stack %q is absent from catalog", selected)
		}
		for _, declaration := range found.Files {
			content, ok := snapshot.Files[declaration.Source]
			if !ok {
				return nil, fail(5, "catalog.incompatible", "catalog source %q is missing", declaration.Source)
			}
			output := replaceTokens(declaration.Output, values)
			rendered := replaceTemplateContent(content, values)
			if unresolvedTokenPattern.MatchString(output) || unresolvedTokenPattern.Match(rendered) {
				return nil, fail(2, "render.invalid", "unresolved placeholder in %q", declaration.Output)
			}
			if err := manifest.ValidatePath(output); err != nil {
				return nil, fail(2, "path.invalid", "%v", err)
			}
			lower := strings.ToLower(output)
			if previous, exists := folded[lower]; exists {
				return nil, fail(4, "ownership.conflict", "rendered paths %q and %q collide", previous, output)
			}
			folded[lower] = output
			hash, err := manifest.HashContent(rendered, declaration.ContentMode)
			if err != nil {
				return nil, fail(2, "render.invalid", "render %q: %v", output, err)
			}
			changes = append(changes, Change{Path: output, Action: "create", Owner: declaration.Owner, Class: declaration.Class, ContentMode: declaration.ContentMode, BaselineSHA256: hash, Content: rendered})
		}
	}
	shared, err := renderShared(snapshot, catalog, config, values)
	if err != nil {
		return nil, err
	}
	changes = append(changes, shared...)
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes, nil
}

func renderShared(snapshot SourceSnapshot, catalog Catalog, config Config, values map[string]string) ([]Change, error) {
	var result []Change
	for _, item := range catalog.Files {
		content, ok := snapshot.Files[item.Source]
		if !ok {
			return nil, fail(5, "catalog.incompatible", "source release is missing %s", item.Source)
		}
		if item.Output == ".github/workflows/ci.yml" {
			var err error
			content, err = renderDownstreamCI(content, config.Stacks)
			if err != nil {
				return nil, err
			}
		}
		hash, err := manifest.HashContent(content, item.ContentMode)
		if err != nil {
			return nil, err
		}
		result = append(result, Change{Path: item.Output, Action: item.Strategy, Owner: item.Owner, Class: item.Class, ContentMode: item.ContentMode, BaselineSHA256: hash, Content: content})
	}
	_ = values
	return result, nil
}

func replaceTokens(value string, values map[string]string) string {
	for key, replacement := range values {
		value = strings.ReplaceAll(value, "{{"+key+"}}", replacement)
	}
	return value
}
func replaceTemplateContent(content []byte, values map[string]string) []byte {
	text := string(content)
	text = strings.ReplaceAll(text, "your-project-name", values["project_name"])
	text = strings.ReplaceAll(text, "github.com/your-org/your-repo", values["module_path"])
	text = strings.ReplaceAll(text, "com.example", values["java_group"])
	text = strings.ReplaceAll(text, "your-service", values["service_name"])
	return []byte(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"))
}

var ciJobMarkerPattern = regexp.MustCompile(`^# soku:job-(begin|end) ([a-z0-9-]+)$`)
var legacyCIJobPattern = regexp.MustCompile(`^  # ([a-z0-9-]+):$`)

var ciJobIDs = map[string]bool{
	"configuration":              true,
	"go":                         true,
	"java-spring":                true,
	"javascript-typescript-node": true,
	"python":                     true,
}

type ciJobBlock struct {
	start int
	end   int
}

func renderDownstreamCI(source []byte, stacks []string) ([]byte, error) {
	text := strings.ReplaceAll(strings.ReplaceAll(string(source), "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")
	blocks, marked, err := parseMarkedCIJobBlocks(lines)
	if err != nil {
		return nil, fail(5, "catalog.incompatible", "parse downstream CI source: %v", err)
	}
	legacySource := false
	if !marked {
		blocks, marked = parseLegacyCIJobBlocks(lines)
		legacySource = marked
	}
	if !marked {
		return nil, fail(5, "catalog.incompatible", "downstream CI source has no complete job blocks")
	}

	selected := map[string]bool{}
	for _, stack := range stacks {
		if ciJobIDs[stack] && stack != "configuration" {
			selected[stack] = true
		}
	}
	if len(selected) == 0 {
		selected["configuration"] = true
	}
	legacyConfiguration := legacySource && selected["configuration"]

	var rendered []string
	active := ""
	for _, line := range lines {
		if match := ciJobMarkerPattern.FindStringSubmatch(line); match != nil {
			if match[1] == "begin" {
				active = match[2]
			} else {
				active = ""
			}
			continue
		}
		if legacySource {
			legacy := legacyCIJobPattern.FindStringSubmatch(line)
			if legacy == nil {
				goto renderLine
			}
			active = legacy[1]
			continue
		}
	renderLine:
		if active != "" {
			if !selected[active] {
				continue
			}
			uncommented, err := uncommentCIJobLine(line)
			if err != nil {
				return nil, fail(5, "catalog.incompatible", "CI job %s contains an un-commented line: %v", active, err)
			}
			rendered = append(rendered, uncommented)
			continue
		}
		rendered = append(rendered, line)
	}
	if legacyConfiguration {
		rendered = append(rendered,
			"  configuration:",
			"    name: Configuration validation",
			"    runs-on: ubuntu-latest",
			"    steps:",
			"      - uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7",
			"      - run: echo 'Configuration files are validated by soku init'",
			"")
	}
	_ = blocks
	return []byte(strings.Join(rendered, "\n")), nil
}

func parseMarkedCIJobBlocks(lines []string) (map[string]ciJobBlock, bool, error) {
	blocks := map[string]ciJobBlock{}
	open := map[string]int{}
	marked := false
	for index, line := range lines {
		if !strings.HasPrefix(line, "# soku:job-") {
			if len(open) != 0 && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "  #") {
				return nil, true, fmt.Errorf("job block contains un-commented line %q", line)
			}
			continue
		}
		marked = true
		match := ciJobMarkerPattern.FindStringSubmatch(line)
		if match == nil {
			return nil, true, fmt.Errorf("invalid marker %q", line)
		}
		kind, id := match[1], match[2]
		if !ciJobIDs[id] {
			return nil, true, fmt.Errorf("unknown job %q", id)
		}
		if kind == "begin" {
			if len(open) != 0 {
				return nil, true, fmt.Errorf("nested job block at line %d", index+1)
			}
			if _, exists := open[id]; exists {
				return nil, true, fmt.Errorf("duplicate begin marker for %q", id)
			}
			if _, exists := blocks[id]; exists {
				return nil, true, fmt.Errorf("duplicate block for %q", id)
			}
			open[id] = index
			continue
		}
		start, exists := open[id]
		if !exists || index <= start {
			return nil, true, fmt.Errorf("end marker without begin for %q", id)
		}
		blocks[id] = ciJobBlock{start: start, end: index}
		delete(open, id)
	}
	if !marked {
		return nil, false, nil
	}
	if len(open) != 0 {
		return nil, true, fmt.Errorf("unclosed job marker")
	}
	if len(blocks) != len(ciJobIDs) {
		return nil, true, fmt.Errorf("expected %d job blocks, found %d", len(ciJobIDs), len(blocks))
	}
	return blocks, true, nil
}

func parseLegacyCIJobBlocks(lines []string) (map[string]ciJobBlock, bool) {
	starts := map[string]int{}
	for index, line := range lines {
		match := legacyCIJobPattern.FindStringSubmatch(line)
		if match == nil || !ciJobIDs[match[1]] {
			continue
		}
		if _, exists := starts[match[1]]; exists {
			return nil, false
		}
		starts[match[1]] = index
	}
	if len(starts) != len(ciJobIDs)-1 {
		return nil, false
	}
	if _, exists := starts["configuration"]; exists {
		return nil, false
	}
	ids := make([]string, 0, len(starts))
	for id := range starts {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return starts[ids[i]] < starts[ids[j]] })
	blocks := make(map[string]ciJobBlock, len(ids))
	for index, id := range ids {
		end := len(lines)
		if index+1 < len(ids) {
			end = starts[ids[index+1]]
		}
		blocks[id] = ciJobBlock{start: starts[id], end: end}
		if err := validateCIJobLines(lines, starts[id]+1, end); err != nil {
			return nil, false
		}
	}
	return blocks, true
}

func validateCIJobLines(lines []string, start, end int) error {
	for _, line := range lines[start:end] {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "  #") {
			continue
		}
		return fmt.Errorf("job block contains un-commented line %q", line)
	}
	return nil
}

func uncommentCIJobLine(line string) (string, error) {
	if line == "" {
		return line, nil
	}
	if strings.HasPrefix(line, "  # ") {
		return line[:2] + line[4:], nil
	}
	if line == "  #" {
		return "  ", nil
	}
	if strings.HasPrefix(line, "# ") {
		return line[2:], nil
	}
	if line == "#" {
		return "", nil
	}
	return "", fmt.Errorf("line %q is not commented", line)
}
