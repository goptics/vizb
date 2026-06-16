// Package heatmap defines the typed Config for heatmap charts. Heatmap data
// is non-linear, so Config intentionally omits Scale and AutoRotate.
package heatmap

import (
	"github.com/goptics/vizb/shared"
)

const Type = "heatmap"

type Config struct {
	Type       string       `json:"type"`
	Swap       string       `json:"swap,omitempty"`
	Sort       *shared.Sort `json:"sort,omitempty"`
	ShowLabels *bool        `json:"showLabels,omitempty"`
}

func (Config) ChartType() string { return Type }

func init() {
	shared.RegisterChartConfig(Type, func() shared.ChartConfig { return &Config{} })
}

type Flags struct {
	Swap, Sort string
	ShowLabels bool
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
	}

	return out
}
