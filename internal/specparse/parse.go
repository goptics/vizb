// Package specparse tokenizes structured CLI specs of the form:
//
//	prefix:prop(;prop)*
//	prop = key | key=scalar | key={ bag }
//	bag  = [ prop (';' prop)* ]   # semicolon-only inside {}
//
// A brace-delimited value is one object prop: ',' and ';' inside the braces do
// not create sibling props, so comma-bearing values (rgba(…), 8,8,0,0) survive
// intact. Unbalanced braces are a parse error.
//
// In legacy mode (no top-level ';'), props are also comma-separated.
// MultiValueKeys and KnownKeys control how comma-bearing values are
// reassembled in that mode.
package specparse

import (
	"fmt"
	"strings"
)

// Prop is one key or key=value entry in a structured spec. A prop whose value
// is brace-delimited carries the parsed bag in Object (non-nil, possibly
// empty) and leaves Value empty.
type Prop struct {
	Key      string // trimmed, original casing preserved
	Value    string // trimmed; empty if bare or object-valued
	HasValue bool   // true when '=' was present (even if value empty)
	Object   []Prop // brace-delimited bag fields; non-nil only for key={…} values
}

// Spec is a parsed prefix:props structured CLI token.
type Spec struct {
	Prefix string
	Props  []Prop // order preserved
}

// Options configures Parse behaviour.
type Options struct {
	// MultiValueKeys: in legacy comma mode, values for these keys may contain
	// commas. Matching is case-sensitive on the raw key (callers normalize).
	MultiValueKeys map[string]struct{}

	// KnownKeys: when non-empty, in legacy comma multi-value scanning, a token
	// that is `k` or `k=...` with k in KnownKeys starts a new prop.
	// When empty, multi-value consumption stops only at the next `k=...` token
	// (key= form), not bare keys — so `stat=center,spread` works alone but
	// `stat=center,spread,labels` needs KnownKeys or semicolon mode.
	KnownKeys map[string]struct{}

	// AllowBareKeys: allow key without '=' (chart). Default false means bare keys error.
	AllowBareKeys bool

	// RequireProps: error if prefix has no props after ':'.
	RequireProps bool
}

// Parse splits "prefix:props" into Spec. Does not validate domain keys/values.
func Parse(input string, opts Options) (Spec, error) {
	colon := strings.Index(input, ":")
	if colon < 0 {
		return Spec{}, fmt.Errorf("specparse: missing ':' in structured spec %q", input)
	}

	prefix := strings.TrimSpace(input[:colon])
	if prefix == "" {
		return Spec{}, fmt.Errorf("specparse: empty prefix in structured spec %q", input)
	}

	rest := strings.TrimSpace(input[colon+1:])
	if rest == "" {
		if opts.RequireProps {
			return Spec{}, fmt.Errorf("specparse: prefix %q requires at least one prop", prefix)
		}
		return Spec{Prefix: prefix}, nil
	}

	var props []Prop
	var err error
	hasSemicolon, sepErr := containsDepth0(rest, ';')
	if sepErr != nil {
		return Spec{}, sepErr
	}
	if hasSemicolon {
		props, err = parseSemicolonProps(rest, opts)
	} else {
		props, err = parseLegacyCommaProps(rest, opts)
	}
	if err != nil {
		return Spec{}, err
	}

	if len(props) == 0 && opts.RequireProps {
		return Spec{}, fmt.Errorf("specparse: prefix %q requires at least one prop", prefix)
	}

	return Spec{Prefix: prefix, Props: props}, nil
}

// SplitList comma-splits value, trims each part, and drops empty strings.
func SplitList(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ParseBag parses a bare semicolon bag of props ("key=value;key=value") — the
// unwrapped flag form of an object value (e.g. `--bg color=…;opacity=0.5`).
// Braces are rejected: the bag is the object body without delimiters.
func ParseBag(input string) ([]Prop, error) {
	if strings.ContainsAny(input, "{}") {
		return nil, fmt.Errorf("specparse: braces are not allowed in bag %q (pass fields directly: key=value;key=value)", input)
	}
	return parseObjectFields(strings.Split(input, ";"))
}

// splitDepth0 splits s on sep only outside any {...} object. Returns an error
// when braces are unbalanced.
func splitDepth0(s string, sep rune) ([]string, error) {
	parts := make([]string, 0, 1)
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("specparse: unmatched '}' in %q", s)
			}
		case sep:
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1 // sep is a one-byte rune
			}
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("specparse: unmatched '{' in %q", s)
	}
	return append(parts, s[start:]), nil
}

