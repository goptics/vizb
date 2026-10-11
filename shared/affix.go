package shared

import (
	"fmt"
	"strings"
)

// Affix labels a dataset: symbols appended (or prepended) to metric names,
// either globally or per column. The dataset JSON always stores the
// structured form {"global": "...", "columns": {...}}.
type Affix struct {
	Global  string            `json:"global,omitempty"`
	Columns map[string]string `json:"columns,omitempty"`
}

// ParseAffix parses a --label-prefix/suffix flag value: a plain symbol
// ("%", "$") treated as global, or a comma/semicolon-separated list of
// global and per-column symbols ("%,col=ms", "col1=%;col2=ms").
// It returns nil only on error.
func ParseAffix(input string) (*Affix, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("label affix is empty")
	}
	sep := ","
	if strings.Contains(input, ";") {
		sep = ";"
	}
	affix := &Affix{}
	for _, tok := range strings.Split(input, sep) {
		tok = strings.TrimSpace(tok)
		key, val, found := strings.Cut(tok, "=")
		if !found {
			if affix.Global != "" {
				return nil, fmt.Errorf("label affix %q: duplicate global symbol", input)
			}
			affix.Global = tok
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			return nil, fmt.Errorf("label affix %q: empty column key", input)
		}
		if val == "" {
			return nil, fmt.Errorf("label affix %q: column %q has empty symbol", input, key)
		}
		if affix.Columns == nil {
			affix.Columns = map[string]string{}
		}
		if _, dup := affix.Columns[key]; dup {
			return nil, fmt.Errorf("label affix %q: duplicate column %q", input, key)
		}
		affix.Columns[key] = val
	}
	return affix, nil
}
