package shared

import (
	"slices"

	internal_charts "github.com/goptics/vizb/internal/charts"
)

// ValidStatMath is the ordered list of accepted stat category values.
var ValidStatMath = []string{
	"counts", "center", "spread", "extremes", "shape", "percentiles", "confidence", "correlations",
}

// StatConfig controls which stat categories appear in the Stats panel per chart.
// Math empty + Enabled true means all categories. Enabled false hides the Stats button.
type StatConfig struct {
	Enabled bool     `json:"enabled"`
	Math    []string `json:"math"` // empty = all categories
}

func (s *StatConfig) StatEnabled() bool {
	return s != nil && s.Enabled
}

func (s *StatConfig) StatMath() []string {
	if s == nil {
		return nil
	}
	return s.Math
}

// MaterialiseStatFlags converts a raw []string flag value into a *StatConfig.
// Nil flags (--stat not passed) returns nil so omitempty elides the field in JSON.
func MaterialiseStatFlags(flags []string) *StatConfig {
	if flags == nil {
		return nil
	}
	math := flags
	if len(math) == 0 {
		math = []string{}
	}
	return &StatConfig{Enabled: true, Math: math}
}

// StatNeedsCorrelation reports whether the correlation heatmap renderer should
// be shipped. True when math is empty (all categories) or explicitly contains "correlations".
func StatNeedsCorrelation(math []string) bool {
	return len(math) == 0 || slices.Contains(math, "correlations")
}

// NeedsCorrelation reports whether this stat config requires the correlation
// heatmap chunk. False when stat is nil or disabled.
func (s *StatConfig) NeedsCorrelation() bool {
	return s != nil && s.Enabled && StatNeedsCorrelation(s.Math)
}

// ChartConfigNeedsCorrelation reports whether a chart config requires the
// correlation heatmap chunk to be shipped.
func ChartConfigNeedsCorrelation(cfg internal_charts.ChartConfig) bool {
	return cfg.StatEnabled() && StatNeedsCorrelation(cfg.StatMath())
}
