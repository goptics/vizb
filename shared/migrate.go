package shared

import (
	"encoding/json"

	config_charts "github.com/goptics/vizb/config/charts"
)

// MigrateDataset applies in-memory schema migrations to ds. The raw bytes from
// which ds was unmarshalled must be supplied so migrations that need the
// original field set (e.g. legacy v0.12.0 shapes that the typed structs no
// longer accept) can recover the values. Pass nil to skip migration.
//
// Two passes run in order:
//  1. Legacy flat meta fields (cpu/os/arch/pkg) → ds.Meta. Pre-dates the Meta
//     struct. Unchanged from the previous version.
//  2. v0.12.0 settings struct (charts/sort/showLabels/scale) → per-chart
//     typed Configs in ds.Settings. Only fires when ds.Settings is empty AND
//     the legacy struct is present. Unregistered chart types in the legacy
//     file are silently dropped (buildLegacyConfig returns an error which
//     the caller skips).
//
// The v0.12.0 → new-shape helpers (axesFromDataPoints, buildLegacyConfig)
// live in this file alongside the migration entry point so the whole
// migration is in one place. They were originally planned to live in
// config/charts/migrate.go (per the design spec, section 5), but the import
// graph is fixed in the wrong direction: per-chart packages import shared
// (for Sort and DataPoint), and the migration helpers need to dispatch back
// through the registry to build typed Configs — moving them to
// config/charts would import shared, which (with ChartConfig in
// config/charts) means config/charts/migrate.go imports shared while
// shared imports config/charts for Dataset.Settings. The only cycle-free
// home is shared itself, so the helpers stay here.
func MigrateDataset(ds *Dataset, rawJSON []byte) {
	if len(rawJSON) == 0 {
		return
	}

	// Pass 1: legacy flat meta fields (cpu/os/arch/pkg → Meta).
	var legacyMeta struct {
		CPU     *CPUInfo `json:"cpu"`
		OS      string   `json:"os"`
		Arch    string   `json:"arch"`
		Pkg     string   `json:"pkg"`
		History []struct {
			CPU *CPUInfo `json:"cpu"`
			OS  string   `json:"os"`
		} `json:"history"`
	}
	if err := json.Unmarshal(rawJSON, &legacyMeta); err != nil {
		return
	}

	if ds.Meta == nil {
		m := &Meta{CPU: legacyMeta.CPU, OS: legacyMeta.OS, Arch: legacyMeta.Arch, Pkg: legacyMeta.Pkg}
		if m.CPU != nil || m.OS != "" || m.Arch != "" || m.Pkg != "" {
			ds.Meta = m
		}
	}

	for i := range ds.History {
		if i >= len(legacyMeta.History) {
			break
		}
		if ds.History[i].Meta == nil {
			leg := legacyMeta.History[i]
			m := &Meta{CPU: leg.CPU, OS: leg.OS}
			if m.CPU != nil || m.OS != "" {
				ds.History[i].Meta = m
			}
		}
	}

	// Pass 2: v0.12.0 settings struct → per-chart typed Configs.
	if len(ds.Settings) > 0 {
		return // already in new shape
	}
	var legacySettings struct {
		Settings struct {
			Charts     []string `json:"charts"`
			Sort       Sort     `json:"sort"`
			ShowLabels bool     `json:"showLabels"`
			Scale      string   `json:"scale"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(rawJSON, &legacySettings); err != nil {
		return
	}
	if len(legacySettings.Settings.Charts) == 0 {
		return
	}

	// Derive Axes from data points (v0.12.0 had no axes field). Empty
	// Axes is fine — the UI handles missing axes gracefully.
	ds.Axes = axesFromDataPoints(ds.Data)

	legacySort := legacySettings.Settings.Sort
	legacyShowLabels := legacySettings.Settings.ShowLabels
	legacyScale := legacySettings.Settings.Scale

	for _, t := range legacySettings.Settings.Charts {
		cfg, err := buildLegacyConfig(t, legacySort, legacyShowLabels, legacyScale)
		if err != nil {
			continue // unknown chart type in legacy file — skip silently
		}
		ds.Settings = append(ds.Settings, cfg)
	}
}

// axesFromDataPoints derives a minimal Axis list from the data points in a
// v0.12.0 file. It scans for non-empty XAxis/YAxis/ZAxis values and emits the
// corresponding Axis{Key: ...} entry. Labels are not recoverable from v0.12.0
// (the format didn't store them), so the Label field stays empty.
func axesFromDataPoints(data []DataPoint) []Axis {
	var axes []Axis
	seen := map[string]bool{}
	for _, dp := range data {
		for _, key := range []string{"x", "y", "z"} {
			if seen[key] {
				continue
			}
			var val string
			switch key {
			case "x":
				val = dp.XAxis
			case "y":
				val = dp.YAxis
			case "z":
				val = dp.ZAxis
			}
			if val != "" {
				axes = append(axes, Axis{Key: key})
				seen[key] = true
			}
		}
		if len(seen) == 3 {
			break
		}
	}
	return axes
}

// buildLegacyConfig constructs a typed ChartConfig for a single chart type
// from the legacy v0.12.0 settings fields. Returns an error when the chart
// type is unknown — callers should silently skip the error (unregistered
// types are dropped, not fatal).
//
// Field assignment (per spec):
//   - bar/line: Type=typ, Sort=&legacySort, ShowLabels=&legacyShowLabels,
//     Scale=legacyScale (or "linear" if empty)
//   - pie/heatmap/radar: Type=typ, Sort=&legacySort, ShowLabels=&legacyShowLabels
//     (no Scale field)
//
// The payload approach (JSON-marshal a map and dispatch via Decode) lets us
// populate the typed Config without importing the per-chart subpackages
// directly (bar/line/pie/heatmap/radar each import shared, which would
// cycle through this package).
func buildLegacyConfig(typ string, sort Sort, showLabels bool, scale string) (config_charts.ChartConfig, error) {
	// Build the legacy payload as a generic map. We always include "scale";
	// bar/line decode it into their Scale field, and pie/heatmap/radar
	// silently drop it (they have no Scale field).
	scaleVal := scale
	if scaleVal == "" {
		scaleVal = "linear"
	}
	payload := map[string]any{
		"type":       typ,
		"sort":       sort,
		"showLabels": showLabels,
		"scale":      scaleVal,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return config_charts.Decode(typ, raw)
}
