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
	charts.Register(Type, func() charts.ChartConfig { return &Config{} })
}

type Flags struct {
	Swap, Sort string
	ShowLabels bool
	Stat       []string
}

func Materialise(flags Flags, override *Config) Config {
	out := Config{Type: Type}

	out.Swap = flags.Swap
	if flags.Sort != "" {
		out.Sort = &shared.Sort{Enabled: true, Order: flags.Sort}
	}
	if flags.ShowLabels {
		v := true
		out.ShowLabels = &v
	}

	out.Stat = shared.MaterialiseStatFlags(flags.Stat)

	if override != nil {
		if override.Swap != "" {
			out.Swap = override.Swap
		}
		if override.Sort != nil {
			out.Sort = override.Sort
		}
		if override.ShowLabels != nil {
			out.ShowLabels = override.ShowLabels
		}
		if override.Stat != nil {
			out.Stat = override.Stat
		}
	}

	return out
}
