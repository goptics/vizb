package parser

import (
	"fmt"
	"strings"
)

// parseCurlyLabel reads an optional {label} suffix starting at open (which must be '{').
func parseCurlyLabel(pattern string, open int) (label string, next int, err error) {
	if open >= len(pattern) || pattern[open] != '{' {
		return "", open, nil
	}
	rest := pattern[open+1:]
	close := strings.Index(rest, "}")
	if close == -1 {
		return "", open, fmt.Errorf("unclosed '{' in --group-pattern %q", pattern)
	}
	label = rest[:close]
	if label == "" {
		return "", open, fmt.Errorf("empty axis label {} in --group-pattern %q", pattern)
	}
	return label, open + 1 + close + 1, nil
}

// parsePatternLabels strips {label} suffixes after dimension tokens for splitting and
// returns the label map keyed by expanded dimension (name, xAxis, yAxis, zAxis).
func parsePatternLabels(pattern string) (string, map[string]string, error) {
	if !strings.Contains(pattern, "{") {
		return pattern, nil, nil
	}

	labels := make(map[string]string)
	var out strings.Builder
	i := 0

	for i < len(pattern) {
		matched := false
		for _, tok := range dimensionTokens {
			if !strings.HasPrefix(pattern[i:], tok.raw) || !isPatternTokenBoundary(pattern, i, len(tok.raw)) {
				continue
			}
			out.WriteString(tok.raw)
			i += len(tok.raw)
			if i < len(pattern) && pattern[i] == '{' {
				lbl, next, err := parseCurlyLabel(pattern, i)
				if err != nil {
					return "", nil, err
				}
				labels[tok.expanded] = lbl
				i = next
			}
			matched = true
			break
		}
		if !matched {
			out.WriteByte(pattern[i])
			i++
		}
	}

	return out.String(), labels, nil
}

func slotLabel(slot PatternSlot, dim string, colLabel string) string {
	if lbl := slot.AxisLabels[dim]; lbl != "" {
		return lbl
	}
	return colLabel
}
