// Package bar defines the typed Config for bar charts and the Materialise
// function that applies the 4-step precedence (override > flags > defaults >
// internal default) to produce a fully-resolved config. Self-registers into
// the charts registry in init().
package bar

import (
	"github.com/goptics/vizb/shared"
)

// Type is the chart-type discriminator written to JSON and used as the
// registry key.
const Type = "bar"

// Config is the per-chart typed config for bar charts. bar/line are the only
// chart types that carry a Scale (linear/log) and AutoRotate (3D) — pie,
// heatmap, and radar omit them.
type Config struct {
	Type       string       `json:"type"` // always "bar"
	Swap       string       `json:"swap,omitempty"`
	Sort       *shared.Sort `json:"sort,omitempty"`
	Scale      string       `json:"scale,omitempty"`
	ShowLabels *bool        `json:"showLabels,omitempty"`
	AutoRotate *bool        `json:"autoRotate,omitempty"`
}

// ChartType returns the chart-type discriminator; satisfies shared.ChartConfig.
func (Config) ChartType() string { return Type }

func init() {
	shared.RegisterChartConfig(Type, func() shared.ChartConfig { return &Config{} })
}

// Flags carries the values Materialise reads from a command's flags (either
// the subcommand's own flags or the global defaults filled in by the root
// command). An empty/zero Flags means "no values supplied"; Materialise then
// falls back to override > internal default.
type Flags struct {
	Swap, Scale, Sort string
	ShowLabels        bool
	AutoRotate        bool
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
