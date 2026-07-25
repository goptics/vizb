// Package bar defines the typed Config for bar charts.
package bar

import (
	"github.com/goptics/vizb/internal/charts"
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
	Stack           *bool              `json:"stack,omitempty"`
	ShowLabels      *bool              `json:"showLabels,omitempty"`
	LabelMode       string             `json:"labelMode,omitempty"`
	ThreeDRotate    *bool              `json:"threeDRotate,omitempty"`
	ThreeD          *bool              `json:"threeD,omitempty"`
	ThreeDVisualMap *bool              `json:"threeDVisualMap,omitempty"`
	Horizontal      *bool              `json:"horizontal,omitempty"`
	Stat            *shared.StatConfig `json:"stat,omitempty"`
}

// ChartType returns the chart-type discriminator; satisfies charts.ChartConfig.
func (Config) ChartType() string { return Type }

func (c Config) StatEnabled() bool  { return c.Stat.StatEnabled() }
func (c Config) StatMath() []string { return c.Stat.StatMath() }
func (c Config) SwapString() string { return c.Swap }

// New returns a fresh zero-value bar chart Config.
func New() charts.ChartConfig { return &Config{} }
