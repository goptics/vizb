// Package line defines the typed Config for line charts. Structurally
// identical to bar — line is a linear chart and shares Scale + ThreeDRotate —
// but kept as its own type so the per-chart Materialise remains typed.
package line

import (
	"slices"

	"github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/shared"
)

const Type = "line"

type Config struct {
	Type            string             `json:"type"`
	Swap            string             `json:"swap,omitempty"`
	Sort            *shared.Sort       `json:"sort,omitempty"`
	Scale           string             `json:"scale,omitempty"`
	ShowLabels      *bool              `json:"showLabels,omitempty"`
	Symbol          string             `json:"symbol,omitempty"`
	SymbolSize      *float64           `json:"symbolSize,omitempty"`
	ThreeDRotate    *bool              `json:"threeDRotate,omitempty"`
	ThreeD          *bool              `json:"threeD,omitempty"`
	ThreeDVisualMap *bool              `json:"threeDVisualMap,omitempty"`
	Stat            *shared.StatConfig `json:"stat,omitempty"`
}

func (Config) ChartType() string { return Type }

func (c Config) StatEnabled() bool  { return c.Stat.StatEnabled() }
func (c Config) StatMath() []string { return c.Stat.StatMath() }
func (c Config) SwapString() string { return c.Swap }

func init() {
	charts.Register(charts.Spec{
		Type:    Type,
		Use:     "line [target]",
		Short:   "Generate a line chart from data",
		Long:    "Generate an interactive line chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Factory: func() charts.ChartConfig { return &Config{} },
		Flags: append(slices.Clone(charts.BaseChartFlags),
			charts.ScaleFlag, charts.ThreeDFlag, charts.ThreeDRotateFlag, charts.ThreeDVisualMapFlag,
			charts.SymbolFlag, charts.SymbolSizeFlag,
		),
	})
}
