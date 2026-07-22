package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const stackJobsToken = "# {{STACK_JOBS}}"

var requiredTemplateJobs = []string{
	"mysql",
	"postgresql",
	"gcloud",
	"aws-azure-config",
}

var ciJobIDs = map[string]bool{
	"configuration":              true,
	"go":                         true,
	"java-spring":                true,
	"javascript-typescript-node": true,
	"python":                     true,
}

var renderJobOrder = []string{
	"javascript-typescript-node",
	"python",
	"go",
	"java-spring",
}

var markerPattern = regexp.MustCompile(`^# soku:job-(begin|end) ([a-z0-9-]+)$`)

var (
	defaultSourcePath   = "templates/_shared/ci/downstream-ci.yml"
	defaultTemplatePath = ".github/workflows/templates-ci.template.yml"
	defaultOutputPath   = ".github/workflows/templates-ci.yml"

	write    = flag.Bool("write", false, "Write output to --output")
	check    = flag.Bool("check", false, "Verify tracked template is already generated")
	source   = flag.String("source", defaultSourcePath, "Path to canonical downstream CI source")
	template = flag.String("template", defaultTemplatePath, "Path to templates-ci template")
	output   = flag.String("output", defaultOutputPath, "Path to generated templates-ci workflow")
)

func main() {
	flag.Parse()

	if *check && *write {
		panic("--check and --write cannot be used together")
	}

	outPath := *output
	rendered, err := render()
	if err != nil {
		panic(err)
	}

	if *check {
		existing, err := os.ReadFile(outPath)
		if err != nil {
			panic(err)
		}
		existingNormalized := normalizeText(string(existing))
		if strings.TrimSuffix(existingNormalized, "\n") != strings.TrimSuffix(rendered, "\n") {
			panic("generated templates-ci is out of date")
		}
		return
	}

	if *write {
		if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
			panic(err)
		}
		return
	}

	fmt.Print(rendered)
}

func render() (string, error) {
	sourceText, err := os.ReadFile(*source)
	if err != nil {
		return "", err
	}

	templateRaw, err := os.ReadFile(*template)
	if err != nil {
		return "", err
	}
	templateText := normalizeText(string(templateRaw))

	blocks, err := parseMarkedCIJobBlocks(string(sourceText))
	if err != nil {
		return "", fmt.Errorf("parse downstream CI source: %w", err)
	}

	stackJobs := make([]string, 0, len(renderJobOrder))
	for _, jobID := range renderJobOrder {
		lines, ok := blocks[jobID]
		if !ok {
			return "", fmt.Errorf("required job block %q is missing from %q", jobID, *source)
		}
		stackJobs = append(stackJobs, strings.Join(lines, "\n"))
	}

	if !strings.Contains(string(templateText), stackJobsToken) {
		return "", errors.New("template missing stack job token")
	}
	templateContent := strings.Replace(string(templateText), stackJobsToken, strings.Join(stackJobs, "\n\n"), 1)
	if err := validateRenderedWorkflow(templateContent); err != nil {
		return "", err
	}

	return ensureTrailingNewline(templateContent), nil
}

func validateRenderedWorkflow(content string) error {
	for _, jobID := range requiredTemplateJobs {
		pattern := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(jobID) + `:`)
		if !pattern.MatchString(content) {
			return fmt.Errorf("required templates-ci workflow job %q is missing", jobID)
		}
	}
	if strings.Contains(content, stackJobsToken) {
		return errors.New("templates-ci workflow token was not replaced")
	}
	return nil
}

func parseMarkedCIJobBlocks(text string) (map[string][]string, error) {
	lines := strings.Split(normalizeText(text), "\n")
	blocks := map[string][]string{}
	open := map[string]int{}
	marked := false

	for index, line := range lines {
		match := markerPattern.FindStringSubmatch(line)
		if match == nil {
			if strings.HasPrefix(line, "# soku:job-") {
				return nil, fmt.Errorf("invalid marker %q", line)
			}
			continue
		}

		marked = true
		kind, id := match[1], match[2]

		if !ciJobIDs[id] {
			return nil, fmt.Errorf("unknown job %q", id)
		}

		switch kind {
		case "begin":
			if len(open) != 0 {
				return nil, fmt.Errorf("nested job block at line %d", index+1)
			}
			if _, exists := open[id]; exists {
				return nil, fmt.Errorf("duplicate begin marker for %q", id)
			}
			if _, exists := blocks[id]; exists {
				return nil, fmt.Errorf("duplicate block for %q", id)
			}
			open[id] = index

		case "end":
			start, exists := open[id]
			if !exists || index <= start {
				return nil, fmt.Errorf("end marker without begin for %q", id)
			}
			if err := validateCIJobBlock(lines[start+1 : index]); err != nil {
				return nil, err
			}
			linesCopy := make([]string, 0, index-start)
			for _, raw := range lines[start+1 : index] {
				uncommented, err := uncommentCIJobLine(raw)
				if err != nil {
					return nil, fmt.Errorf("job %q: %w", id, err)
				}
				linesCopy = append(linesCopy, uncommented)
			}
			blocks[id] = linesCopy
			delete(open, id)
		}
	}

	if !marked {
		return nil, errors.New("downstream CI source has no complete job blocks")
	}
	if len(open) != 0 {
		for id := range open {
			return nil, fmt.Errorf("unclosed job marker for %q", id)
		}
	}

	if len(blocks) == 0 {
		return nil, errors.New("downstream CI source has no complete job blocks")
	}

	return blocks, nil
}

func validateCIJobBlock(lines []string) error {
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "  #") {
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
	return "", fmt.Errorf("job block contains un-commented line %q", line)
}

func ensureTrailingNewline(value string) string {
	return strings.TrimSuffix(value, "\n") + "\n"
}

func normalizeText(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n")
}
