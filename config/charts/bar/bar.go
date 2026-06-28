// Package bar defines the typed Config for bar charts and the Materialise
// function that applies the 4-step precedence (override > flags > defaults >
// internal default) to produce a fully-resolved config. Self-registers into
// the charts registry in init().
package bar

import (
	"slices"

	"github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/shared"
)

// Type is the chart-type discriminator written to JSON and used as the
// registry key.
const Type = "bar"

// Config is the per-chart typed config for bar charts. bar/line are the only
// chart types that carry a Scale (linear/log) and ThreeDRotate (3D) — pie,
// heatmap, and radar omit them.
type Config struct {
	Type            string             `json:"type"` // always "bar"
	Swap            string             `json:"swap,omitempty"`
	Sort            *shared.Sort       `json:"sort,omitempty"`
	Scale           string             `json:"scale,omitempty"`
	ShowLabels      *bool              `json:"showLabels,omitempty"`
	ThreeDRotate    *bool              `json:"threeDRotate,omitempty"`
	ThreeD          *bool              `json:"threeD,omitempty"`
	ThreeDVisualMap *bool              `json:"threeDVisualMap,omitempty"`
	Stat            *shared.StatConfig `json:"stat,omitempty"`
}

// ChartType returns the chart-type discriminator; satisfies charts.ChartConfig.
func (Config) ChartType() string { return Type }

func (c Config) StatEnabled() bool  { return c.Stat.StatEnabled() }
func (c Config) StatMath() []string { return c.Stat.StatMath() }
func (c Config) SwapString() string { return c.Swap }

func init() {
	charts.Register(charts.Spec{
		Type:    Type,
		Use:     "bar [target]",
		Short:   "Generate a bar chart from data",
		Long:    "Generate an interactive bar chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Factory: func() charts.ChartConfig { return &Config{} },
		Flags: append(slices.Clone(charts.BaseChartFlags),
			charts.ScaleFlag, charts.ThreeDFlag, charts.ThreeDRotateFlag, charts.ThreeDVisualMapFlag,
		),
	})
}
