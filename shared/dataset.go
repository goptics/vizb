package shared

type Stat struct {
	Type   string  `json:"type"`
	Value  float64 `json:"value,omitempty"`
	Symbol string  `json:"symbol,omitempty"`
}

type DataPoint struct {
	Name  string `json:"name,omitempty"`
	XAxis string `json:"xAxis,omitempty"`
	YAxis string `json:"yAxis,omitempty"`
	ZAxis string `json:"zAxis,omitempty"`
	Stats []Stat `json:"stats"`
}

type CPUInfo struct {
	Name  string `json:"name,omitempty"`
	Cores int    `json:"cores,omitempty"`
}

// Axis holds the key and optional human-readable label for a data dimension.
// Key is one of "name", "x", "y", "z" (in serial order).
type Axis struct {
	Key   string `json:"key"`
	Label string `json:"label,omitempty"`
}

// Sort controls sort direction for chart data.
type Sort struct {
	Enabled bool   `json:"enabled"`
	Order   string `json:"order"`
}

// ChartSettings holds per-chart overrides; nil/"" means inherit the global setting.
type ChartSettings struct {
	Swap       string `json:"swap,omitempty"`
	Sort       *Sort  `json:"sort,omitempty"`
	Scale      string `json:"scale,omitempty"`
	ShowLabels *bool  `json:"showLabels,omitempty"`
	AutoRotate *bool  `json:"autoRotate,omitempty"`
}

// DatasetSettings holds global chart settings and per-chart overrides.
type DatasetSettings struct {
	Charts        []string                 `json:"charts"`
	Sort          Sort                     `json:"sort"`
	ShowLabels    bool                     `json:"showLabels"`
	Scale         string                   `json:"scale"`
	Axes          []Axis                   `json:"axes,omitempty"`
	ChartSettings map[string]ChartSettings `json:"chartSettings,omitempty"`
}

type HistoryEntry struct {
	Tag       string   `json:"tag"`
	Timestamp string   `json:"timestamp"`
	CPU       *CPUInfo `json:"cpu,omitempty"`
	OS        string   `json:"os,omitempty"`
}

type Dataset struct {
	Tag         string          `json:"tag,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Name        string          `json:"name"`
	History     []HistoryEntry  `json:"history,omitempty"`
	Description string          `json:"description,omitempty"`
	CPU         CPUInfo         `json:"cpu"`
	OS          string          `json:"os,omitempty"`
	Arch        string          `json:"arch,omitempty"`
	Pkg         string          `json:"pkg,omitempty"`
	Settings    DatasetSettings `json:"settings"`
	Data        []DataPoint     `json:"data"`
}
