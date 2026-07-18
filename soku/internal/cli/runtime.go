package cli

import (
	"io"
	"io/fs"
	"os"

	"golang.org/x/term"
)

// Runtime isolates environment observations so command behavior is testable.
type Runtime interface {
	Stat(name string) (fs.FileInfo, error)
	Open(name string) (io.ReadCloser, error)
	IsTerminal() bool
}

type osRuntime struct {
	stdin *os.File
}

func (r osRuntime) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (r osRuntime) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (r osRuntime) IsTerminal() bool {
	if r.stdin == nil {
		return false
	}
	return term.IsTerminal(int(r.stdin.Fd()))
}
