package initcmd

import (
	"bytes"
	"encoding/json"
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
			content = generateCI(config.Stacks)
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

func generateCI(stacks []string) []byte {
	var b strings.Builder
	b.WriteString("name: CI\n\non:\n  pull_request:\n  push:\n    branches:\n      - main\n\njobs:\n")
	jobs := map[string]string{
		"javascript-typescript-node": "  javascript-typescript-node:\n    name: JS/TS/Node\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7\n      - uses: actions/setup-node@820762786026740c76f36085b0efc47a31fe5020 # v7\n        with:\n          node-version: \"22\"\n      - run: npm ci\n      - run: npm run lint\n      - run: npm run typecheck\n      - run: npm test\n      - run: npm run build\n      - run: npm run format:check\n",
		"python":                     "  python:\n    name: Python\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7\n      - uses: actions/setup-python@ece7cb06caefa5fff74198d8649806c4678c61a1 # v6\n        with:\n          python-version: \"3.12\"\n      - run: pip install -r requirements-lock.txt -e \".[dev]\"\n      - run: ruff check .\n      - run: mypy .\n      - run: pyink --check .\n      - run: pytest\n",
		"go":                         "  go:\n    name: Go\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7\n      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7\n        with:\n          go-version: \"1.26\"\n      - run: go test ./...\n",
		"java-spring":                "  java-spring:\n    name: Java/Spring\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7\n      - uses: actions/setup-java@03ad4de0992f5dab5e18fcb136590ce7c4a0ac95 # v5\n        with:\n          distribution: temurin\n          java-version: \"21\"\n      - run: mvn -B verify\n",
	}
	count := 0
	for _, stack := range stacks {
		if job := jobs[stack]; job != "" {
			if count > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(job)
			count++
		}
	}
	if count == 0 {
		b.WriteString("  configuration:\n    name: Configuration validation\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7\n      - run: echo \"Configuration files are validated by soku init\"\n")
	}
	return []byte(b.String())
}
