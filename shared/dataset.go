package shared

import (
	"encoding/json"
	"fmt"
	"strings"

	internal_charts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/pkg/style"
)

type Stat struct {
	Type   string   `json:"type,omitempty"`
	Value  *float64 `json:"value,omitempty"`
	Symbol string   `json:"symbol,omitempty"`
}

// F64 returns a pointer to f, used when setting Stat.Value so that zero
// measurements serialize as "value":0 rather than being omitted.
func F64(f float64) *float64 { return &f }

type DataPoint struct {
	Name   string `json:"name,omitempty"`
	XAxis  string `json:"xAxis,omitempty"`
	YAxis  string `json:"yAxis,omitempty"`
	ZAxis  string `json:"zAxis,omitempty"`
	Metric string `json:"metric,omitempty"` // value-mode visual metric (4th numeric column)
	Stats  []Stat `json:"stats,omitempty"`
}

type CPUInfo struct {
	Name  string `json:"name,omitempty"`
	Cores int    `json:"cores,omitempty"`
}

// Axis holds the key and optional human-readable label for a data dimension.
// Key is one of "name", "x", "y", "z" (in serial order). Type is "" (category,
// the default) or "value" (a continuous numeric coordinate axis, used by --axes
// value mode).
type Axis struct {
	Key   string `json:"key"`
	Label string `json:"label,omitempty"`
	Type  string `json:"type,omitempty"`
}

// Sort controls sort direction for chart data.
type Sort struct {
	Enabled bool   `json:"enabled"`
	Order   string `json:"order"` // "asc" or "desc"
}

// ChartConfig is the tiny contract every per-chart config implements. The
// canonical definition lives in config/charts/contract.go. The interface is
// imported here (under the internal_charts alias) so Dataset.Settings can be
// []ChartConfig without a per-chart package needing to know about shared.

type Meta struct {
	CPU  *CPUInfo `json:"cpu,omitempty"`
	OS   string   `json:"os,omitempty"`
	Arch string   `json:"arch,omitempty"`
	Pkg  string   `json:"pkg,omitempty"`
}

type HistoryEntry struct {
	Tag       string `json:"tag"`
	Timestamp string `json:"timestamp"`
	Meta      *Meta  `json:"meta,omitempty"`
}

// Theme is a fully expanded color theme embedded on a dataset.
// When Dataset.Themes is non-empty, Themes[0] is the active theme;
// there is no separate active-theme field on the wire format.
type Theme struct {
	Name            string   `json:"name"`
	Colors          []string `json:"colors"`
	VisualMapColors []string `json:"visualMapColors"`
}

