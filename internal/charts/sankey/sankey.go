// Package sankey defines the typed Config for sankey charts. Sankey is a
// flow diagram (source→target→value), so Config intentionally omits Scale,
// stack, ThreeDRotate, and visualMap — those do not apply.
package sankey

import (
	"github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/shared"
)

const Type = "sankey"

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

// New returns a fresh zero-value sankey chart Config.
func New() charts.ChartConfig { return &Config{} }