// containsDepth0 reports whether sep occurs in s outside any {...} object.
// Returns an error when braces are unbalanced.
func containsDepth0(s string, sep rune) (bool, error) {
	parts, err := splitDepth0(s, sep)
	return len(parts) > 1, err
}

func parseSemicolonProps(rest string, opts Options) ([]Prop, error) {
	parts, err := splitDepth0(rest, ';')
	if err != nil {
		return nil, err
	}
	props := make([]Prop, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		prop, err := parsePropSegment(part, opts)
		if err != nil {
			return nil, err
		}
		props = append(props, prop)
	}
	return props, nil
}

func parseLegacyCommaProps(rest string, opts Options) ([]Prop, error) {
	tokens, err := splitDepth0(rest, ',')
	if err != nil {
		return nil, err
	}
	props := make([]Prop, 0, len(tokens))

	for i := 0; i < len(tokens); {
		tok := strings.TrimSpace(tokens[i])
		if tok == "" {
			i++
			continue
		}

		prop, err := parsePropSegment(tok, opts)
		if err != nil {
			return nil, err
		}
		i++

		if prop.HasValue && keyInSet(prop.Key, opts.MultiValueKeys) {
			for i < len(tokens) {
				next := strings.TrimSpace(tokens[i])
				if next == "" {
					i++
					continue
				}
				if isPropStart(next, opts) {
					break
				}
				if prop.Value == "" {
					prop.Value = next
				} else {
					prop.Value = prop.Value + "," + next
				}
				i++
			}
		}

		props = append(props, prop)
	}

	return props, nil
}

func parsePropSegment(segment string, opts Options) (Prop, error) {
	eq := strings.Index(segment, "=")
	if eq < 0 {
		key := strings.TrimSpace(segment)
		if key == "" {
			return Prop{}, fmt.Errorf("specparse: empty key in prop %q", segment)
		}
		if !opts.AllowBareKeys {
			return Prop{}, fmt.Errorf("specparse: bare key %q not allowed (use key=value)", key)
		}
		return Prop{Key: key, HasValue: false}, nil
	}

	key := strings.TrimSpace(segment[:eq])
	if key == "" {
		return Prop{}, fmt.Errorf("specparse: empty key in prop %q", segment)
	}
	value := strings.TrimSpace(segment[eq+1:])
	if strings.HasPrefix(value, "{") {
		object, err := parseObjectBag(value)
		if err != nil {
			return Prop{}, err
		}
		return Prop{Key: key, HasValue: true, Object: object}, nil
	}
	return Prop{Key: key, Value: value, HasValue: true}, nil
}

// parseObjectBag parses a brace-delimited object value "{field=value;…}" into
// its bag fields. Only ';' separates fields inside the braces; ',' stays
// literal so values like rgba(…) and 8,8,0,0 survive.
func parseObjectBag(value string) ([]Prop, error) {
	if !strings.HasSuffix(value, "}") {
		return nil, fmt.Errorf("specparse: malformed object value %q (expected key={field=value;…})", value)
	}
	inner := strings.TrimSpace(value[1 : len(value)-1])
	if inner == "" {
		return []Prop{}, nil
	}
	if strings.ContainsAny(inner, "{}") {
		return nil, fmt.Errorf("specparse: nested braces are not allowed in object value %q", value)
	}
	return parseObjectFields(strings.Split(inner, ";"))
}

// parseObjectFields parses semicolon-separated scalar object fields. Empty
// segments are skipped.
func parseObjectFields(parts []string) ([]Prop, error) {
	fields := make([]Prop, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		field, err := parsePropSegment(part, Options{})
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

// isPropStart reports whether token should begin a new prop while scanning a
// multi-value value in legacy comma mode.
func isPropStart(token string, opts Options) bool {
	key, hasEq := splitPropKey(token)
	if key == "" {
		return false
	}
	if len(opts.KnownKeys) > 0 {
		return keyInSet(key, opts.KnownKeys)
	}
	// No KnownKeys: only a key= form starts a new prop.
	return hasEq
}

func splitPropKey(token string) (key string, hasEq bool) {
	eq := strings.Index(token, "=")
	if eq < 0 {
		return strings.TrimSpace(token), false
	}
	return strings.TrimSpace(token[:eq]), true
}

func keyInSet(key string, set map[string]struct{}) bool {
	if len(set) == 0 {
		return false
	}
	_, ok := set[key]
	return ok
}
