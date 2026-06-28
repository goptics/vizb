// Package heatmap defines the typed Config for heatmap charts. Heatmap data
// is non-linear, so Config intentionally omits Scale and ThreeDRotate.
package heatmap

import (
	"github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/shared"
)

const Type = "heatmap"

type Config struct {
	Type       string             `json:"type"`
	Swap       string             `json:"swap,omitempty"`
	Sort       *shared.Sort       `json:"sort,omitempty"`
	ShowLabels *bool              `json:"showLabels,omitempty"`
	Stat       *shared.StatConfig `json:"stat,omitempty"`
}

func (Config) ChartType() string { return Type }

func (c Config) StatEnabled() bool  { return c.Stat.StatEnabled() }
func (c Config) StatMath() []string { return c.Stat.StatMath() }
func (c Config) SwapString() string { return c.Swap }

func init() {
	charts.Register(charts.Spec{
		Type:    Type,
		Use:     "heatmap [target]",
		Short:   "Generate a heatmap chart from data",
		Long:    "Generate an interactive heatmap chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Factory: func() charts.ChartConfig { return &Config{} },
		Flags:   charts.BaseChartFlags,
	})
}
