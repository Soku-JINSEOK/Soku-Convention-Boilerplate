package initcmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const publicProviderRoot = "../../providers/ci-cd-control-plane-v1"

const publicMirrorCommit = "c5435ea36d88dbe3b4b2c373265206943c53fcbf"

func TestPublishedControlPlaneProviderBundle(t *testing.T) {
	metadata := mustRead(t, publicProviderRoot+"/provider-v1.json")
	files := map[string][]byte{
		"provider-v1.json": metadata,
		"configuration.schema.json": mustRead(
			t, publicProviderRoot+"/configuration.schema.json",
		),
		"example-config.yml": mustRead(
			t, publicProviderRoot+"/example-config.yml",
		),
		"templates/control-plane-provider.yml": mustRead(
			t, publicProviderRoot+"/templates/control-plane-provider.yml",
		),
	}
	bundle, err := DecodeProviderBundle(metadata, files)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Ref != nil {
		t.Fatal("production provider metadata retained deprecated ref")
	}
	if bundle.Source != "github:Soku-JINSEOK/Soku-Convention-Boilerplate/soku/providers/ci-cd-control-plane-v1" {
		t.Fatalf("provider source = %q", bundle.Source)
	}
	configurationHash, err := integrationConfigurationHash(
		publicProviderRoot + "/example-config.yml",
	)
	if err != nil || configurationHash != bundle.ConfigurationHash {
		t.Fatalf("example configuration hash = %q, %v", configurationHash, err)
	}

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile("../../schema/provider-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(metadata))
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(instance); err != nil {
		t.Fatal(err)
	}

	files["unexpected.txt"] = []byte("unexpected\n")
	if _, err := DecodeProviderBundle(metadata, files); failureCode(err) != 5 {
		t.Fatalf("unknown bundle file error = %v", err)
	}
}

func TestControlPlaneProviderLiteralByteProvenance(t *testing.T) {
	root := filepath.Clean("../../..")
	path := filepath.Join(
		root,
		"soku/providers/provenance/ci-cd-control-plane-v1.json",
	)
	var ledger map[string]any
	if err := json.Unmarshal(mustRead(t, path), &ledger); err != nil {
		t.Fatal(err)
	}
	if ledger["schema_version"] != float64(1) ||
		ledger["provider_id"] != "ci-cd-control-plane-v1" ||
		ledger["hash_algorithm"] != "sha256-raw-bytes" ||
		ledger["delivery_enabled"] != false {
		t.Fatalf("invalid public provenance identity: %#v", ledger)
	}
	mirror, ok := ledger["public_mirror"].(map[string]any)
	if !ok || mirror["state"] != "published" ||
		mirror["commit"] != publicMirrorCommit {
		t.Fatalf("invalid published public mirror: %#v", mirror)
	}
	entries, ok := ledger["public_files"].([]any)
	if !ok || len(entries) != 8 {
		t.Fatalf("public provenance files = %#v", ledger["public_files"])
	}
	seen := map[string]bool{}
	for _, value := range entries {
		entry, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("invalid provenance entry: %#v", value)
		}
		entryPath, pathOK := entry["path"].(string)
		expected, hashOK := entry["sha256"].(string)
		if !pathOK || !hashOK || seen[entryPath] {
			t.Fatalf("invalid provenance entry: %#v", entry)
		}
		seen[entryPath] = true
		content, err := exec.Command(
			"git", "-C", root, "show", "HEAD:"+filepath.ToSlash(entryPath),
		).Output()
		if err != nil {
			t.Fatalf("read Git blob %s: %v", entryPath, err)
		}
		sum := sha256.Sum256(content)
		if hex.EncodeToString(sum[:]) != expected {
			t.Fatalf("literal-byte hash mismatch: %s", entryPath)
		}
	}
}

func TestControlPlaneProviderCallerContract(t *testing.T) {
	caller := string(mustRead(
		t,
		"../../../docs/callers/ci-cd-control-plane-v1.yml",
	))
	actionPin := "Soku-JINSEOK/Soku-Convention-Boilerplate/" +
		"soku/actions/ci-cd-control-plane-v1@" + publicMirrorCommit
	if strings.Count(caller, actionPin) != 1 {
		t.Fatalf("caller must use the immutable public action once")
	}
	for _, argument := range []string{
		"--integration-source '${{ steps.provider.outputs.integration-source }}'",
		"--integration-ref '${{ steps.provider.outputs.integration-ref }}'",
		"--integration-config '${{ steps.provider.outputs.integration-config }}'",
	} {
		if strings.Count(caller, argument) != 1 {
			t.Fatalf("caller argument %q must appear once", argument)
		}
	}
	for _, forbidden := range []string{
		"secrets.", "token:", "password", "private", "curl ", "wget ",
		"git clone", "go run", "python ", "bash -c", "sh -c",
	} {
		if strings.Contains(strings.ToLower(caller), forbidden) {
			t.Fatalf("caller contains forbidden capability %q", forbidden)
		}
	}
}
