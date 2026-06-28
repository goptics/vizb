package parser

import (
	"fmt"
	"strings"
)

// ParseSelectViewFlag parses one solo --select value (e.g. "region,latency" or
// "x:region,y:latency,z:sales") into 2–3 column specs with x/y/z axis keys.
// Reuses tokenizeSelectFlag from select_spec.go for {label}/quoting.
func ParseSelectViewFlag(raw string) ([]ColumnSpec, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("--select requires 2 or 3 columns (x,y[,z]); got 0")
	}

	tokens, err := tokenizeSelectFlag(raw)
	if err != nil {
		return nil, err
	}
	if n := len(tokens); n < 2 || n > 3 {
		return nil, fmt.Errorf("--select requires 2 or 3 columns (x,y[,z]); got %d", n)
	}

	seenCol := map[string]bool{}
	seenKey := map[string]bool{}
	specs := make([]ColumnSpec, 0, len(tokens))
	explicitCount := 0

	for i, tok := range tokens {
		spec, key, isExplicit, err := parseAxisToken(tok)
		if err != nil {
			return nil, err
		}
		if spec.Source == "" {
			return nil, fmt.Errorf("empty column name in --select")
		}
		if seenCol[spec.Source] {
			return nil, fmt.Errorf("duplicate column '%s' in --select", spec.Source)
		}
		seenCol[spec.Source] = true

		if isExplicit {
			explicitCount++
			if seenKey[key] {
				return nil, fmt.Errorf("duplicate axis key '%s' in --select", key)
			}
			seenKey[key] = true
			spec.AxisKey = key
		} else {
			keys := []string{"x", "y", "z"}
			spec.AxisKey = keys[i]
		}
		specs = append(specs, spec)
	}

	if explicitCount > 0 && explicitCount != len(tokens) {
		return nil, fmt.Errorf("--select: use explicit x:/y:/z: syntax for every column, or omit prefixes for all")
	}
	if explicitCount > 0 {
		if err := validateExplicitSelectAxisKeys(specs); err != nil {
			return nil, err
		}
	}

	return specs, nil
}

func parseAxisToken(tok string) (ColumnSpec, string, bool, error) {
	colon := strings.Index(tok, ":")
	if colon <= 0 {
		spec, err := parseColumnToken(tok)
		return spec, "", false, err
	}

	key := strings.TrimSpace(tok[:colon])
	if key != "x" && key != "y" && key != "z" {
		spec, err := parseColumnToken(tok)
		return spec, "", false, err
	}

	spec, err := parseColumnToken(strings.TrimSpace(tok[colon+1:]))
	if err != nil {
		return spec, key, true, err
	}
	return spec, key, true, nil
}

func validateExplicitSelectAxisKeys(specs []ColumnSpec) error {
	has := map[string]bool{}
	for _, s := range specs {
		has[s.AxisKey] = true
	}
	if !has["x"] {
		return fmt.Errorf("--select explicit syntax requires x:column (e.g. x:region,y:latency)")
	}
	if !has["y"] {
		return fmt.Errorf("--select explicit syntax requires y:column")
	}
	if len(specs) == 3 && !has["z"] {
		return fmt.Errorf("--select with 3 columns requires z:column in explicit syntax")
	}
	return nil
}

// HasSelect reports whether the user supplied any --select configuration.
func HasSelect(cfg Config) bool {
	return len(cfg.Select) > 0 || len(cfg.SelectViews) > 0
}

// IsExplicitGrouping reports whether the user supplied explicit grouping flags.
func IsExplicitGrouping(cfg Config) bool {
	return len(cfg.Group) > 0 ||
		cfg.GroupRegex != "" ||
		(cfg.GroupPattern != "" && cfg.GroupPattern != "x")
}

// IsSelectAxisMode reports solo --select axis mode: select views without explicit grouping.
func IsSelectAxisMode(cfg Config) bool {
	return len(cfg.SelectViews) > 0 && !IsExplicitGrouping(cfg)
}

// AxisColumnLabel returns the flag name for axis-column error messages.
func AxisColumnLabel(selectMode bool) string {
	if selectMode {
		return "--select"
	}
	return "--axes"
}
