package shared

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	config_charts "github.com/goptics/vizb/config/charts"
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
// Returns a map keyed by chart type, with values being the typed *<chart>.Config
// (e.g. *bar.Config for "bar"). The chart-type set is taken from the registry
// in config/charts; unknown chart types in specs return an error. The map is
// nil when no specs are supplied.
//
// Multiple specs for the same chart type are merged into a single Config
// (later tokens in a spec override earlier ones; later specs for the same
// chart type merge into the same Config).
//
// Chart-type-specific field validation (e.g. "pie has no Scale") is
// intentionally deferred to per-chart Validate(axes) methods. The
// JSON-payload approach used here lets any field be set in the payload; Decode
// drops fields that don't exist on the target Config struct.
func ParseOverrides(specs []string, charts []string, axes []Axis) (map[string]config_charts.ChartConfig, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	// Build a fast lookup for active chart types.
	activeCharts := make(map[string]bool, len(charts))
	for _, c := range charts {
		activeCharts[strings.ToLower(c)] = true
	}

	// Accumulate per-chart payloads; multiple specs for the same chart type
	// merge into the same payload rather than overwriting.
	payloads := make(map[string]map[string]any)
	var order []string

	for _, spec := range specs {
		chartType, rest, ok := strings.Cut(spec, ":")
		if !ok {
			return nil, fmt.Errorf("--chart: malformed spec %q: expected <type>:<key>=<val>,... or <type>:<flag>", spec)
		}
		if rest == "" {
			return nil, fmt.Errorf("--chart: spec %q has no settings after ':'; expected <type>:<key>=<val> or <type>:<flag>", spec)
		}
		chartType = strings.ToLower(chartType)

		// Validate the chart type is registered (replaces the old validChartTypes map).
		if _, err := config_charts.New(chartType); err != nil {
			return nil, fmt.Errorf("--chart: %w", err)
		}

		// Validate the chart type is active (selected via --charts).
		if !activeCharts[chartType] {
			activeList := make([]string, 0, len(activeCharts))
			for k := range activeCharts {
				activeList = append(activeList, k)
			}
			sort.Strings(activeList)
			return nil, fmt.Errorf("--chart: chart type %q is not in the active --charts list (%s)", chartType, strings.Join(activeList, ", "))
		}

		payload, ok := payloads[chartType]
		if !ok {
			payload = map[string]any{"type": chartType}
			payloads[chartType] = payload
			order = append(order, chartType)
		}

		// The payload approach (also used by shared/migrate.go) lets us populate
		// any typed Config without importing the per-chart subpackages — which
		// would cycle through this package (per-chart subpackages import shared
		// for *Sort).
		tokens := strings.Split(rest, ",")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}

			key, val, hasEq := strings.Cut(token, "=")
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)

			if !hasEq {
				// Bare flag.
				switch key {
				case "labels":
					payload["showLabels"] = true
				case "3d-rotate":
					payload["threeDRotate"] = true
				case "3d":
					payload["threeD"] = true
				case "3d-visualmap":
					payload["threeDVisualMap"] = true
				case "visualmap":
					if chartType != "scatter" {
						return nil, fmt.Errorf("--chart: bare flag %q is only valid for scatter charts", key)
					}
					payload["visualMap"] = true
				default:
					return nil, fmt.Errorf("--chart: malformed token %q in spec %q: unknown bare flag (valid bare flags: labels, 3d-rotate, 3d, 3d-visualmap, visualmap)", token, spec)
				}
				continue
			}

			switch key {
			case "swap":
				if err := ValidateSwap(val, axes); err != nil {
					return nil, fmt.Errorf("--chart: %w", err)
				}
				payload["swap"] = val

			case "sort":
				normalized := strings.ToLower(val)
				if normalized != "asc" && normalized != "desc" {
					return nil, fmt.Errorf("--chart: sort value %q is invalid (must be \"asc\" or \"desc\")", val)
				}
				payload["sort"] = Sort{Enabled: true, Order: normalized}

			case "scale":
				normalized := strings.ToLower(val)
				if normalized != "linear" && normalized != "log" {
					return nil, fmt.Errorf("--chart: scale value %q is invalid (must be \"linear\" or \"log\")", val)
				}
				payload["scale"] = normalized

			case "labels":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return nil, fmt.Errorf("--chart: key %q value %q must be true or false", key, val)
				}
				payload["showLabels"] = b

			case "3d-rotate":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return nil, fmt.Errorf("--chart: key %q value %q must be true or false", key, val)
				}
				payload["threeDRotate"] = b

			case "3d":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return nil, fmt.Errorf("--chart: key %q value %q must be true or false", key, val)
				}
				payload["threeD"] = b

			case "3d-visualmap":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return nil, fmt.Errorf("--chart: key %q value %q must be true or false", key, val)
				}
				payload["threeDVisualMap"] = b

			case "visualmap":
				if chartType != "scatter" {
					return nil, fmt.Errorf("--chart: key %q is only valid for scatter charts", key)
				}
				b, err := strconv.ParseBool(val)
				if err != nil {
					return nil, fmt.Errorf("--chart: key %q value %q must be true or false", key, val)
				}
				payload["visualMap"] = b

			default:
				return nil, fmt.Errorf("--chart: unknown key %q in spec %q (valid keys: swap, sort, scale, labels, 3d-rotate, 3d, 3d-visualmap, visualmap)", key, spec)
			}
		}
	}

	if len(payloads) == 0 {
		return nil, nil
	}

	result := make(map[string]config_charts.ChartConfig, len(payloads))
	for _, chartType := range order {
		raw, err := json.Marshal(payloads[chartType])
		if err != nil {
			return nil, fmt.Errorf("--chart: marshal payload: %w", err)
		}
		cfg, err := config_charts.Decode(chartType, raw)
		if err != nil {
			return nil, fmt.Errorf("--chart: decode %s: %w", chartType, err)
		}
		result[chartType] = cfg
	}

	return result, nil
}
