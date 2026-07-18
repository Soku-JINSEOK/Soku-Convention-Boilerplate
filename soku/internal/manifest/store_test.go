package manifest

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreWriteLoadAndMode(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)
	document := validDocument()
	if err := store.Write(document); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil || loaded.SokuVersion != document.SokuVersion {
		t.Fatalf("Load() = %#v, %v", loaded, err)
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(ManifestPath)))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("manifest mode = %o, want 600", info.Mode().Perm())
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(PendingPath))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("pending state remains: %v", err)
	}

	document.SokuVersion = "v0.2.1"
	if err := store.Write(document); err != nil {
		t.Fatal(err)
	}
	loaded, err = store.Load()
	if err != nil || loaded.SokuVersion != "v0.2.1" {
		t.Fatalf("replacement = %#v, %v", loaded, err)
	}
}

func TestStoreRecoveryRulesPreserveAmbiguousState(t *testing.T) {
	tests := []struct {
		name         string
		manifest     []byte
		pending      []byte
		wantVersion  string
		wantError    bool
		pendingAfter bool
	}{
		{name: "promote pending", pending: encodedDocument(t, "pending"), wantVersion: "pending"},
		{name: "discard pending beside valid manifest", manifest: encodedDocument(t, "durable"), pending: encodedDocument(t, "pending"), wantVersion: "durable"},
		{name: "preserve malformed pending", pending: []byte("{"), wantError: true, pendingAfter: true},
		{name: "preserve malformed manifest ambiguity", manifest: []byte("{"), pending: encodedDocument(t, "pending"), wantError: true, pendingAfter: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			state := filepath.Join(root, ".soku")
			if err := os.Mkdir(state, 0o700); err != nil {
				t.Fatal(err)
			}
			if test.manifest != nil {
				if err := os.WriteFile(filepath.Join(state, "manifest.json"), test.manifest, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if test.pending != nil {
				if err := os.WriteFile(filepath.Join(state, "manifest.json.pending"), test.pending, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			document, err := NewStore(root).Recover()
			if (err != nil) != test.wantError {
				t.Fatalf("Recover() error = %v, wantError %v", err, test.wantError)
			}
			if !test.wantError && document.SokuVersion != test.wantVersion {
				t.Fatalf("version = %q, want %q", document.SokuVersion, test.wantVersion)
			}
			_, pendingErr := os.Stat(filepath.Join(state, "manifest.json.pending"))
			if test.pendingAfter != (pendingErr == nil) {
				t.Fatalf("pending error = %v, pendingAfter=%v", pendingErr, test.pendingAfter)
			}
		})
	}
}

func TestLoadReportsPendingWithoutMutation(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, ".soku")
	if err := os.Mkdir(state, 0o700); err != nil {
		t.Fatal(err)
	}
	pending := encodedDocument(t, "pending")
	path := filepath.Join(state, "manifest.json.pending")
	if err := os.WriteFile(path, pending, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewStore(root).Load(); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatalf("Load() error = %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, pending) {
		t.Fatalf("pending was changed: %v", err)
	}
}

func TestLoadTreatsInvalidPendingCombinationsAsValidation(t *testing.T) {
	for _, test := range []struct {
		name     string
		manifest []byte
		pending  []byte
	}{
		{name: "unsupported pending only", pending: []byte(`{"schema_version":2}`)},
		{name: "unsupported durable beside pending", manifest: []byte(`{"schema_version":2}`), pending: encodedDocument(t, "pending")},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if test.manifest != nil {
				writeStoreState(t, root, "manifest.json", test.manifest)
			}
			writeStoreState(t, root, "manifest.json.pending", test.pending)
			if _, err := NewStore(root).Load(); err == nil {
				t.Fatal("invalid pending combination was accepted")
			} else {
				var unsupported *UnsupportedSchemaError
				if errors.As(err, &unsupported) {
					t.Fatalf("pending ambiguity was treated as compatibility: %v", err)
				}
			}
		})
	}
}

func encodedDocument(t *testing.T, version string) []byte {
	t.Helper()
	document := validDocument()
	document.SokuVersion = version
	data, err := MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeStoreState(t *testing.T, root, name string, data []byte) {
	t.Helper()
	state := filepath.Join(root, ".soku")
	if err := os.MkdirAll(state, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, name), data, 0o600); err != nil {
		t.Fatal(err)
	}
}
