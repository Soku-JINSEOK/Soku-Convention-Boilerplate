package cicd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "cicd", name)
}

func TestDecodeConfigAcceptsPortableContract(t *testing.T) {
	data, err := os.ReadFile(fixturePath(t, filepath.Join("valid", "github-public.yml")))
	if err != nil {
		t.Fatal(err)
	}
	decision, err := DecodeConfig(data)
	if err != nil {
		t.Fatalf("DecodeConfig() error = %v", err)
	}
	if decision.SchemaVersion != 1 || decision.Mode != "ci-only" || decision.Verification.PR != "ci-quick" {
		t.Fatalf("unexpected decision: %#v", decision)
	}
}

func TestDecodeConfigRejectsUnknownFieldWithStableCode(t *testing.T) {
	data, err := os.ReadFile(fixturePath(t, filepath.Join("invalid", "unknown-field.yml")))
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecodeConfig(data)
	var planningError *Error
	if !errors.As(err, &planningError) || planningError.Code != "ci-cd.schema.invalid" {
		t.Fatalf("error = %#v, want ci-cd.schema.invalid", err)
	}
}

func TestDecodeConfigRejectsUnsortedCapabilitiesAndDelivery(t *testing.T) {
	base := `schema_version: 1
mode: ci-only
source_host: github.com
workload: library
artifact: package
required_os: [linux]
network_scope: public
cloud_authority: none
operations_owner: absent
capabilities: [native, hardware]
verification: {pr: ci-quick, full: full}
delivery: {enabled: false}
`
	if _, err := DecodeConfig([]byte(base)); err == nil {
		t.Fatal("unsorted capabilities unexpectedly accepted")
	}
	delivery := []byte(`schema_version: 1
mode: ci-only
source_host: github.com
workload: library
artifact: package
required_os: [linux]
network_scope: public
cloud_authority: none
operations_owner: absent
capabilities: []
verification: {pr: ci-quick, full: full}
delivery: {enabled: true}
`)
	if _, err := DecodeConfig(delivery); err == nil {
		t.Fatal("delivery mutation unexpectedly accepted")
	}
}
