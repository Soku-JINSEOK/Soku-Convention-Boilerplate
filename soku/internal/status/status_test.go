package status

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

const (
	statusCommit = "0123456789abcdef0123456789abcdef01234567"
	statusHash   = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestInspectUninitializedAndMalformed(t *testing.T) {
	result, err := Inspect(t.TempDir())
	if err != nil || result.Code != 3 || result.Report.State != "uninitialized" {
		t.Fatalf("uninitialized result = %#v, %v", result, err)
	}

	root := t.TempDir()
	writeStateFile(t, root, "manifest.json", []byte("{"))
	if _, err := Inspect(root); err == nil {
		t.Fatal("malformed manifest was reported as a diagnostic success")
	}
}

func TestInspectFileStatesAndSorting(t *testing.T) {
	root := t.TempDir()
	writeProjectFile(t, root, "changed.txt", []byte("changed"), 0o600)
	writeProjectFile(t, root, "clean.txt", []byte("clean\r\n"), 0o600)
	writeProjectFile(t, root, "obsolete.txt", []byte("old"), 0o600)
	writeProjectFile(t, root, "project.txt", []byte("owned"), 0o600)
	if err := os.Mkdir(filepath.Join(root, "wrong-type"), 0o700); err != nil {
		t.Fatal(err)
	}
	cleanHash, _ := manifest.HashContent([]byte("clean\n"), "text")
	document := statusDocument()
	document.Files = []manifest.File{
		statusFile("changed.txt", statusHash, "current"),
		statusFile("clean.txt", cleanHash, "current"),
		statusFile("missing.txt", statusHash, "current"),
		statusFile("obsolete.txt", statusHash, "obsolete"),
		{Path: "project.txt", Owner: "project", Class: "project-owned", LifecycleState: "unmanaged-expected"},
		statusFile("wrong-type", statusHash, "current"),
	}
	writeManifest(t, root, document)

	result, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Code != 3 || result.Report.State != "drifted" {
		t.Fatalf("result = %#v", result)
	}
	want := []string{"changed", "clean", "missing", "obsolete", "unmanaged-expected", "type-mismatch"}
	for index, state := range want {
		if result.Report.Files[index].State != state {
			t.Errorf("file[%d] state = %q, want %q", index, result.Report.Files[index].State, state)
		}
	}
	if result.Report.Counts.Clean != 1 || result.Report.Counts.Missing != 1 ||
		result.Report.Counts.Changed != 1 || result.Report.Counts.Obsolete != 1 ||
		result.Report.Counts.UnmanagedExpected != 1 || result.Report.Counts.TypeMismatch != 1 {
		t.Fatalf("counts = %#v", result.Report.Counts)
	}
}

func TestInspectCleanPendingDriftAndCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		state     string
		api       string
		wantCode  int
		wantState string
	}{
		{name: "clean", state: "connected", api: "1", wantState: "clean"},
		{name: "pending", state: "pending", api: "1", wantCode: 3, wantState: "drifted"},
		{name: "drifted", state: "drifted", api: "1", wantCode: 3, wantState: "drifted"},
		{name: "recorded incompatible", state: "incompatible", api: "1", wantCode: 5, wantState: "incompatible"},
		{name: "incompatible provider", state: "connected", api: "2", wantCode: 5, wantState: "incompatible"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			document := statusDocument()
			document.Integrations = []manifest.Integration{{
				ID: "provider", Source: "https://github.com/example/provider", Ref: statusCommit,
				ProviderAPIVersion: test.api, ProviderSchemaVersion: "1", ConfigurationHash: statusHash,
				LifecycleState: test.state, ManagedFiles: []string{},
			}}
			writeManifest(t, root, document)
			result, err := Inspect(root)
			if err != nil || result.Code != test.wantCode || result.Report.State != test.wantState {
				t.Fatalf("Inspect() = %#v, %v", result, err)
			}
		})
	}
}

func TestInspectUnsupportedSchemaIsDiagnostic(t *testing.T) {
	root := t.TempDir()
	data := []byte(`{"schema_version":2}`)
	writeStateFile(t, root, "manifest.json", data)
	result, err := Inspect(root)
	if err != nil || result.Code != 5 || result.Report.State != "incompatible" || result.Report.Manifest.SchemaVersion != 2 {
		t.Fatalf("Inspect() = %#v, %v", result, err)
	}
}

func TestInspectRecoveryRequiredDoesNotChangePending(t *testing.T) {
	root := t.TempDir()
	document := statusDocument()
	data, err := manifest.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	writeStateFile(t, root, "manifest.json.pending", data)
	result, err := Inspect(root)
	if err != nil || result.Code != 3 || result.Report.State != "recovery-required" {
		t.Fatalf("Inspect() = %#v, %v", result, err)
	}
	after, err := os.ReadFile(filepath.Join(root, ".soku", "manifest.json.pending"))
	if err != nil || !bytes.Equal(after, data) {
		t.Fatalf("pending changed: %v", err)
	}
}

func TestInspectSymlinkEscapeAndUnreadable(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape.txt")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink is unavailable: %v", err)
		}
		t.Fatal(err)
	}
	document := statusDocument()
	document.Files = []manifest.File{statusFile("escape.txt", statusHash, "current")}
	writeManifest(t, root, document)
	result, err := Inspect(root)
	if err != nil || result.Report.Files[0].State != "symlink-escape" {
		t.Fatalf("Inspect() = %#v, %v", result, err)
	}

	unreadableRoot := t.TempDir()
	writeProjectFile(t, unreadableRoot, "private.txt", []byte("private"), 0o000)
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(unreadableRoot, "private.txt"), 0o600) })
	unreadableDocument := statusDocument()
	unreadableDocument.Files = []manifest.File{statusFile("private.txt", statusHash, "current")}
	writeManifest(t, unreadableRoot, unreadableDocument)
	unreadable, err := Inspect(unreadableRoot)
	if err == nil {
		if os.Geteuid() == 0 {
			t.Skip("root can read mode-000 test files")
		}
		t.Fatalf("Inspect() = %#v, want validation error", unreadable)
	}
	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("error = %T %v, want ValidationError", err, err)
	}
}

func statusDocument() manifest.Document {
	return manifest.Document{
		SchemaVersion: 1, SokuVersion: "v0.2.0",
		Boilerplate: manifest.Boilerplate{Source: "https://github.com/example/boilerplate", Release: "v1.0.0", ResolvedCommit: statusCommit},
		Selection:   manifest.Selection{Profile: "team", Stacks: []string{}, ConfigurationHash: statusHash},
		Files:       []manifest.File{}, Integrations: []manifest.Integration{},
	}
}

func statusFile(path, hash, state string) manifest.File {
	return manifest.File{Path: path, Owner: "core", Class: "core-managed", ContentMode: "text", BaselineSHA256: hash, LifecycleState: state}
}

func writeManifest(t *testing.T, root string, document manifest.Document) {
	t.Helper()
	if err := manifest.NewStore(root).Write(document); err != nil {
		t.Fatal(err)
	}
}

func writeProjectFile(t *testing.T, root, name string, content []byte, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		t.Fatal(err)
	}
}

func writeStateFile(t *testing.T, root, name string, content []byte) {
	t.Helper()
	state := filepath.Join(root, ".soku")
	if err := os.MkdirAll(state, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state, name), content, 0o600); err != nil {
		t.Fatal(err)
	}
}
