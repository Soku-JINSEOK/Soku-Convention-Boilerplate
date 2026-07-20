package initcmd

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommandRunner func(context.Context, string, []string) error

func verifyPlan(ctx context.Context, root string, changes []Change, stacks []string, runner CommandRunner) ([]Verification, error) {
	stage, err := os.MkdirTemp("", "soku-init-verify-")
	if err != nil {
		return nil, fail(2, "verification.failed", "create verification tree: %v", err)
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err := copyProject(root, stage); err != nil {
		return nil, fail(2, "verification.failed", "stage project: %v", err)
	}
	for _, change := range changes {
		target := filepath.Join(stage, filepath.FromSlash(change.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, change.Content, 0o644); err != nil {
			return nil, err
		}
	}
	if runner == nil {
		runner = func(ctx context.Context, directory string, argv []string) error {
			command := exec.CommandContext(ctx, argv[0], argv[1:]...)
			command.Dir = directory
			command.Stdout = os.Stderr
			command.Stderr = os.Stderr
			return command.Run()
		}
	}
	commands := map[string][][]string{"javascript-typescript-node": {{"npm", "ci"}, {"npm", "run", "lint"}, {"npm", "run", "typecheck"}, {"npm", "test"}, {"npm", "run", "build"}, {"npm", "run", "format:check"}}, "python": {{"python", "-m", "pip", "install", "-r", "requirements-lock.txt", "-e", ".[dev]"}, {"ruff", "check", "."}, {"mypy", "."}, {"black", "--check", "."}, {"pytest"}}, "go": {{"go", "test", "./..."}}, "java-spring": {{"mvn", "-B", "verify"}}}
	var report []Verification
	for _, stack := range stacks {
		if internalStack(stack) {
			if err := verifyDeclarative(stage, stack); err != nil {
				return report, fail(2, "verification.failed", "%s validation failed: %v", stack, err)
			}
			report = append(report, Verification{Stack: stack, Command: []string{"soku", "internal-validate", stack}, Status: "passed"})
			continue
		}
		for _, argv := range commands[stack] {
			item := Verification{Stack: stack, Command: argv, Status: "passed"}
			if err := runner(ctx, stage, argv); err != nil {
				item.Status = "failed"
				report = append(report, item)
				return report, fail(2, "verification.failed", "verification command %q failed: %v", strings.Join(argv, " "), err)
			}
			report = append(report, item)
		}
	}
	return report, nil
}

func copyProject(root, target string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		if relative == "." {
			return nil
		}
		relative = filepath.ToSlash(relative)
		if relative == ".git" || strings.HasPrefix(relative, ".git/") || relative == ".soku" || strings.HasPrefix(relative, ".soku/") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("verification input contains a symbolic link")
		}
		destination := filepath.Join(target, filepath.FromSlash(relative))
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		return os.WriteFile(destination, data, 0o644)
	})
}
func internalStack(stack string) bool {
	return contains([]string{"mysql", "postgresql", "gcp", "aws", "azure"}, stack)
}
func verifyDeclarative(root, stack string) error {
	paths := map[string][]string{"mysql": {"db/mysql/schema.sql"}, "postgresql": {"db/postgresql/schema.sql"}, "gcp": {"cloudbuild.yaml", "Dockerfile"}, "aws": {"buildspec.yml"}, "azure": {"azure-pipelines.yml"}}
	for _, relative := range paths[stack] {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return err
		}
		text := string(data)
		if strings.Contains(text, "your-service") || strings.Contains(text, "your-project-name") || strings.Contains(text, "{{") {
			return errors.New("unresolved placeholder")
		}
		if strings.HasSuffix(relative, ".json") {
			var value any
			if err := json.Unmarshal(data, &value); err != nil {
				return err
			}
		}
		if strings.TrimSpace(text) == "" {
			return errors.New("empty configuration file")
		}
		switch stack {
		case "mysql", "postgresql":
			if !strings.Contains(strings.ToUpper(text), "CREATE TABLE") || !strings.Contains(text, ";") {
				return errors.New("schema must contain a terminated CREATE TABLE statement")
			}
		case "gcp":
			if relative == "cloudbuild.yaml" && (!strings.Contains(text, "steps:") || !strings.Contains(text, "images:") || !strings.Contains(text, "substitutions:")) {
				return errors.New("cloud Build configuration is missing required fields")
			}
			if relative == "Dockerfile" && !strings.Contains(strings.ToUpper(text), "FROM ") {
				return errors.New("dockerfile is missing FROM")
			}
		case "aws":
			if !strings.Contains(text, "version:") || !strings.Contains(text, "phases:") {
				return errors.New("CodeBuild configuration is missing required fields")
			}
		case "azure":
			if !strings.Contains(text, "trigger:") || !strings.Contains(text, "pool:") || !strings.Contains(text, "steps:") {
				return errors.New("azure pipeline is missing required fields")
			}
		}
	}
	return nil
}
