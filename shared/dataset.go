package shared

import (
	"encoding/json"
	"fmt"

	internal_charts "github.com/goptics/vizb/internal/charts"
)

type Stat struct {
	Type   string   `json:"type"`
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

type Dataset struct {
	ID          string                        `json:"id,omitempty"`
	Tag         string                        `json:"tag,omitempty"`
	Timestamp   string                        `json:"timestamp,omitempty"`
	Name        string                        `json:"name"`
	History     []HistoryEntry                `json:"history,omitempty"`
	Description string                        `json:"description,omitempty"`
	Meta        *Meta                         `json:"meta,omitempty"`
	Axes        []Axis                        `json:"axes"`
	Settings    []internal_charts.ChartConfig `json:"settings"`
	Data        []DataPoint                   `json:"data"`
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
		ID          string          `json:"id,omitempty"`
		Tag         string          `json:"tag,omitempty"`
		Timestamp   string          `json:"timestamp,omitempty"`
		Name        string          `json:"name"`
		History     []HistoryEntry  `json:"history,omitempty"`
		Description string          `json:"description,omitempty"`
		Meta        *Meta           `json:"meta,omitempty"`
		Axes        []Axis          `json:"axes"`
		Settings    json.RawMessage `json:"settings"`
		Data        []DataPoint     `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	d.ID = raw.ID
	d.Tag = raw.Tag
	d.Timestamp = raw.Timestamp
	d.Name = raw.Name
	d.History = raw.History
	d.Description = raw.Description
	d.Meta = raw.Meta
	d.Axes = raw.Axes
	d.Data = raw.Data

	// No settings, JSON null, or legacy v0.12.0 single object — leave
	// Settings nil so MigrateDataset can populate it from the legacy struct.
	if len(raw.Settings) == 0 || raw.Settings[0] != '[' {
		d.Settings = nil
		return nil
	}

	var entries []json.RawMessage
	if err := json.Unmarshal(raw.Settings, &entries); err != nil {
		return fmt.Errorf("dataset settings: expected JSON array: %w", err)
	}
	if len(entries) == 0 {
		d.Settings = nil
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
	return nil
}
