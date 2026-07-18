package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGoInstallToTemporaryGOBIN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go install integration test in short mode")
	}
	gobin := t.TempDir()
	command := exec.Command("go", "install", ".")
	command.Env = append(os.Environ(), "GOBIN="+gobin)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("go install failed: %v\n%s", err, output)
	}

	binary := filepath.Join(gobin, "soku")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	version := exec.Command(binary, "--version")
	output, err := version.CombinedOutput()
	if err != nil {
		t.Fatalf("installed binary failed: %v\n%s", err, output)
	}
	if !strings.HasPrefix(string(output), "soku ") {
		t.Fatalf("unexpected version output: %q", output)
	}
}
