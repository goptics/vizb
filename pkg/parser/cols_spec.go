package parser

import (
	"fmt"
	"strings"
)

// ColumnSpec selects one csv/json value column and an optional chart label.
type ColumnSpec struct {
	Source string
	Label  string
}

// DisplayLabel returns the chart series label, defaulting to Source.
func (c ColumnSpec) DisplayLabel() string {
	if c.Label != "" {
		return c.Label
	}
	return c.Source
}

// ParseColsFlag parses --cols=price{Unit price},count into column specs.
func ParseColsFlag(raw string) ([]ColumnSpec, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	tokens, err := tokenizeColsFlag(raw)
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
			return nil, fmt.Errorf("empty column name in --cols")
		}
		if seen[spec.Source] {
			return nil, fmt.Errorf("duplicate column '%s' in --cols", spec.Source)
		}
		seen[spec.Source] = true
		specs = append(specs, spec)
	}
	return specs, nil
}

func parseColumnToken(tok string) (ColumnSpec, error) {
	if len(tok) >= 2 && tok[0] == '"' {
		source, err := parseQuotedColumnSource(tok)
		if err != nil {
			return ColumnSpec{}, err
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
		return ColumnSpec{}, fmt.Errorf("%w in --cols", err)
	}
	return ColumnSpec{Source: source, Label: label}, nil
}

func parseQuotedColumnSource(tok string) (string, error) {
	if tok[len(tok)-1] != '"' {
		return "", fmt.Errorf("unclosed '\"' in --cols")
	}
	inner := tok[1 : len(tok)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\\' && i+1 < len(inner) {
			b.WriteByte(inner[i+1])
			i++
			continue
		}
		b.WriteByte(inner[i])
	}
	return b.String(), nil
}

func tokenizeColsFlag(raw string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	inQuote := false

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case c == '"' && !inQuote:
			if cur.Len() > 0 {
				return nil, fmt.Errorf("unexpected '\"' in --cols")
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
		case c == '\\' && inQuote && i+1 < len(raw):
			cur.WriteByte(c)
			cur.WriteByte(raw[i+1])
			i++
		default:
			cur.WriteByte(c)
		}
	}

	if inQuote {
		return nil, fmt.Errorf("unclosed '\"' in --cols")
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens, nil
}

// ValidateColsGroupOverlap rejects columns named in both --cols and --group.
func ValidateColsGroupOverlap(cfg Config) error {
	if len(cfg.Cols) == 0 {
		return nil
	}
	groupSet := map[string]bool{}
	for _, g := range EffectiveGroupColumns(cfg) {
		groupSet[g] = true
	}
	for _, col := range cfg.Cols {
		if groupSet[col.Source] {
			return fmt.Errorf("column '%s' cannot be in both --cols and --group", col.Source)
		}
	}
	return nil
}