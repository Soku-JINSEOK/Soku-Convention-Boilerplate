package cli

import (
	"os"
	"testing"
)

func TestOSRuntimeRejectsNonTerminalCharacterDeviceAndPipe(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = devNull.Close() })

	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pipeReader.Close() })
	t.Cleanup(func() { _ = pipeWriter.Close() })

	for name, input := range map[string]*os.File{
		"null device": devNull,
		"pipe":        pipeReader,
	} {
		t.Run(name, func(t *testing.T) {
			runtime := osRuntime{stdin: input}
			if runtime.IsTerminal() {
				t.Fatalf("%s was incorrectly detected as a terminal", name)
			}
		})
	}
}
