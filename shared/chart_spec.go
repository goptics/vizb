package shared

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// validChartTypes is the set of chart types recognised by the --chart flag.
var validChartTypes = map[string]bool{
	"bar":     true,
	"line":    true,
	"pie":     true,
	"heatmap": true,
	"radar":   true,
}

// notAllowedForLimitedCharts lists settings not valid for pie/heatmap/radar.
var notAllowedForLimitedCharts = map[string]bool{
	"pie":     true,
	"heatmap": true,
	"radar":   true,
}

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

// ParseChartSpecs parses --chart flag specs into per-chart ChartSettings.
//
// specs: each element is one --chart value, e.g. "bar:swap=yxn,sort=asc"
// charts: the active chart types from --charts (e.g. ["bar","line","pie"])
// axes: ordered axes from GroupAxes() — used to validate swap permutations
func ParseChartSpecs(specs []string, charts []string, axes []Axis) (map[string]ChartSettings, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	// Build a fast lookup for active chart types.
	activeCharts := make(map[string]bool, len(charts))
	for _, c := range charts {
		activeCharts[strings.ToLower(c)] = true
	}

	// Compute the axis identity string for swap validation.
	identity := axisIdentity(axes)

	result := make(map[string]ChartSettings, len(specs))

	for _, spec := range specs {
		chartType, rest, ok := strings.Cut(spec, ":")
		if !ok {
			return nil, fmt.Errorf("--chart: malformed spec %q: expected <type>:<key>=<val>,... or <type>:<flag>", spec)
		}
		if rest == "" {
			return nil, fmt.Errorf("--chart: spec %q has no settings after ':'; expected <type>:<key>=<val> or <type>:<flag>", spec)
		}
		chartType = strings.ToLower(chartType)

		// Validate chart type is known.
		if !validChartTypes[chartType] {
			return nil, fmt.Errorf("--chart: unknown chart type %q (must be one of: bar, line, pie, heatmap, radar)", chartType)
		}

		// Validate chart type is active.
		if !activeCharts[chartType] {
			activeList := make([]string, 0, len(activeCharts))
			for k := range activeCharts {
				activeList = append(activeList, k)
			}
			sort.Strings(activeList)
			return nil, fmt.Errorf("--chart: chart type %q is not in the active --charts list (%s)", chartType, strings.Join(activeList, ", "))
		}

		// Parse comma-separated tokens.
		tokens := strings.Split(rest, ",")

		settings := result[chartType]

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
					t := true
					settings.ShowLabels = &t
				case "rotate":
					if notAllowedForLimitedCharts[chartType] {
						return nil, fmt.Errorf("--chart: key %q is not valid for chart type %q", key, chartType)
					}
					t := true
					settings.AutoRotate = &t
				default:
					return nil, fmt.Errorf("--chart: malformed token %q in spec %q: unknown bare flag (valid bare flags: labels, rotate)", token, spec)
				}
				continue
			}

			switch key {
			case "swap":
				if len(axes) > 0 {
					if !isPermutation(identity, val) {
						return nil, fmt.Errorf("--chart: swap value %q is not a permutation of the axis identity %q", val, identity)
					}
				} else {
					if val == "" {
						return nil, fmt.Errorf("--chart: swap value must be a non-empty string")
					}
				}
				settings.Swap = val

			case "sort":
				normalized := strings.ToLower(val)
				if normalized != "asc" && normalized != "desc" {
					return nil, fmt.Errorf("--chart: sort value %q is invalid (must be \"asc\" or \"desc\")", val)
				}
				settings.Sort = &Sort{Enabled: true, Order: normalized}

			case "scale":
				if notAllowedForLimitedCharts[chartType] {
					return nil, fmt.Errorf("--chart: key %q is not valid for chart type %q", key, chartType)
				}
				normalized := strings.ToLower(val)
				if normalized != "linear" && normalized != "log" {
					return nil, fmt.Errorf("--chart: scale value %q is invalid (must be \"linear\" or \"log\")", val)
				}
				settings.Scale = normalized

			case "labels":
				b, err := strconv.ParseBool(val)
				if err != nil {
					return nil, fmt.Errorf("--chart: key %q value %q must be true or false", key, val)
				}
				settings.ShowLabels = &b

			case "rotate":
				if notAllowedForLimitedCharts[chartType] {
					return nil, fmt.Errorf("--chart: key %q is not valid for chart type %q", key, chartType)
				}
				b, err := strconv.ParseBool(val)
				if err != nil {
					return nil, fmt.Errorf("--chart: key %q value %q must be true or false", key, val)
				}
				settings.AutoRotate = &b

			default:
				return nil, fmt.Errorf("--chart: unknown key %q in spec %q (valid keys: swap, sort, scale, labels, rotate)", key, spec)
			}
		}

		result[chartType] = settings
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}
