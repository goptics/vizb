package shared

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	internal_charts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/internal/specparse"
)

// axisChar returns the single character that represents an axis key in a swap
// permutation string.
func axisChar(key string) byte {
	if key == "name" {
		return 'n'
	}
	if len(key) > 0 {
		return key[0] // "x" → 'x', "y" → 'y', "z" → 'z'
	}
	return 0
}

// axisIdentity builds the identity swap string from an ordered slice of Axis
// values (e.g. [{key:"x"},{key:"y"},{key:"name"}] → "xyn").
func axisIdentity(axes []Axis) string {
	b := make([]byte, 0, len(axes))
	for _, a := range axes {
		if a.Key == "metric" {
			continue
		}
		c := axisChar(a.Key)
		if c != 0 {
			b = append(b, c)
		}
	}
	return string(b)
}

// isPermutation reports whether candidate is a permutation of identity.
func isPermutation(identity, candidate string) bool {
	if len(identity) != len(candidate) {
		return false
	}
	// Count characters in both strings and compare.
	counts := make(map[rune]int, len(identity))
	for _, c := range identity {
		counts[c]++
	}
	for _, c := range candidate {
		counts[c]--
		if counts[c] < 0 {
			return false
		}
	}
	return true
}

// ValidateSwap reports whether swap is a valid axis permutation for the given
// axes. An empty swap is always valid (no override). When axes are unknown
// (none configured) any non-empty swap is accepted. Reused by the per-chart
// --swap flag and the --chart spec parser.
func ValidateSwap(swap string, axes []Axis) error {
	if swap == "" {
		return nil
	}
	if len(axes) == 0 {
		return nil
	}
	identity := axisIdentity(axes)
	if !isPermutation(identity, swap) {
		return fmt.Errorf("swap value %q is not a permutation of the axis identity %q", swap, identity)
	}
	return nil
}

// ParseOverrides parses --chart flag specs into typed per-chart Configs.
//
// specs: each element is one --chart value, e.g. "bar:swap=yxn,sort=asc"
// charts: the active chart types from --charts (e.g. ["bar","line","pie"])
// axes: ordered axes from GroupAxes() — used to validate swap permutations
//
// Returns a map keyed by chart type (values are the typed *<chart>.Config),
// plus a slice of human-readable warnings for keys that were dropped because
// they are not applicable to the target chart. Each valid --chart key is the
// Name of a flag descriptor registered by the chart in config/charts; the key
// set is therefore derived, not hard-coded.
//
// Key handling is uniform: a key valid for the target chart is applied; a key
// valid for some other chart (e.g. `pie:scale`, `bar:visualmap`) is dropped
// with a warning; a key valid for no chart (a typo) is a hard error. Swap is
// validated against the runtime axes here, since descriptors are axis-unaware.
func ParseOverrides(specs []string, charts []string, axes []Axis) (map[string]internal_charts.ChartConfig, []string, error) {
	if len(specs) == 0 {
		return nil, nil, nil
	}

	// Build a fast lookup for active chart types.
	activeCharts := make(map[string]bool, len(charts))
	for _, c := range charts {
		activeCharts[strings.ToLower(c)] = true
	}

	allFlagNames := internal_charts.AllFlagNames()

	// Accumulate per-chart payloads; multiple specs for the same chart type
	// merge into the same payload rather than overwriting.
	payloads := make(map[string]map[string]any)
	var order []string
	var warnings []string

	for _, spec := range specs {
		// Chart type is needed before tokenization so MultiValueKeys / KnownKeys
		// can be built from that chart's flag registry.
		chartType, _, ok := strings.Cut(spec, ":")
		if !ok {
			return nil, nil, fmt.Errorf("--chart: malformed spec %q: expected <type>:<key>=<val>,... or <type>:<flag>", spec)
		}
		chartType = strings.ToLower(strings.TrimSpace(chartType))
		if chartType == "" {
			return nil, nil, fmt.Errorf("--chart: malformed spec %q: expected <type>:<key>=<val>,... or <type>:<flag>", spec)
		}

		// Validate the chart type is registered.
		if _, err := internal_charts.New(chartType); err != nil {
			return nil, nil, fmt.Errorf("--chart: %w", err)
		}

		// Validate the chart type is active (selected via --charts).
		if !activeCharts[chartType] {
			activeList := make([]string, 0, len(activeCharts))
			for k := range activeCharts {
				activeList = append(activeList, k)
			}
			sort.Strings(activeList)
			return nil, nil, fmt.Errorf("--chart: chart type %q is not in the active --charts list (%s)", chartType, strings.Join(activeList, ", "))
		}

		// Build a name → descriptor lookup for this chart's valid keys.
		chartFlags := internal_charts.FlagsFor(chartType)
		flagByName := make(map[string]flags.Flag, len(chartFlags))
		for _, f := range chartFlags {
			flagByName[f.EffectiveKey()] = f
		}

		parsed, err := specparse.Parse(spec, chartSpecParseOptions(chartFlags))
		if err != nil {
			return nil, nil, formatChartSpecParseError(spec, err)
		}

		payload, ok := payloads[chartType]
		if !ok {
			payload = map[string]any{"type": chartType}
			payloads[chartType] = payload
			order = append(order, chartType)
		}

		for _, prop := range parsed.Props {
			key := prop.Key
			f, known := flagByName[key]
			if !known {
				// Valid for another chart → drop with a warning; valid nowhere → typo.
				if allFlagNames[key] {
					warnings = append(warnings, fmt.Sprintf("--chart: key %q is not applicable to %s charts; ignored", key, chartType))
					continue
				}
				if !prop.HasValue {
					return nil, nil, fmt.Errorf("--chart: malformed token %q in spec %q: unknown bare flag (valid: %s)", key, spec, flagNameList(chartFlags))
				}
				return nil, nil, fmt.Errorf("--chart: unknown key %q in spec %q (valid keys: %s)", key, spec, flagNameList(chartFlags))
			}

			pv, err := convertFlagValue(f, prop, axes)
			if err != nil {
				return nil, nil, err
			}
			payload[f.JSONKey] = pv
		}
	}

	result := make(map[string]internal_charts.ChartConfig, len(payloads))
	for _, chartType := range order {
		raw, err := json.Marshal(payloads[chartType])
		if err != nil {
			return nil, nil, fmt.Errorf("--chart: marshal payload: %w", err)
		}
		cfg, err := internal_charts.Decode(chartType, raw)
		if err != nil {
			return nil, nil, fmt.Errorf("--chart: decode %s: %w", chartType, err)
		}
		result[chartType] = cfg
	}

	return result, warnings, nil
}

