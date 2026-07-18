package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testRuntime struct {
	terminal bool
	stat     func(string) (fs.FileInfo, error)
	open     func(string) (io.ReadCloser, error)
}

func (r testRuntime) Stat(name string) (fs.FileInfo, error) {
	if r.stat != nil {
		return r.stat(name)
	}
	return os.Stat(name)
}

func (r testRuntime) Open(name string) (io.ReadCloser, error) {
	if r.open != nil {
		return r.open(name)
	}
	return os.Open(name)
}

func (r testRuntime) IsTerminal() bool {
	return r.terminal
}

type runResult struct {
	code   int
	stdout string
	stderr string
}

func execute(args []string, runtime Runtime, handlers Handlers) runResult {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWith(args, &stdout, &stderr, dependencies{
		runtime:  runtime,
		handlers: handlers,
		metadata: BuildMetadata{Version: "v1.2.3", Commit: "abc123", BuiltAt: "2026-07-18T00:00:00Z"},
	})
	return runResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func successHandlers(handler Handler) Handlers {
	return Handlers{Init: handler, Status: handler, Diff: handler, Upgrade: handler}
}

func TestHelpAndVersionOutput(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStdout string
		json       bool
	}{
		{name: "human help", args: []string{"--help"}, wantStdout: "Usage:\n  soku <command> [flags]"},
		{name: "human subcommand help", args: []string{"init", "--help"}, wantStdout: "Usage:\n  soku init [flags]"},
		{name: "human version", args: []string{"--version"}, wantStdout: "soku v1.2.3\n"},
		{name: "json help", args: []string{"--help", "--json"}, json: true},
		{name: "json version", args: []string{"--json", "--version"}, json: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := execute(test.args, testRuntime{}, defaultHandlers())
			if result.code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr=%q", result.code, result.stderr)
			}
			if result.stderr != "" {
				t.Fatalf("stderr = %q, want empty", result.stderr)
			}
			if test.json {
				assertSingleJSONEnvelope(t, result.stdout)
				var decoded map[string]any
				if err := json.Unmarshal([]byte(result.stdout), &decoded); err != nil {
					t.Fatal(err)
				}
				data := decoded["data"].(map[string]any)
				if strings.Contains(test.name, "version") {
					if data["version"] != "v1.2.3" || data["commit"] != "abc123" || data["built_at"] != "2026-07-18T00:00:00Z" {
						t.Fatalf("unexpected version data: %#v", data)
					}
				} else if _, ok := data["help"]; !ok {
					t.Fatalf("help data missing: %#v", data)
				}
				return
			}
			if !strings.Contains(result.stdout, test.wantStdout) {
				t.Fatalf("stdout = %q, want substring %q", result.stdout, test.wantStdout)
			}
		})
	}
}

func TestExitCodeContract(t *testing.T) {
	codes := []ExitCode{
		ExitSuccess,
		ExitInternalError,
		ExitValidationFailure,
		ExitChangesFound,
		ExitSafetyRefusal,
		ExitCompatibilityFailure,
		ExitSourceFailure,
		ExitApplyRolledBack,
		ExitRollbackFailure,
	}
	for want, code := range codes {
		if int(code) != want {
			t.Fatalf("exit code at index %d = %d", want, code)
		}
	}
}

func TestPublicCommandSurface(t *testing.T) {
	result := execute([]string{"--help"}, testRuntime{}, defaultHandlers())
	for _, command := range []string{"init", "status", "diff", "upgrade"} {
		if !strings.Contains(result.stdout, "  "+command+" ") {
			t.Errorf("help does not list %q", command)
		}
	}
	if strings.Contains(result.stdout, "completion") || strings.Contains(result.stdout, "\n  help ") {
		t.Fatalf("help exposes a non-public command:\n%s", result.stdout)
	}
	for _, command := range []string{"completion", "help", "_help_disabled", "__complete", "__completeNoDesc"} {
		result := execute([]string{command}, testRuntime{}, defaultHandlers())
		if result.code != 2 {
			t.Errorf("%s exit code = %d, want 2", command, result.code)
		}
	}
	if result := execute([]string{"-v"}, testRuntime{}, defaultHandlers()); result.code != 2 {
		t.Errorf("-v exit code = %d, want 2", result.code)
	}
}

