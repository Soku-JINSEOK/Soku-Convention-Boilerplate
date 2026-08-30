package cicd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(workingDirectory, "..", "..", ".."))
}

func TestBuildPlanIsDeterministicAndPortable(t *testing.T) {
	root := repositoryRoot(t)
	config := filepath.Join(root, "soku", "testdata", "cicd", "valid", "github-public.yml")
	first, err := BuildPlan(root, config)
	if err != nil {
		t.Fatalf("first BuildPlan() error = %v", err)
	}
	second, err := BuildPlan(root, config)
	if err != nil {
		t.Fatalf("second BuildPlan() error = %v", err)
	}
	firstJSON, err := CanonicalJSON(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := CanonicalJSON(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) || first.PlanDigest != second.PlanDigest {
		t.Fatal("repeated plans are not byte-stable")
	}
	if first.Platform != platformGitHubHosted || !first.Installability || first.AdapterResolution.Status != "resolved" {
		t.Fatalf("unexpected hosted plan: %#v", first)
	}
	if first.Verification.PR.Argv[0] != "scripts/verify.sh" || first.Verification.Full.Argv[2] != "full" {
		t.Fatalf("unexpected verification argv: %#v", first.Verification)
	}
	if strings.Contains(string(firstJSON), filepath.Base(root)) || strings.Contains(string(firstJSON), root) {
		t.Fatal("plan contains a repository path or name")
	}
	var decoded map[string]any
	if err := json.Unmarshal(firstJSON, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["plan_digest"] != first.PlanDigest {
		t.Fatalf("digest not present in canonical output: %#v", decoded["plan_digest"])
	}
}

func TestBuildPlanSelectionRules(t *testing.T) {
	root := repositoryRoot(t)
	cases := []struct {
		name     string
		fixture  string
		platform string
	}{
		{name: "gcp", fixture: "gcp-private.yml", platform: platformGCPManaged},
		{name: "specialized", fixture: "github-specialized.yml", platform: platformGitHubSelfHosted},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			plan, err := BuildPlan(root, filepath.Join(root, "soku", "testdata", "cicd", "valid", test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if plan.Platform != test.platform || !plan.Installability || plan.AdapterResolution.Status != "resolved" {
				t.Fatalf("plan = %#v", plan)
			}
		})
	}
}

func TestBuildPlanRejectsUnsupportedAuthority(t *testing.T) {
	root := repositoryRoot(t)
	data := []byte(`schema_version: 1
mode: ci-only
source_host: github.com
workload: api
artifact: none
required_os: [linux]
network_scope: public
cloud_authority: aws
operations_owner: declared
capabilities: []
verification: {pr: ci-quick, full: full}
delivery: {enabled: false}
`)
	config := filepath.Join(t.TempDir(), "decision.yml")
	if err := os.WriteFile(config, data, 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(root, config)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Platform != platformUndecided || plan.Installability || plan.AdapterResolution.Status != "undecided" {
		t.Fatalf("unexpected unsupported plan: %#v", plan)
	}
}