// convertFlagValue converts a --chart token value for flag f into its
// JSON-primitive payload value, validating along the way. prop.HasValue reports
// whether the token had an `=value`; bare bool, stat, and object flags (no `=`,
// or a brace-delimited object value) are valid. Swap is validated against axes
// (descriptors are axis-unaware).
func convertFlagValue(f flags.Flag, prop specparse.Prop, axes []Axis) (any, error) {
	val, hasEq := prop.Value, prop.HasValue
	switch f.Kind {
	case flags.KindObject:
		return convertObjectFlagValue(f, prop)
	case flags.KindBool:
		v := true
		if hasEq {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf("--chart: key %q value %q must be true or false", f.Name, val)
			}
			v = b
		}
		return encode(f, v), nil

	case flags.KindStat:
		if !hasEq {
			// bare stat token: enable stats with all categories
			return map[string]any{"enabled": true, "math": []string{}}, nil
		}
		// stat=a,b — validate each category and produce the typed payload
		parts := specparse.SplitList(val)
		math := make([]string, 0, len(parts))
		for _, p := range parts {
			if !slices.Contains(ValidStatMath, p) {
				return nil, fmt.Errorf("--chart: stat category %q is invalid (valid: %s)", p, strings.Join(ValidStatMath, ", "))
			}
			math = append(math, p)
		}
		return map[string]any{"enabled": true, "math": math}, nil

	default: // KindString, KindFloat, KindInt
		if f.Name == "scale" {
			return convertScaleFlagValue(f, prop)
		}
		if !hasEq {
			return nil, fmt.Errorf("--chart: key %q requires a value (e.g. %s=<value>)", f.Name, f.Name)
		}
		if f.Name == "swap" {
			if err := ValidateSwap(val, axes); err != nil {
				return nil, fmt.Errorf("--chart: %w", err)
			}
		} else if f.Validate != nil {
			if err := f.Validate(val); err != nil {
				return nil, fmt.Errorf("--chart: %w", err)
			}
		}
		if f.Kind == flags.KindFloat {
			n, _ := strconv.ParseFloat(val, 64) // Validate already ensured it parses
			return encode(f, n), nil
		}
		if f.Kind == flags.KindInt {
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("--chart: key %q value %q must be an integer", f.Name, val)
			}
			return encode(f, n), nil
		}
		return encode(f, val), nil
	}
}

