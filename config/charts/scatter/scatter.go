// Package scatter defines the typed Config for scatter charts. Structurally
// identical to line — scatter is a linear chart and shares Scale + ThreeDRotate —
// but kept as its own type so the per-chart Materialise remains typed.
package scatter

import (
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
	charts.Register(Type, func() charts.ChartConfig { return &Config{} })
}

type Flags struct {
	Swap, Scale, Sort string
	ShowLabels        bool
	ThreeDRotate      bool
	ThreeD            bool
	ThreeDVisualMap   *bool
	VisualMap         *bool
	Stat              []string
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
	if flags.ThreeDRotate {
		v := true
		out.ThreeDRotate = &v
	}
	if flags.ThreeD {
		v := true
		out.ThreeD = &v
	}
	if flags.ThreeDVisualMap != nil {
		out.ThreeDVisualMap = flags.ThreeDVisualMap
	}
	if flags.VisualMap != nil {
		out.VisualMap = flags.VisualMap
	}

	out.Stat = shared.MaterialiseStatFlags(flags.Stat)

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
		if override.ThreeDRotate != nil {
			out.ThreeDRotate = override.ThreeDRotate
		}
		if override.ThreeD != nil {
			out.ThreeD = override.ThreeD
		}
		if override.ThreeDVisualMap != nil {
			out.ThreeDVisualMap = override.ThreeDVisualMap
		}
		if override.VisualMap != nil {
			out.VisualMap = override.VisualMap
		}
		if override.Stat != nil {
			out.Stat = override.Stat
		}
	}

	if out.Scale == "" {
		out.Scale = "linear"
	}
	return out
}