func TestDefaultLifecycleHandlersAreUnavailable(t *testing.T) {
	for _, test := range []struct {
		command string
		args    []string
	}{
		{command: "init", args: []string{"init", "--dry-run"}},
		{command: "status", args: []string{"status"}},
		{command: "diff", args: []string{"diff"}},
		{command: "upgrade", args: []string{"upgrade", "--dry-run"}},
	} {
		t.Run(test.command, func(t *testing.T) {
			result := execute(test.args, testRuntime{}, defaultHandlers())
			if result.code != 5 {
				t.Fatalf("exit code = %d, want 5; stderr=%q", result.code, result.stderr)
			}
			if !strings.Contains(result.stderr, "not available") || result.stdout != "" {
				t.Fatalf("stdout=%q stderr=%q", result.stdout, result.stderr)
			}
		})
	}
}

func TestInvocationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		command string
	}{
		{name: "unknown command", args: []string{"migrate"}, command: "migrate"},
		{name: "unknown flag", args: []string{"status", "--wat"}, command: "status"},
		{name: "extra argument", args: []string{"diff", "extra"}, command: "diff"},
		{name: "invalid safety pair", args: []string{"init", "--dry-run", "--yes"}, command: "init"},
		{name: "non tty mutation", args: []string{"upgrade"}, command: "upgrade"},
		{name: "explicit non interactive", args: []string{"init", "--non-interactive"}, command: "init"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := execute(test.args, testRuntime{terminal: test.name == "explicit non interactive"}, defaultHandlers())
			if result.code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr=%q", result.code, result.stderr)
			}
			if result.stdout != "" || !strings.HasPrefix(result.stderr, "Error: ") {
				t.Fatalf("stdout=%q stderr=%q", result.stdout, result.stderr)
			}
		})
	}
}

func TestJSONIsDetectedBeforeParsing(t *testing.T) {
	for _, test := range []struct {
		name        string
		args        []string
		wantCommand string
	}{
		{name: "after unknown flag", args: []string{"status", "--invalid", "--json"}, wantCommand: "status"},
		{name: "after unknown command", args: []string{"migrate", "--invalid", "--json"}, wantCommand: "migrate"},
		{name: "after extra argument", args: []string{"diff", "extra", "--json"}, wantCommand: "diff"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := execute(test.args, testRuntime{}, defaultHandlers())
			if result.code != 2 || result.stderr != "" {
				t.Fatalf("code=%d stdout=%q stderr=%q", result.code, result.stdout, result.stderr)
			}
			assertSingleJSONEnvelope(t, result.stdout)
			if !strings.Contains(result.stdout, `"command":"`+test.wantCommand+`"`) {
				t.Fatalf("wrong command in %s", result.stdout)
			}
		})
	}
}

func TestMutationRuntimeRulesAndRequest(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		terminal     bool
		wantCode     int
		wantDryRun   bool
		wantYes      bool
		wantInteract bool
		wantNonInt   bool
	}{
		{name: "tty confirmation path", args: []string{"init"}, terminal: true, wantInteract: true},
		{name: "non tty yes", args: []string{"init", "--yes"}, wantYes: true, wantNonInt: true},
		{name: "non tty dry run", args: []string{"upgrade", "--dry-run"}, wantDryRun: true, wantNonInt: true},
		{name: "explicit override", args: []string{"upgrade", "--non-interactive", "--yes"}, terminal: true, wantYes: true, wantNonInt: true},
		{name: "missing decision", args: []string{"init"}, wantCode: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got Request
			handler := HandlerFunc(func(_ context.Context, request Request) error {
				got = request
				return nil
			})
			result := execute(test.args, testRuntime{terminal: test.terminal}, successHandlers(handler))
			if result.code != test.wantCode {
				t.Fatalf("exit code = %d, want %d; stderr=%q", result.code, test.wantCode, result.stderr)
			}
			if test.wantCode != 0 {
				return
			}
			if got.DryRun != test.wantDryRun || got.Yes != test.wantYes || got.Interactive != test.wantInteract || got.NonInteractive != test.wantNonInt {
				t.Fatalf("request = %#v", got)
			}
		})
	}
}

