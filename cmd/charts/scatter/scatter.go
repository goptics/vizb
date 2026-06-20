// Package scatter registers the `vizb scatter` subcommand: a scatter chart with the full
// set of linear flags plus --scale and --3d-rotate (3D-only).
package scatter

import (
	"fmt"
	"strings"

	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() { cli.Register(NewCommand) }

// Options adds the scatter-specific flags (--scale, --3d-rotate) on top of the shared
// chart options. pie/heatmap/radar omit these, enforced at compile time.
type Options struct {
	cli.ChartOptions
	Axes            string
	Scale           string
	ThreeDRotate    bool
	ThreeD          bool
	ThreeDVisualMap bool
}

// Bind registers the shared chart flags plus --scale and --3d-rotate.
func (o *Options) Bind(fs *pflag.FlagSet) {
	o.ChartOptions.Bind(fs)
	fs.StringVarP(&o.Scale, "scale", "S", "linear", "Scale type (linear, log)")
	fs.BoolVar(&o.ThreeD, "3d", false, "Enable value 3D for x+y data (y categories on depth, metric on height)")
	fs.BoolVar(&o.ThreeDRotate, "3d-rotate", false, "Auto-rotate the 3D scene (only applies when z-axis data is present)")
	fs.BoolVar(&o.ThreeDVisualMap, "3d-visualmap", false, "Color 3D bars/lines by metric value (visualMap gradient)")
	fs.StringVar(&o.Axes, "axes", "", "csv/json only: plot 2-3 numeric columns as raw x,y[,z] coordinates; with --group use exactly 1 axes column (2 group dims); mutually exclusive with --select/--group-regex")
}

// ParseConfig converts scatter flags into parser.Config, including scatter-only
// --axes handling (pure value mode or hybrid group+axes mode).
func (o *Options) ParseConfig() parser.Config {
	cfg := o.CommonOptions.ParseConfig()
	axesRaw := strings.TrimSpace(o.Axes)
	if axesRaw == "" {
		return cfg
	}

	if strings.TrimSpace(o.Select) != "" {
		shared.ExitWithError("--axes cannot be combined with --select", nil)
	}
	if strings.TrimSpace(o.GroupRegex) != "" {
		shared.ExitWithError("--axes cannot be combined with --group-regex", nil)
	}

	if len(o.Group) > 0 {
		axes, err := parser.ParseSelectFlag(axesRaw)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		if len(axes) != 1 {
			shared.ExitWithError(fmt.Sprintf("--axes with --group requires exactly 1 column; got %d", len(axes)), nil)
		}
		groupCols := parser.EffectiveGroupColumns(cfg)
		if len(groupCols) != 2 {
			shared.ExitWithError(fmt.Sprintf("--axes with --group requires exactly 2 group dimensions; got %d", len(groupCols)), nil)
		}
		groupSet := map[string]bool{}
		for _, g := range groupCols {
			groupSet[g] = true
		}
		if groupSet[axes[0].Source] {
			shared.ExitWithError(fmt.Sprintf("column '%s' cannot be in both --group and --axes", axes[0].Source), nil)
		}
		cfg.Axes = axes
		return cfg
	}

	axes, err := parser.ParseAxesFlag(axesRaw)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	cfg.Axes = axes
	return cfg
}

// NewCommand builds the `vizb scatter` cobra command.
func NewCommand() *cobra.Command {
	o := &Options{}
	cmd := &cobra.Command{
		Use:   "scatter [target]",
		Short: "Generate a scatter chart from data",
		Long:  "Generate an interactive scatter chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			o.LinearOptions.Validate()
			cli.ValidateScale(&o.Scale)

			var threeDVisualMap *bool
			if cmd.Flags().Changed("3d-visualmap") {
				v := o.ThreeDVisualMap
				threeDVisualMap = &v
			}

			cfg := scatterchart.Materialise(scatterchart.Flags{
				Swap:            o.Swap,
				Scale:           o.Scale,
				Sort:            o.Sort,
				ShowLabels:      o.ShowLabels,
				ThreeDRotate:    o.ThreeDRotate,
				ThreeD:          o.ThreeD,
				ThreeDVisualMap: threeDVisualMap,
				Stat:            o.Stat,
			}, nil)

			parserCfg := o.ParseConfig()
			axes := parser.GroupAxes(parserCfg)
			if len(parserCfg.Axes) > 0 && len(parserCfg.Group) == 0 {
				axes = parser.ValueAxes(parserCfg)
			}
			if err := shared.ValidateSwap(cfg.Swap, axes); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			cli.RunSingleChartWithConfig(cmd, args, o.CommonOptions, parserCfg, []config_charts.ChartConfig{cfg})
		},
	}
	o.Bind(cmd.Flags())
	return cmd
}
