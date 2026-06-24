// Package bar defines the typed Config for bar charts and the Materialise
// function that applies the 4-step precedence (override > flags > defaults >
// internal default) to produce a fully-resolved config. Self-registers into
// the charts registry in init().
package bar

import (
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
	charts.Register(Type, func() charts.ChartConfig { return &Config{} })
}

// Flags carries the values Materialise reads from a command's flags (either
// the subcommand's own flags or the global defaults filled in by the root
// command). An empty/zero Flags means "no values supplied"; Materialise then
// falls back to override > internal default.
type Flags struct {
	Swap, Scale, Sort string
	ShowLabels        bool
	ThreeDRotate      bool
	ThreeD            bool
	ThreeDVisualMap   *bool
	Stat              []string
}

// Materialise produces a fully-resolved Config from the 4-step precedence:
// override > flags > defaults > internal default. See the design spec
// section 4 for the full description of each step.
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
		if override.Stat != nil {
			out.Stat = override.Stat
		}
	}

	if out.Scale == "" {
		out.Scale = "linear"
	}
	return out
}
