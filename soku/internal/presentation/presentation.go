// Package presentation renders terminal-friendly human output without changing
// the stable plain-text representation used by pipes and snapshots.
package presentation

import (
	"regexp"
	"strings"
)

// ColorMode controls whether ANSI styling is applied to human output.
type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

var statusPattern = regexp.MustCompile(`(?m)(^Soku [^:\n]+: )([^\n]+)$`)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	blue   = "\x1b[34m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
)

// Enabled resolves the public color contract for human output.
func Enabled(mode ColorMode, terminal bool, term string, noColor bool) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		return terminal && term != "dumb" && !noColor
	}
}

// Style applies restrained ANSI styling while preserving all visible words and
// ASCII markers. StripANSI(Style(text)) is always equal to text.
func Style(text string) string {
	styled := statusPattern.ReplaceAllStringFunc(text, func(line string) string {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			return line
		}
		return bold + blue + parts[0] + reset + ": " + stateColor(parts[1]) + parts[1] + reset
	})
	lines := strings.SplitAfter(styled, "\n")
	for index, line := range lines {
		plain := strings.TrimSuffix(line, "\n")
		suffix := strings.TrimPrefix(line, plain)
		if strings.HasPrefix(plain, "- ") {
			lines[index] = yellow + "-" + reset + plain[1:] + suffix
			continue
		}
		if strings.HasPrefix(plain, "Next: ") {
			lines[index] = bold + yellow + "Next:" + reset + plain[len("Next:"):] + suffix
			continue
		}
		if label, rest, ok := strings.Cut(plain, ":"); ok && !strings.Contains(label, "\x1b[") {
			lines[index] = bold + label + ":" + reset + rest + suffix
		}
	}
	return strings.Join(lines, "")
}

func stateColor(state string) string {
	fields := strings.Fields(state)
	if len(fields) == 0 {
		return blue
	}
	switch strings.ToLower(fields[0]) {
	case "clean", "ready", "applied", "initialized", "installed", "complete", "completed", "ok", "valid":
		return green
	case "drifted", "uninitialized", "recovery-required", "pending", "planned", "dry-run", "changes":
		return yellow
	case "incompatible", "failed", "error", "invalid", "unreadable":
		return red
	default:
		return blue
	}
}
