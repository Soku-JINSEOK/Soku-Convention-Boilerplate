package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCICDHelpAndInitSafetyBoundary(t *testing.T) {
	help := execute([]string{"ci-cd", "--help"}, testRuntime{}, defaultHandlers())
	if help.code != 0 || !strings.Contains(help.stdout, "soku ci-cd <plan|init> [flags]") {
		t.Fatalf("help = %#v", help)
	}

	config := filepath.Join("..", "..", "testdata", "cicd", "valid", "github-public.yml")
	result := execute([]string{"ci-cd", "init", "--config", config, "--dry-run"}, testRuntime{}, defaultHandlers())
	if result.code != int(ExitSafetyRefusal) || !strings.Contains(result.stderr, "not available") {
		t.Fatalf("init safety result = %#v", result)
	}

	invalidFlags := execute([]string{"ci-cd", "init", "--config", config}, testRuntime{}, defaultHandlers())
	if invalidFlags.code != int(ExitValidationFailure) || !strings.Contains(invalidFlags.stderr, "exactly one") {
		t.Fatalf("init flag result = %#v", invalidFlags)
	}
}
