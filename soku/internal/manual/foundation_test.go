package manual

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFoundationPublishedExampleBuildsDeterministicPlan(t *testing.T) {
	first, err := BuildPlan(".", "assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildPlan(".", "assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("plans differ:\n%#v\n%#v", first, second)
	}
	if first.Authenticity != "runtime-authentic" || len(first.Captures) != 1 {
		t.Fatalf("unexpected plan: %#v", first)
	}
}

func TestFoundationStrictConfigRejectsUnknownDuplicateAndSecretInputs(t *testing.T) {
	valid, err := os.ReadFile("assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(string) string
	}{
		{name: "unknown", mutate: func(value string) string { return value + "\nunknown: true\n" }},
		{name: "duplicate", mutate: func(value string) string {
			return strings.Replace(value, "schema_version: 1", "schema_version: 1\nschema_version: 1", 1)
		}},
		{name: "secret", mutate: func(value string) string {
			return value + "\napi_key: example\n"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(
				filepath.Join(root, "capture.yml"),
				[]byte(test.mutate(string(valid))),
				0o600,
			); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadConfig(root, "capture.yml"); err == nil {
				t.Fatal("unsafe configuration was accepted")
			}
		})
	}
}

func TestFoundationDoctorUsesFixedLocalChecks(t *testing.T) {
	root := t.TempDir()
	config, err := os.ReadFile("assets/examples/capture.example.yml")
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "capture.yml")
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := Doctor(context.Background(), root, "capture.yml", false)
	if err != nil {
		t.Fatal(err)
	}
	if report.Probe {
		t.Fatal("static doctor unexpectedly executed a probe")
	}
	if len(report.Checks) == 0 {
		t.Fatal("doctor returned no checks")
	}
}
