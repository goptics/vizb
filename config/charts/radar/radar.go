// Package radar defines the typed Config for radar charts. Radar data is
// non-linear, so Config intentionally omits Scale and ThreeDRotate.
package radar

import (
	"github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/shared"
)

const Type = "radar"

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
		Use:     "radar [target]",
		Short:   "Generate a radar chart from data",
		Long:    "Generate an interactive radar chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Factory: func() charts.ChartConfig { return &Config{} },
		Flags:   charts.BaseChartFlags,
	})
}
