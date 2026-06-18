package parser

import (
	"fmt"
	"strings"

	"github.com/goptics/vizb/shared"
)

// PatternSlot is one top-level --group-pattern slot aligned with a --group column.
type PatternSlot struct {
	ValueSplit   bool
	InnerPattern string
	Dimension    string
	AxisLabels   map[string]string // optional {label} overrides per dimension in this slot
}

// TabularPattern describes CSV/JSON column-slot grouping (--group-pattern with optional [...]).
type TabularPattern struct {
	Slots         []PatternSlot
	TopSeparators []string
}

// ParseTabularPattern parses a tabular --group-pattern into column-aligned slots.
// Bracket slots [...] split a column's cell value; plain tokens map the whole cell.
func ParseTabularPattern(pattern string) (TabularPattern, error) {
	if pattern == "" {
		return TabularPattern{}, errPatternEmpty
	}
	if strings.Contains(pattern, "[") {
		return parseBracketedTabularPattern(pattern)
	}
	return parseFlatTabularPattern(pattern)
}

func parseFlatTabularPattern(pattern string) (TabularPattern, error) {
	_, allLabels, err := parsePatternLabels(pattern)
	if err != nil {
		return TabularPattern{}, err
	}

	parts, seps, err := tokenizePattern(pattern)
	if err != nil {
		return TabularPattern{}, err
	}

	var slots []PatternSlot
	for _, p := range parts {
		if p == "" {
			continue
		}
		slotLabels := map[string]string{}
		if lbl, ok := allLabels[p]; ok {
			slotLabels[p] = lbl
		}
		slots = append(slots, PatternSlot{Dimension: p, AxisLabels: slotLabels})
	}

	tp := TabularPattern{Slots: slots, TopSeparators: seps}
	if err := validateTabularPattern(tp); err != nil {
		return TabularPattern{}, err
	}
	return tp, nil
}

func parseBracketedTabularPattern(pattern string) (TabularPattern, error) {
	var (
		slots    []PatternSlot
		topSeps  []string
		i        int
		needSlot = true
	)

	for i < len(pattern) {
		switch {
		case needSlot:
			slot, n, err := parseTabularSlot(pattern, i)
			if err != nil {
				return TabularPattern{}, err
			}
			slots = append(slots, slot)
			i = n
			needSlot = false

		default:
			start := i
			for i < len(pattern) && pattern[i] != '[' && !isTopLevelDimStart(pattern, i) {
				i++
			}
			if start == i {
				return TabularPattern{}, fmt.Errorf("invalid character %q in --group-pattern %q", pattern[i], pattern)
			}
			topSeps = append(topSeps, pattern[start:i])
			needSlot = true
		}
	}

	tp := TabularPattern{Slots: slots, TopSeparators: topSeps}
	if err := validateTabularPattern(tp); err != nil {
		return TabularPattern{}, err
	}
	return tp, nil
}

func parseTabularSlot(pattern string, i int) (PatternSlot, int, error) {
	if i >= len(pattern) {
		return PatternSlot{}, i, fmt.Errorf("unexpected end of --group-pattern %q", pattern)
	}

	if pattern[i] == '[' {
		end := findMatchingBracket(pattern, i)
		if end == -1 {
			return PatternSlot{}, i, fmt.Errorf("unclosed '[' in --group-pattern %q", pattern)
		}
		inner := pattern[i+1 : end]
		if inner == "" {
			return PatternSlot{}, i, fmt.Errorf("empty bracket slot [] in --group-pattern %q", pattern)
		}
		if strings.Contains(inner, "[") {
			return PatternSlot{}, i, fmt.Errorf("nested bracket slots are not supported in %q", pattern)
		}
		stripped, labels, err := parsePatternLabels(inner)
		if err != nil {
			return PatternSlot{}, i, err
		}
		if err := ValidateGroupPattern(stripped); err != nil {
			return PatternSlot{}, i, fmt.Errorf("invalid inner pattern %q: %w", inner, err)
		}
		return PatternSlot{ValueSplit: true, InnerPattern: stripped, AxisLabels: labels}, end + 1, nil
	}

	matches, err := findPatternMatches(pattern[i:])
	if err != nil || len(matches) == 0 {
		return PatternSlot{}, i, fmt.Errorf("expected dimension token or '[' at position %d in %q", i, pattern)
	}
	m := matches[0]
	if m.start != 0 {
		return PatternSlot{}, i, fmt.Errorf("expected dimension token or '[' at position %d in %q", i, pattern)
	}

	pos := i + len(m.raw)
	slotLabels := map[string]string{}
	if pos < len(pattern) && pattern[pos] == '{' {
		lbl, next, err := parseCurlyLabel(pattern, pos)
		if err != nil {
			return PatternSlot{}, i, err
		}
		slotLabels[m.expanded] = lbl
		pos = next
	}

	return PatternSlot{Dimension: m.expanded, AxisLabels: slotLabels}, pos, nil
}

