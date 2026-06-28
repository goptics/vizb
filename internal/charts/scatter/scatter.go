// Package scatter defines the typed Config for scatter charts. Structurally
// identical to line — scatter is a linear chart and shares Scale + ThreeDRotate —
// but kept as its own type so the per-chart Materialise remains typed.
package scatter

import (
	"github.com/goptics/vizb/internal/charts"
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

// New returns a fresh zero-value scatter chart Config.
func New() charts.ChartConfig { return &Config{} }
