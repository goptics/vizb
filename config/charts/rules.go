package charts

import (
	"encoding/json"
	"fmt"

	"github.com/goptics/vizb/config/flags"
)

// AxisInfo is the lightweight axis metadata rules need. Avoids importing
// shared (which would create a cycle: shared imports config/charts).
type AxisInfo struct {
	Key  string // "x", "y", "z", "name"
	Type string // "value" for continuous, "" for categorical
}

// RuleContext carries the runtime context for evaluating flag applicability
// rules. Axes and ChartType are set once per Config; Value is set per flag.
type RuleContext struct {
	ChartType string     // e.g. "bar", "line"
	Axes      []AxisInfo // data-derived axes (post-parse, includes AutoGroup cases)
	Value     any        // this flag's current value from the marshalled Config
}

func axisMap(axes []AxisInfo) map[string]AxisInfo {
	m := make(map[string]AxisInfo, len(axes))
	for _, a := range axes {
		m[a.Key] = a
	}
	return m
}

func axisKeys(axes []AxisInfo) []string {
	out := make([]string, 0, len(axes))
	for _, a := range axes {
		out = append(out, a.Key)
	}
	return out
}

// RequiresAxes returns a rule that Skips the flag unless all named axes are
// present in the rule context. Panics with a clear message if keys is empty
// (programmer error).
func RequiresAxes(keys ...string) flags.RuleFn {
	if len(keys) == 0 {
		panic("RequiresAxes: at least one axis key is required")
	}
	return func(ctx any) (flags.Outcome, string) {
		rc, ok := ctx.(RuleContext)
		if !ok {
			return flags.Fatal, "internal: expected charts.RuleContext"
		}
		present := axisMap(rc.Axes)
		for _, k := range keys {
			if _, exists := present[k]; !exists {
				return flags.Skip, fmt.Sprintf("requires axis %q (present: %v)", k, axisKeys(rc.Axes))
			}
		}
		return flags.Keep, ""
	}
}

// RequiresZAxis is a convenience wrapper: RequiresAxes("z").
func RequiresZAxis() flags.RuleFn {
	return RequiresAxes("z")
}

// Requires3DMode returns a rule that Skips the flag when the runtime context
// doesn't support 3D rendering. 3D rendering is supported when either:
//   - a z-axis is present (explicit 3D data), or
//   - x, y, and z axes are all present with type "value" (value-mode xyz,
//     detected by auto-enable logic too late for the flag descriptor to see).
func Requires3DMode() flags.RuleFn {
	return func(ctx any) (flags.Outcome, string) {
		rc, ok := ctx.(RuleContext)
		if !ok {
			return flags.Fatal, "internal: expected charts.RuleContext"
		}
		hasZ := false
		hasX, hasY := false, false
		xyzValue := true
		for _, a := range rc.Axes {
			switch a.Key {
			case "x":
				hasX = true
				if a.Type != "value" {
					xyzValue = false
				}
			case "y":
				hasY = true
				if a.Type != "value" {
					xyzValue = false
				}
			case "z":
				hasZ = true
				if a.Type != "value" {
					xyzValue = false
				}
			}
		}
		// 3D mode is active if z-axis data exists, or if value-mode xyz.
		if hasZ || (hasX && hasY && hasX == hasY && hasX == xyzValue) {
			return flags.Keep, ""
		}
		return flags.Skip, "requires 3D-capable axes (z-axis or value-mode xyz); ignoring"
	}
}

// OnlyScatter2D returns a rule that Skips --visualmap when scatter is in xyz
// value-mode (where autoEnableValueMode3D forces 3D rendering). Checks that
// x, y, and z axes are all present with type "value".
func OnlyScatter2D() flags.RuleFn {
	return func(ctx any) (flags.Outcome, string) {
		rc, ok := ctx.(RuleContext)
		if !ok {
			return flags.Fatal, "internal: expected charts.RuleContext"
		}
		hasX, hasY, hasZ := false, false, false
		allValue := true
		for _, a := range rc.Axes {
			switch a.Key {
			case "x":
				hasX = true
				if a.Type != "value" {
					allValue = false
				}
			case "y":
				hasY = true
				if a.Type != "value" {
					allValue = false
				}
			case "z":
				hasZ = true
				if a.Type != "value" {
					allValue = false
				}
			}
		}
		if hasX && hasY && hasZ && allValue {
			return flags.Skip, "visualmap skipped: xyz value-mode forces 3D rendering"
		}
		return flags.Keep, ""
	}
}

// ApplyRules is the central pipeline pass. It evaluates every chart-flag
// descriptor's Rule list against each materialised Config, post-parse, with
// full data-derived axes.
//
// For each Config:
//  1. Marshal it to map[string]any (JSON round-trip, same pattern Materialise uses)
//  2. Walk the chart's flag descriptors via FlagsFor(chartType)
//  3. For each flag where len(Rule) > 0 and JSONKey is present in the map:
//     a. Build RuleContext{ChartType, Axes, Value: map[JSONKey]}
//     b. Evaluate every RuleFn; worst outcome per flag wins (Fatal > Skip > WarnKeep > Keep)
//     c. On Fatal → return immediately with the error (caller exits non-zero)
//     d. On Skip → delete JSONKey from the map, append warning message
//     e. On WarnKeep → append warning (keep value in map)
//  4. Re-decode the filtered map back to a typed ChartConfig via Decode(chartType, raw)
//  5. Replace the entry in the configs slice with the new filtered Config
//  6. Return accumulated warnings + nil error (or nil + fatal error)
func ApplyRules(ctx RuleContext, configs []ChartConfig) (warnings []string, fatal error) {
	for i, cfg := range configs {
		chartType := cfg.ChartType()
		ff := FlagsFor(chartType)
		if ff == nil {
			continue
		}

		raw, err := json.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("apply rules marshal %s: %w", chartType, err)
		}
		m := make(map[string]any)
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("apply rules decode %s: %w", chartType, err)
		}

		for _, f := range ff {
			if len(f.Rule) == 0 {
				continue
			}
			val, present := m[f.JSONKey]
			if !present {
				continue
			}

			perFlagCtx := RuleContext{
				ChartType: chartType,
				Axes:      ctx.Axes,
				Value:     val,
			}

			worst := flags.Keep
			worstMsg := ""
			for _, rule := range f.Rule {
				outcome, msg := rule(perFlagCtx)
				if outcome > worst {
					worst = outcome
					worstMsg = msg
				}
			}

			switch worst {
			case flags.Fatal:
				return nil, fmt.Errorf("flag %q: %s", f.Name, worstMsg)
			case flags.Skip:
				delete(m, f.JSONKey)
				warnings = append(warnings, fmt.Sprintf("flag %q skipped: %s", f.Name, worstMsg))
			case flags.WarnKeep:
				warnings = append(warnings, fmt.Sprintf("flag %q: %s", f.Name, worstMsg))
			}
		}

		filtered, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("apply rules marshal filtered %s: %w", chartType, err)
		}
		decoded, err := Decode(chartType, filtered)
		if err != nil {
			return nil, fmt.Errorf("apply rules decode filtered %s: %w", chartType, err)
		}
		configs[i] = decoded
	}

	return warnings, nil
}
