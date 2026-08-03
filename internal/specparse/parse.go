// Package specparse tokenizes structured CLI specs of the form:
//
//	prefix:prop(;prop)*
//	prop = key | key=value
//
// In legacy mode (no ';'), props are also comma-separated. MultiValueKeys and
// KnownKeys control how comma-bearing values are reassembled in that mode.
package specparse

import (
	"fmt"
	"strings"
)

// Prop is one key or key=value entry in a structured spec.
type Prop struct {
	Key      string // trimmed, original casing preserved
	Value    string // trimmed; empty if bare
	HasValue bool   // true when '=' was present (even if value empty)
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
	if strings.Contains(rest, ";") {
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

func parseSemicolonProps(rest string, opts Options) ([]Prop, error) {
	parts := strings.Split(rest, ";")
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
	tokens := strings.Split(rest, ",")
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
	return Prop{Key: key, Value: value, HasValue: true}, nil
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
