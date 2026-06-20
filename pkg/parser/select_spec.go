package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// ColumnSpec selects one csv/json value column and an optional chart label.
type ColumnSpec struct {
	Source string
	Label  string
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