package initcmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRegisteredProviderLiteralByteProvenance(t *testing.T) {
	root := filepath.Clean("../../..")
	path := filepath.Join(
		root, "soku/providers/provenance/registered-downstream-v1.json",
	)
	var ledger struct {
		SchemaVersion int    `json:"schema_version"`
		HashAlgorithm string `json:"hash_algorithm"`
		ControlPlane  struct {
			MergeCommit string `json:"merge_commit"`
		} `json:"control_plane"`
		SourceRewrite   string `json:"source_rewrite"`
		DeliveryEnabled bool   `json:"delivery_enabled"`
		Bundles         []struct {
			ProviderID string `json:"provider_id"`
			Files      []struct {
				Path   string `json:"path"`
				SHA256 string `json:"sha256"`
			} `json:"files"`
		} `json:"bundles"`
	}
	if err := json.Unmarshal(mustRead(t, path), &ledger); err != nil {
		t.Fatal(err)
	}
	if ledger.SchemaVersion != 1 ||
		ledger.HashAlgorithm != "sha256-raw-bytes" ||
		ledger.ControlPlane.MergeCommit !=
			"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" ||
		ledger.SourceRewrite != "provider-v1.json:source-only" ||
		ledger.DeliveryEnabled || len(ledger.Bundles) != 4 {
		t.Fatalf("invalid registered provider provenance: %#v", ledger)
	}
	seen := map[string]bool{}
	for _, bundle := range ledger.Bundles {
		if seen[bundle.ProviderID] || len(bundle.Files) != 4 {
			t.Fatalf("invalid provider entry: %#v", bundle)
		}
		seen[bundle.ProviderID] = true
		for _, file := range bundle.Files {
			content, err := exec.Command(
				"git", "-C", root, "show", "HEAD:"+filepath.ToSlash(file.Path),
			).Output()
			if err != nil {
				t.Fatalf("read Git blob %s: %v", file.Path, err)
			}
			sum := sha256.Sum256(content)
			if hex.EncodeToString(sum[:]) != file.SHA256 {
				t.Fatalf("literal-byte hash mismatch: %s", file.Path)
			}
		}
	}
}
