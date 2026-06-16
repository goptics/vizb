// Package line defines the typed Config for line charts. Structurally
// identical to bar — line is a linear chart and shares Scale + AutoRotate —
// but kept as its own type so the per-chart Materialise remains typed.
package line

import (
	"github.com/goptics/vizb/shared"
)

const Type = "line"

type Config struct {
	Type       string       `json:"type"`
	Swap       string       `json:"swap,omitempty"`
	Sort       *shared.Sort `json:"sort,omitempty"`
	Scale      string       `json:"scale,omitempty"`
	ShowLabels *bool        `json:"showLabels,omitempty"`
	AutoRotate *bool        `json:"autoRotate,omitempty"`
}

func (Config) ChartType() string { return Type }

func init() {
	shared.RegisterChartConfig(Type, func() shared.ChartConfig { return &Config{} })
}

type Flags struct {
	Swap, Scale, Sort string
	ShowLabels        bool
	AutoRotate        bool
}

func Materialise(flags Flags, override *Config) Config {
	out := Config{Type: Type}

	out.Swap = flags.Swap
	out.Scale = flags.Scale
	if flags.Sort != "" {
		out.Sort = &shared.Sort{Enabled: true, Order: flags.Sort}
	}
	if flags.ShowLabels {
		v := true
		out.ShowLabels = &v
	}
	if flags.AutoRotate {
		v := true
		out.AutoRotate = &v
	}

	if override != nil {
		if override.Swap != "" {
			out.Swap = override.Swap
		}
		if override.Sort != nil {
			out.Sort = override.Sort
		}
		if override.Scale != "" {
			out.Scale = override.Scale
		}
		if override.ShowLabels != nil {
			out.ShowLabels = override.ShowLabels
		}
		if override.AutoRotate != nil {
			out.AutoRotate = override.AutoRotate
		}
	}

	if out.Scale == "" {
		out.Scale = "linear"
	}
	return out
}