func TestQuietIsPassedWithoutSuppressingErrors(t *testing.T) {
	var got Request
	handler := HandlerFunc(func(_ context.Context, request Request) error {
		got = request
		return unavailableError(request.Command)
	})
	result := execute([]string{"status", "--quiet"}, testRuntime{}, successHandlers(handler))
	if result.code != 5 || !got.Quiet || result.stderr == "" {
		t.Fatalf("request=%#v result=%#v", got, result)
	}
}

func TestConfigValidation(t *testing.T) {
	tempDir := t.TempDir()
	validPath := filepath.Join(tempDir, "soku.yaml")
	if err := os.WriteFile(validPath, []byte("version: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	handler := HandlerFunc(func(context.Context, Request) error { return nil })
	tests := []struct {
		name     string
		path     string
		runtime  Runtime
		wantCode int
	}{
		{name: "missing", path: filepath.Join(tempDir, "missing.yaml"), runtime: testRuntime{}, wantCode: 2},
		{name: "directory", path: tempDir, runtime: testRuntime{}, wantCode: 2},
		{name: "valid", path: validPath, runtime: testRuntime{}},
		{
			name: "unreadable",
			path: validPath,
			runtime: testRuntime{open: func(string) (io.ReadCloser, error) {
				return nil, fs.ErrPermission
			}},
			wantCode: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := execute([]string{"status", "--config", test.path}, test.runtime, successHandlers(handler))
			if result.code != test.wantCode {
				t.Fatalf("exit code=%d want=%d stderr=%q", result.code, test.wantCode, result.stderr)
			}
		})
	}

	for _, args := range [][]string{
		{"--config", filepath.Join(tempDir, "missing.yaml"), "--help"},
		{"--config", filepath.Join(tempDir, "missing.yaml"), "--version"},
	} {
		if result := execute(args, testRuntime{}, defaultHandlers()); result.code != 0 {
			t.Fatalf("help/version blocked by config: args=%v result=%#v", args, result)
		}
	}
}

func TestHandlerFailuresMapToInternalError(t *testing.T) {
	tests := []struct {
		name    string
		handler Handler
	}{
		{name: "error", handler: HandlerFunc(func(context.Context, Request) error { return errors.New("secret detail") })},
		{name: "panic", handler: HandlerFunc(func(context.Context, Request) error { panic("secret detail") })},
		{name: "nil", handler: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handlers := successHandlers(test.handler)
			result := execute([]string{"status", "--json"}, testRuntime{}, handlers)
			if result.code != 1 || result.stderr != "" {
				t.Fatalf("result=%#v", result)
			}
			if result.stdout != "{\"ok\":false,\"command\":\"status\",\"error\":{\"code\":\"internal.error\",\"message\":\"internal command failure\"},\"data\":null}\n" {
				t.Fatalf("unexpected JSON: %s", result.stdout)
			}
		})
	}
}

func TestDeterministicJSONSuccessAndStreamSeparation(t *testing.T) {
	handler := HandlerFunc(func(context.Context, Request) error { return nil })
	first := execute([]string{"status", "--json"}, testRuntime{}, successHandlers(handler))
	second := execute([]string{"--json", "status"}, testRuntime{}, successHandlers(handler))
	want := "{\"ok\":true,\"command\":\"status\",\"error\":null,\"data\":{}}\n"
	if first.stdout != want || second.stdout != want || first.stderr != "" || second.stderr != "" {
		t.Fatalf("first=%#v second=%#v", first, second)
	}
}

func assertSingleJSONEnvelope(t *testing.T, value string) {
	t.Helper()
	if strings.Count(value, "\n") != 1 || !json.Valid([]byte(value)) {
		t.Fatalf("not one JSON envelope: %q", value)
	}
}