type Dataset struct {
	ID        string `json:"id,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Name      string `json:"name"`
	// Themes is the data-owned theme catalog. Themes[0] is active when present.
	// New output writes Themes only (not the legacy Theme string).
	Themes []Theme `json:"themes,omitempty"`
	// Theme is the legacy single theme name/spec (pre-themes-array wire).
	// Used only when unmarshalling old files; migrateLegacyTheme expands it
	// into Themes and clears this field so re-marshal does not emit both.
	Theme       string                        `json:"theme,omitempty"`
	History     []HistoryEntry                `json:"history,omitempty"`
	Description string                        `json:"description,omitempty"`
	Meta        *Meta                         `json:"meta,omitempty"`
	Axes        []Axis                        `json:"axes"`
	Settings    []internal_charts.ChartConfig `json:"settings"`
	Data        []DataPoint                   `json:"data"`
	// PreserveRows tells the UI not to average duplicate (x,y,z) keys. True for
	// ungrouped csv/json tabular data (including solo/multi --select); false for
	// --group aggregations and benchmark parsers where rows are already collapsed.
	PreserveRows bool `json:"preserveRows,omitempty"`
}

// UnmarshalJSON decodes a Dataset, dispatching each entry in "settings" to the
// chart-type-specific Config via the charts registry. The new wire format is
//
//	"settings": [{"type":"bar",...}, {"type":"pie",...}]
//
// A v0.12.0 file uses
//
//	"settings": {"charts":[...], "sort":{...}, "showLabels":bool, "scale":string}
//
// — a single object. UnmarshalJSON cannot decode that into []ChartConfig, so
// it leaves Settings nil and MigrateDataset converts the legacy struct to the
// new shape. The default Marshal path (no MarshalJSON override) iterates the
// slice and writes each struct's `type` field naturally.
func (d *Dataset) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID           string          `json:"id,omitempty"`
		Tag          string          `json:"tag,omitempty"`
		Timestamp    string          `json:"timestamp,omitempty"`
		Name         string          `json:"name"`
		Themes       []Theme         `json:"themes,omitempty"`
		Theme        string          `json:"theme,omitempty"`
		History      []HistoryEntry  `json:"history,omitempty"`
		Description  string          `json:"description,omitempty"`
		Meta         *Meta           `json:"meta,omitempty"`
		Axes         []Axis          `json:"axes"`
		Settings     json.RawMessage `json:"settings"`
		Data         []DataPoint     `json:"data"`
		PreserveRows bool            `json:"preserveRows,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	d.ID = raw.ID
	d.Tag = raw.Tag
	d.Timestamp = raw.Timestamp
	d.Name = raw.Name
	d.Themes = raw.Themes
	d.Theme = raw.Theme
	d.History = raw.History
	d.Description = raw.Description
	d.Meta = raw.Meta
	d.Axes = raw.Axes
	d.Data = raw.Data
	d.PreserveRows = raw.PreserveRows

	// No settings, JSON null, or legacy v0.12.0 single object — leave
	// Settings nil so MigrateDataset can populate it from the legacy struct.
	if len(raw.Settings) == 0 || raw.Settings[0] != '[' {
		d.Settings = nil
		d.migrateLegacyTheme()
		return nil
	}

	var entries []json.RawMessage
	if err := json.Unmarshal(raw.Settings, &entries); err != nil {
		return fmt.Errorf("dataset settings: expected JSON array: %w", err)
	}
	if len(entries) == 0 {
		d.Settings = nil
		d.migrateLegacyTheme()
		return nil
	}

	d.Settings = make([]internal_charts.ChartConfig, 0, len(entries))
	for _, entry := range entries {
		var peek struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(entry, &peek); err != nil {
			return fmt.Errorf("dataset settings entry: %w", err)
		}
		if peek.Type == "" {
			return fmt.Errorf("dataset settings entry missing 'type' field: %s", entry)
		}
		cfg, err := internal_charts.Decode(peek.Type, entry)
		if err != nil {
			return err
		}
		d.Settings = append(d.Settings, cfg)
	}
	d.migrateLegacyTheme()
	return nil
}

// migrateLegacyTheme expands a legacy Theme string into Themes when Themes is
// empty. Themes[0] is the active theme when present. Built-in "default" and
// empty specs leave Themes empty (UI owns the default palette). Invalid specs
// leave Theme untouched so data is not discarded.
func (d *Dataset) migrateLegacyTheme() {
	if len(d.Themes) > 0 {
		// Prefer Themes; drop legacy string so re-marshal does not emit both.
		d.Theme = ""
		return
	}

	legacy := strings.TrimSpace(d.Theme)
	if legacy == "" || strings.EqualFold(legacy, "default") {
		d.Theme = ""
		return
	}

	parsed, err := style.ParseThemeSpec(legacy)
	if err != nil {
		return
	}
	if strings.EqualFold(parsed.Name, "default") {
		d.Theme = ""
		return
	}

	d.Themes = []Theme{themeFromStyle(parsed)}
	d.Theme = ""
}

func themeFromStyle(t style.Theme) Theme {
	return Theme{
		Name:            t.Name,
		Colors:          t.Colors,
		VisualMapColors: t.VisualMapColors,
	}
}