func isTopLevelDimStart(pattern string, i int) bool {
	matches, err := findPatternMatches(pattern[i:])
	return err == nil && len(matches) > 0 && matches[0].start == 0
}

func findMatchingBracket(s string, open int) int {
	if open >= len(s) || s[open] != '[' {
		return -1
	}
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func validateTabularPattern(tp TabularPattern) error {
	if len(tp.Slots) == 0 {
		return fmt.Errorf("--group-pattern must define at least one slot")
	}

	var (
		hasXAxis bool
		hasYAxis bool
		hasZAxis bool
	)
	seen := map[string]bool{}

	for _, slot := range tp.Slots {
		var dims []string
		if slot.ValueSplit {
			dims = parsePatternParts(slot.InnerPattern)
		} else if slot.Dimension != "" {
			dims = []string{slot.Dimension}
		}

		for _, dim := range dims {
			if dim == "" {
				continue
			}
			if seen[dim] {
				return fmt.Errorf("duplicate dimension %q in --group-pattern", shortAxisKey(dim))
			}
			seen[dim] = true

			switch dim {
			case "xAxis":
				hasXAxis = true
			case "yAxis":
				hasYAxis = true
			case "zAxis":
				hasZAxis = true
			}
		}
	}

	if !hasXAxis && !hasYAxis {
		return errPatternNeedsXY
	}
	if hasZAxis && (!hasXAxis || !hasYAxis) {
		return errPatternZNeedsXY
	}
	return nil
}

// GroupTabularRow maps one row's group-column values through tabular pattern slots.
func GroupTabularRow(values []string, cfg Config) (map[string]string, error) {
	if cfg.TabularPattern == nil {
		return nil, fmt.Errorf("tabular pattern is not configured")
	}
	tp := *cfg.TabularPattern
	if len(values) != len(tp.Slots) {
		return nil, fmt.Errorf("expected %d group value(s) for %d pattern slot(s)", len(values), len(tp.Slots))
	}

	result := map[string]string{
		"name":  "",
		"xAxis": "",
		"yAxis": "",
		"zAxis": "",
	}

	for i, slot := range tp.Slots {
		val := strings.TrimSpace(values[i])
		if slot.ValueSplit {
			inner, err := ParseBenchmarkNameToGroups(val, slot.InnerPattern)
			if err != nil {
				return nil, fmt.Errorf("splitting column value %q with pattern %q: %w", val, slot.InnerPattern, err)
			}
			if err := mergeGroupMaps(result, inner); err != nil {
				return nil, err
			}
			continue
		}

		if result[slot.Dimension] != "" {
			return nil, fmt.Errorf("duplicate dimension %q in --group-pattern", shortAxisKey(slot.Dimension))
		}
		result[slot.Dimension] = val
	}

	return result, nil
}

func mergeGroupMaps(dst, src map[string]string) error {
	for dim, val := range src {
		if val == "" {
			continue
		}
		if dst[dim] != "" {
			return fmt.Errorf("duplicate dimension %q across pattern slots", shortAxisKey(dim))
		}
		dst[dim] = val
	}
	return nil
}

// TabularFilterLabel joins group-column values for --filter matching.
func TabularFilterLabel(values []string, cfg Config) string {
	if cfg.TabularPattern == nil {
		return JoinLabelParts(values, EffectiveLabelSeparators(cfg))
	}
	return JoinLabelParts(values, cfg.TabularPattern.TopSeparators)
}

func groupAxesFromTabular(cfg Config) []shared.Axis {
	if cfg.TabularPattern == nil {
		return nil
	}

	labels := EffectiveGroupColumns(cfg)
	var axes []shared.Axis
	labelIdx := 0

	for _, slot := range cfg.TabularPattern.Slots {
		colLabel := ""
		if labelIdx < len(labels) {
			colLabel = labels[labelIdx]
		}
		labelIdx++

		if slot.ValueSplit {
			for _, dim := range parsePatternParts(slot.InnerPattern) {
				if dim == "" {
					continue
				}
				axes = append(axes, shared.Axis{Key: shortAxisKey(dim), Label: slotLabel(slot, dim, colLabel)})
			}
			continue
		}

		axes = append(axes, shared.Axis{
			Key:   shortAxisKey(slot.Dimension),
			Label: slotLabel(slot, slot.Dimension, colLabel),
		})
	}

	return axes
}
