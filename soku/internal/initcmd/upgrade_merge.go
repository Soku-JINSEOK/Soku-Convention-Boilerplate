package initcmd

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"
)

func mergeGitignoreThreeWay(base, current, target []byte) ([]byte, error) {
	baseSet := gitignoreSet(base)
	targetSet := gitignoreSet(target)
	removed := map[string]bool{}
	for line := range baseSet {
		if !targetSet[line] {
			removed[line] = true
		}
	}
	present := map[string]bool{}
	var output []string
	for _, raw := range strings.Split(strings.TrimSuffix(normalizeText(current), "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line != "" && !strings.HasPrefix(line, "#") {
			if removed[line] {
				continue
			}
			present[line] = true
		}
		output = append(output, raw)
	}
	var additions []string
	for line := range targetSet {
		if !baseSet[line] && !present[line] {
			additions = append(additions, line)
		}
	}
	sort.Strings(additions)
	if len(additions) > 0 && len(output) > 0 && output[len(output)-1] != "" {
		output = append(output, "")
	}
	output = append(output, additions...)
	return []byte(strings.Join(output, "\n") + "\n"), nil
}

func gitignoreSet(data []byte) map[string]bool {
	result := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			result[line] = true
		}
	}
	return result
}

type editorValues map[string]map[string]string

func mergeEditorconfigThreeWay(base, current, target []byte) ([]byte, error) {
	baseValues, err := parseEditorValues(base)
	if err != nil {
		return nil, fail(5, "catalog.incompatible", "base .editorconfig is invalid: %v", err)
	}
	currentValues, err := parseEditorValues(current)
	if err != nil {
		return nil, fail(4, "merge.conflict", "current .editorconfig is invalid: %v", err)
	}
	targetValues, err := parseEditorValues(target)
	if err != nil {
		return nil, fail(5, "catalog.incompatible", "target .editorconfig is invalid: %v", err)
	}
	result := cloneEditorValues(currentValues)
	sections := map[string]bool{}
	for section := range baseValues {
		sections[section] = true
	}
	for section := range targetValues {
		sections[section] = true
	}
	for section := range sections {
		keys := map[string]bool{}
		for key := range baseValues[section] {
			keys[key] = true
		}
		for key := range targetValues[section] {
			keys[key] = true
		}
		for key := range keys {
			baseValue, inBase := baseValues[section][key]
			targetValue, inTarget := targetValues[section][key]
			currentValue, inCurrent := currentValues[section][key]
			switch {
			case !inBase && inTarget:
				if inCurrent && currentValue != targetValue {
					return nil, fail(4, "merge.conflict", ".editorconfig %s %s was added differently by project and target", section, key)
				}
				setEditorValue(result, section, key, targetValue)
			case inBase && !inTarget:
				if inCurrent && currentValue != baseValue {
					return nil, fail(4, "merge.conflict", ".editorconfig %s %s was changed locally and removed by target", section, key)
				}
				deleteEditorValue(result, section, key)
			case inBase && inTarget && baseValue != targetValue:
				if inCurrent && currentValue != baseValue && currentValue != targetValue {
					return nil, fail(4, "merge.conflict", ".editorconfig %s %s changed differently in project and target", section, key)
				}
				setEditorValue(result, section, key, targetValue)
			}
		}
	}
	return marshalEditorValues(result), nil
}

func parseEditorValues(data []byte) (editorValues, error) {
	values := editorValues{"<root>": {}}
	section := "<root>"
	for number, raw := range strings.Split(normalizeText(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") || len(line) < 3 {
				return nil, fmt.Errorf("line %d has an invalid section", number+1)
			}
			section = strings.ToLower(line)
			if _, exists := values[section]; exists {
				return nil, fmt.Errorf("section %s is repeated", line)
			}
			values[section] = map[string]string{}
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.ToLower(strings.TrimSpace(key))
		if !ok || key == "" {
			return nil, fmt.Errorf("line %d is ambiguous", number+1)
		}
		if _, exists := values[section][key]; exists {
			return nil, fmt.Errorf("key %s is repeated in %s", key, section)
		}
		values[section][key] = strings.TrimSpace(value)
	}
	return values, nil
}

func cloneEditorValues(values editorValues) editorValues {
	result := editorValues{}
	for section, entries := range values {
		result[section] = map[string]string{}
		for key, value := range entries {
			result[section][key] = value
		}
	}
	return result
}

func setEditorValue(values editorValues, section, key, value string) {
	if values[section] == nil {
		values[section] = map[string]string{}
	}
	values[section][key] = value
}

func deleteEditorValue(values editorValues, section, key string) {
	delete(values[section], key)
	if section != "<root>" && len(values[section]) == 0 {
		delete(values, section)
	}
}

func marshalEditorValues(values editorValues) []byte {
	var output []string
	appendKeys := func(section string) {
		keys := make([]string, 0, len(values[section]))
		for key := range values[section] {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			output = append(output, key+" = "+values[section][key])
		}
	}
	appendKeys("<root>")
	var sections []string
	for section := range values {
		if section != "<root>" {
			sections = append(sections, section)
		}
	}
	sort.Strings(sections)
	for _, section := range sections {
		if len(output) > 0 {
			output = append(output, "")
		}
		output = append(output, section)
		appendKeys(section)
	}
	return []byte(strings.Join(output, "\n") + "\n")
}
