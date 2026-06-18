// Package line registers the `vizb line` subcommand: a line chart with the full
// set of linear flags plus --scale and --3d-rotate (3D-only).
package line

import (
	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	linechart "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() { cli.Register(NewCommand) }

// Options adds the line-specific flags (--scale, --3d-rotate) on top of the shared
// chart options. pie/heatmap/radar omit these, enforced at compile time.
type Options struct {
	cli.ChartOptions
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
}

// NewCommand builds the `vizb line` cobra command.
func NewCommand() *cobra.Command {
	o := &Options{}
	cmd := &cobra.Command{
		Use:   "line [target]",
		Short: "Generate a line chart from data",
		Long:  "Generate an interactive line chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			o.LinearOptions.Validate()
			cli.ValidateScale(&o.Scale)

			var threeDVisualMap *bool
			if cmd.Flags().Changed("3d-visualmap") {
				v := o.ThreeDVisualMap
				threeDVisualMap = &v
			}

			cfg := linechart.Materialise(linechart.Flags{
				Swap:            o.Swap,
				Scale:           o.Scale,
				Sort:            o.Sort,
				ShowLabels:      o.ShowLabels,
				ThreeDRotate:    o.ThreeDRotate,
				ThreeD:          o.ThreeD,
				ThreeDVisualMap: threeDVisualMap,
			}, nil)

			axes := parser.GroupAxes(o.CommonOptions.ParseConfig())
			if err := shared.ValidateSwap(cfg.Swap, axes); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			cli.RunSingleChart(cmd, args, o.CommonOptions, []config_charts.ChartConfig{cfg})
		},
	}
	o.Bind(cmd.Flags())
	return cmd
}
