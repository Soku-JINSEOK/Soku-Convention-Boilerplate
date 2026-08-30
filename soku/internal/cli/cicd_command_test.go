package cli

import (
	"context"
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
	var request Request
	handlers := defaultHandlers()
	handlers.CICDInit = HandlerFunc(func(_ context.Context, got Request) error {
		request = got
		return nil
	})
	result := execute([]string{"ci-cd", "init", "--config", config, "--dry-run"}, testRuntime{}, handlers)
	if result.code != 0 || !request.DryRun || request.Yes || request.Command != "ci-cd init" {
		t.Fatalf("init request result = %#v request=%#v", result, request)
	}

	invalidFlags := execute([]string{"ci-cd", "init", "--config", config}, testRuntime{}, defaultHandlers())
	if invalidFlags.code != int(ExitValidationFailure) || !strings.Contains(invalidFlags.stderr, "exactly one") {
		t.Fatalf("init flag result = %#v", invalidFlags)
	}
}
