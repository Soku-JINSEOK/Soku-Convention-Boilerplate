package initcmd

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

const gitignoreBegin = "# >>> soku managed patterns >>>"
const gitignoreEnd = "# <<< soku managed patterns <<<"

func mergeGitignore(existing, desired []byte) ([]byte, error) {
	text := normalizeText(existing)
	if strings.Count(text, gitignoreBegin) != strings.Count(text, gitignoreEnd) || strings.Count(text, gitignoreBegin) > 1 {
		return nil, fail(4, "merge.conflict", ".gitignore has ambiguous Soku block markers")
	}
	present := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			present[line] = true
		}
	}
	var missing []string
	scanner = bufio.NewScanner(bytes.NewReader(desired))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") && !present[line] {
			missing = append(missing, line)
			present[line] = true
		}
	}
	if len(missing) == 0 {
		return []byte(text), nil
	}
	block := gitignoreBegin + "\n" + strings.Join(missing, "\n") + "\n" + gitignoreEnd
	if strings.Contains(text, gitignoreBegin) {
		start := strings.Index(text, gitignoreBegin)
		end := strings.Index(text, gitignoreEnd) + len(gitignoreEnd)
		current := strings.TrimSuffix(text[start:end], "\n")
		current = strings.TrimSuffix(current, gitignoreEnd)
		current = strings.TrimRight(current, "\n")
		block = current + "\n" + strings.Join(missing, "\n") + "\n" + gitignoreEnd
		return []byte(text[:start] + block + text[end:]), nil
	}
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	if text != "" {
		text += "\n"
	}
	return []byte(text + block + "\n"), nil
}

type editorSection struct {
	name  string
	lines []string
	keys  map[string]bool
}

func mergeEditorconfig(existing, desired []byte) ([]byte, error) {
	current, err := parseEditorconfig(normalizeText(existing))
	if err != nil {
		return nil, err
	}
	wanted, err := parseEditorconfig(normalizeText(desired))
	if err != nil {
		return nil, fail(5, "catalog.incompatible", "source .editorconfig is invalid: %v", err)
	}
	byName := map[string]*editorSection{}
	for index := range current {
		byName[strings.ToLower(current[index].name)] = &current[index]
	}
	missingBySection := map[string][]string{}
	var absentSections []string
	for _, section := range wanted {
		target := byName[strings.ToLower(section.name)]
		if target == nil {
			absentSections = append(absentSections, strings.Join(section.lines, "\n"))
			continue
		}
		var keys []string
		for _, line := range section.lines {
			key, _, ok := strings.Cut(strings.TrimSpace(line), "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			if !target.keys[strings.ToLower(key)] {
				keys = append(keys, line)
				target.keys[strings.ToLower(key)] = true
			}
		}
		if len(keys) > 0 {
			missingBySection[strings.ToLower(section.name)] = keys
		}
	}
	if len(missingBySection) == 0 && len(absentSections) == 0 {
		return []byte(normalizeText(existing)), nil
	}
	lines := strings.Split(strings.TrimSuffix(normalizeText(existing), "\n"), "\n")
	var output []string
	section := "<root>"
	flush := func() {
		keys := missingBySection[strings.ToLower(section)]
		if len(keys) == 0 {
			return
		}
		if len(output) > 0 && output[len(output)-1] != "" {
			output = append(output, "")
		}
		output = append(output, fmt.Sprintf("# Soku additions for %s", section))
		output = append(output, keys...)
		delete(missingBySection, strings.ToLower(section))
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			flush()
			section = trimmed
		}
		output = append(output, line)
	}
	flush()
	for _, block := range absentSections {
		if len(output) > 0 && output[len(output)-1] != "" {
			output = append(output, "")
		}
		output = append(output, strings.Split(block, "\n")...)
	}
	return []byte(strings.Join(output, "\n") + "\n"), nil
}

func parseEditorconfig(text string) ([]editorSection, error) {
	sections := []editorSection{{name: "<root>", keys: map[string]bool{}}}
	seenSections := map[string]bool{"<root>": true}
	current := &sections[0]
	for number, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") || len(line) < 3 {
				return nil, fail(4, "merge.conflict", ".editorconfig line %d has an invalid section", number+1)
			}
			name := line
			lower := strings.ToLower(name)
			if seenSections[lower] {
				return nil, fail(4, "merge.conflict", ".editorconfig repeats section %s", name)
			}
			seenSections[lower] = true
			sections = append(sections, editorSection{name: name, lines: []string{raw}, keys: map[string]bool{}})
			current = &sections[len(sections)-1]
			continue
		}
		key, _, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fail(4, "merge.conflict", ".editorconfig line %d is ambiguous", number+1)
		}
		lower := strings.ToLower(strings.TrimSpace(key))
		if current.keys[lower] {
			return nil, fail(4, "merge.conflict", ".editorconfig repeats key %s in %s", key, current.name)
		}
		current.keys[lower] = true
		current.lines = append(current.lines, raw)
	}
	return sections, nil
}

func normalizeText(data []byte) string {
	return strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n")
}
