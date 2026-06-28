// Package scatter defines the typed Config for scatter charts. Structurally
// identical to line — scatter is a linear chart and shares Scale + ThreeDRotate —
// but kept as its own type so the per-chart Materialise remains typed.
package scatter

import (
	"slices"

	"github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/shared"
)

const Type = "scatter"

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
	VisualMap       *bool              `json:"visualMap,omitempty"`
	Stat            *shared.StatConfig `json:"stat,omitempty"`
}

func (Config) ChartType() string { return Type }

func (c Config) StatEnabled() bool  { return c.Stat.StatEnabled() }
func (c Config) StatMath() []string { return c.Stat.StatMath() }
func (c Config) SwapString() string { return c.Swap }

func init() {
	charts.Register(charts.Spec{
		Type:    Type,
		Use:     "scatter [target]",
		Short:   "Generate a scatter chart from data",
		Long:    "Generate an interactive scatter chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Factory: func() charts.ChartConfig { return &Config{} },
		Flags: append(slices.Clone(charts.BaseChartFlags),
			charts.ScaleFlag, charts.ThreeDFlag, charts.ThreeDRotateFlag, charts.ThreeDVisualMapFlag,
			charts.VisualMapFlag, charts.SymbolFlag, charts.SymbolSizeFlag,
		),
	})
}