// encode applies the flag's payload transform when present.
func encode(f flags.Flag, v any) any {
	if f.Encode != nil {
		return f.Encode(v)
	}
	return v
}

// convertScaleFlagValue converts the hybrid scale --chart token: scale=log
// (string), scale={type=log;axes=x} (braces), or an unwrapped bag
// (scale=type=log;…) which is a hard error with a brace hint.
func convertScaleFlagValue(f flags.Flag, prop specparse.Prop) (any, error) {
	switch {
	case !prop.HasValue:
		return nil, fmt.Errorf("--chart: key %q requires a value (e.g. %s=log or %s={type=log;axes=x})", f.Name, f.Name, f.Name)
	case prop.Object != nil:
		bag, err := ParseObjectBag(prop.Object, f.ObjectFields)
		if err != nil {
			return nil, fmt.Errorf("--chart: key %q: %w", f.Name, err)
		}
		return ScaleFromBag(bag).Payload(), nil
	default:
		if IsScaleBag(prop.Value) {
			return nil, fmt.Errorf(
				"--chart: key %q value %q must be a brace-delimited object, e.g. %s={type=log;axes=x}",
				f.Name, prop.Value, f.Name,
			)
		}
		if f.Validate != nil {
			if err := f.Validate(prop.Value); err != nil {
				return nil, fmt.Errorf("--chart: %w", err)
			}
		}
		return strings.ToLower(prop.Value), nil
	}
}

// convertObjectFlagValue converts a KindObject --chart token into its payload:
// bare `bar:bg` and empty `bar:bg={}` map to the empty bag (the flag's Encode
// injects the on-switch), `bar:bg={field=value;…}` to the typed bag, and an
// unwrapped scalar (`bar:bg=color=…`) is a hard error with a brace hint.
func convertObjectFlagValue(f flags.Flag, prop specparse.Prop) (any, error) {
	switch {
	case !prop.HasValue:
		return encode(f, map[string]any{}), nil
	case prop.Object != nil:
		bag, err := ParseObjectBag(prop.Object, f.ObjectFields)
		if err != nil {
			return nil, fmt.Errorf("--chart: key %q: %w", f.Name, err)
		}
		return encode(f, bag), nil
	default:
		return nil, fmt.Errorf(
			"--chart: key %q value %q must be a brace-delimited object, e.g. %s={color=rgba(180,180,180,0.2);borderColor=#000}",
			f.Name, prop.Value, f.Name,
		)
	}
}

// chartSpecParseOptions builds tokenizer options for a chart's --chart specs.
// KnownKeys are all EffectiveKey values; MultiValueKeys are flags whose values
// may contain commas (MultiValue, KindStat, or KindStringSlice).
func chartSpecParseOptions(chartFlags []flags.Flag) specparse.Options {
	known := make(map[string]struct{}, len(chartFlags))
	multi := make(map[string]struct{})
	for _, f := range chartFlags {
		key := f.EffectiveKey()
		known[key] = struct{}{}
		if f.MultiValue || f.Kind == flags.KindStat || f.Kind == flags.KindStringSlice {
			multi[key] = struct{}{}
		}
	}
	return specparse.Options{
		AllowBareKeys:  true,
		RequireProps:   true,
		KnownKeys:      known,
		MultiValueKeys: multi,
	}
}

// formatChartSpecParseError maps structural tokenizer errors to user-facing
// --chart messages without leaking the "specparse:" prefix.
func formatChartSpecParseError(spec string, err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "missing ':'"):
		return fmt.Errorf("--chart: malformed spec %q: expected <type>:<key>=<val>,... or <type>:<flag>", spec)
	case strings.Contains(msg, "requires at least one prop"):
		return fmt.Errorf("--chart: spec %q has no settings after ':'; expected <type>:<key>=<val> or <type>:<flag>", spec)
	default:
		return fmt.Errorf("--chart: malformed spec %q: %s", spec, strings.TrimPrefix(msg, "specparse: "))
	}
}

// flagNameList renders the comma-separated valid key names for an error message.
func flagNameList(fs []flags.Flag) string {
	names := make([]string, len(fs))
	for i, f := range fs {
		names[i] = f.EffectiveKey()
	}
	return strings.Join(names, ", ")
}
