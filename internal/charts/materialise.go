package charts

import (
	"encoding/json"
	"fmt"
	"maps"
)

// Materialise builds a fully-resolved ChartConfig for chartType by merging, in
// increasing precedence: the chart's internal flag defaults (Flag.Default), the
// command seed (the values the invoking command supplied), and the --chart
// override.
//
// seed is a payload keyed by Config json tag — e.g.
// {"scale":"log","sort":{"enabled":true,"order":"asc"}}; it may be nil. Values
// may be JSON primitives or any json-marshalable value (e.g. a
// *shared.StatConfig the caller pre-built). override, when non-nil, is a typed
// Config (from ParseOverrides) whose set fields win over the seed; it is
// flattened to a sparse map via its omitempty json tags. The returned
// ChartConfig holds a *Config, since Decode allocates a pointer.
func Materialise(chartType string, seed map[string]any, override ChartConfig) (ChartConfig, error) {
	merged := map[string]any{"type": chartType}
	for _, f := range FlagsFor(chartType) {
		if f.Default != nil {
			merged[f.JSONKey] = f.Default
		}
	}
	maps.Copy(merged, seed)

	if override != nil {
		ob, err := json.Marshal(override)
		if err != nil {
			return nil, fmt.Errorf("materialise %s: marshal override: %w", chartType, err)
		}
		var om map[string]any
		if err := json.Unmarshal(ob, &om); err != nil {
			return nil, fmt.Errorf("materialise %s: decode override: %w", chartType, err)
		}
		maps.Copy(merged, om)
	}

	raw, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("materialise %s: %w", chartType, err)
	}
	return Decode(chartType, raw)
}
