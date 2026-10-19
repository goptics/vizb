package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goptics/vizb/shared"
)

// ColumnSpec selects one csv/json column and an optional chart label.
// AxisKey is set when --select solo mode uses explicit placement (x:col,y:col).
// AxisType is category or value after ResolveAxesTypes (mixed/value routing).
type ColumnSpec struct {
	Source   string
	Label    string
	AxisKey  string // x, y, or z from solo --select axis placement
	AxisType string // category | value (set by ResolveAxesTypes)
	Better   string // shared.BetterHigh/BetterLow from a trailing :high/:low
}

// ParseSelectFlag parses --select=price{Unit price},count into column specs.
func ParseSelectFlag(raw string) ([]ColumnSpec, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	tokens, err := tokenizeSelectFlag(raw)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	specs := make([]ColumnSpec, 0, len(tokens))
	for _, tok := range tokens {
		spec, err := parseColumnToken(tok)
		if err != nil {
			return nil, err
		}
		if spec.Source == "" {
			return nil, fmt.Errorf("empty column name in --select")
		}
		if seen[spec.Source] {
			return nil, fmt.Errorf("duplicate column '%s' in --select", spec.Source)
		}
		seen[spec.Source] = true
		specs = append(specs, spec)
	}
	return specs, nil
}

func parseColumnToken(tok string) (ColumnSpec, error) {
	tok, better := splitBetterSuffix(tok)
	spec, err := parseColumnSource(tok)
	spec.Better = better
	return spec, err
}

// splitBetterSuffix strips a trailing :high or :low. The colon must follow any
// closing quote or } so colons inside quoted names and labels stay literal.
func splitBetterSuffix(tok string) (string, string) {
	colon := strings.LastIndex(tok, ":")
	if colon == -1 {
		return tok, ""
	}
	dir, ok := shared.ParseBetterDir(tok[colon+1:])
	if !ok {
		return tok, ""
	}
	head := strings.TrimSpace(tok[:colon])
	if strings.HasPrefix(head, `"`) {
		if q, err := strconv.QuotedPrefix(head); err != nil || q != head {
			return tok, ""
		}
	} else if strings.LastIndex(head, "}") < strings.LastIndex(head, "{") {
		return tok, ""
	}
	return head, dir
}

func parseColumnSource(tok string) (ColumnSpec, error) {
	if len(tok) >= 2 && tok[0] == '"' {
		source, err := strconv.Unquote(tok)
		if err != nil {
			return ColumnSpec{}, fmt.Errorf("invalid quoted column in --select: %v", err)
		}
		return ColumnSpec{Source: source}, nil
	}

	open := strings.Index(tok, "{")
	if open == -1 {
		return ColumnSpec{Source: strings.TrimSpace(tok)}, nil
	}

	source := strings.TrimSpace(tok[:open])
	label, _, err := parseCurlyLabel(tok, open)
	if err != nil {
		return ColumnSpec{}, fmt.Errorf("%w in --select", err)
	}
	return ColumnSpec{Source: source, Label: label}, nil
}

func tokenizeSelectFlag(raw string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	inQuote := false

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c == '"' && !inQuote:
			if cur.Len() > 0 {
				return nil, fmt.Errorf("unexpected '\"' in --select")
			}
			inQuote = true
			cur.WriteByte(c)
		case c == '"' && inQuote:
			cur.WriteByte(c)
			inQuote = false
		case c == ',' && !inQuote:
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}

	if inQuote {
		return nil, fmt.Errorf("unclosed '\"' in --select")
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens, nil
}
