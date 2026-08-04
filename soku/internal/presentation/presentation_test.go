package presentation

import (
	"regexp"
	"testing"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		mode               ColorMode
		tty, noColor, want bool
		term               string
	}{
		{ColorAuto, true, false, true, "xterm-256color"},
		{ColorAuto, false, false, false, "xterm-256color"},
		{ColorAuto, true, true, false, "xterm-256color"},
		{ColorAuto, true, false, false, "dumb"},
		{ColorAlways, false, true, true, "dumb"},
		{ColorNever, true, false, false, "xterm-256color"},
	}
	for _, test := range tests {
		if got := Enabled(test.mode, test.tty, test.term, test.noColor); got != test.want {
			t.Errorf("Enabled(%q, %t, %q, %t) = %t, want %t", test.mode, test.tty, test.term, test.noColor, got, test.want)
		}
	}
}

func TestStylePreservesPlainText(t *testing.T) {
	plain := "Soku status: drifted\nManifest: .soku/manifest.json\n- file: changed\nNext: Review it.\n"
	styled := Style(plain)
	if styled == plain {
		t.Fatal("Style did not add ANSI sequences")
	}
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	if got := ansi.ReplaceAllString(styled, ""); got != plain {
		t.Fatalf("stripped output = %q, want %q", got, plain)
	}
}
