// Package bar registers the `vizb bar` subcommand: a bar chart with the full
// set of linear flags plus --scale and --3d-rotate (3D-only).
package bar

import (
	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() { cli.Register(NewCommand) }

// Options adds the bar-specific flags (--scale, --3d-rotate) on top of the shared
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
	fs.BoolVar(&o.ThreeDRotate, "3d-rotate", false, "Auto-rotate the 3D scene (only applies when z-axis data is present)")
	fs.BoolVar(&o.ThreeD, "3d", false, "Enable value 3D for x+y data (y categories on depth, metric on height)")
	fs.BoolVar(&o.ThreeDVisualMap, "3d-visualmap", false, "Color 3D bars/lines by metric value (visualMap gradient)")
	fs.StringVar(&o.Axes, "axes", "", "csv/json only: plot 2-3 numeric columns as raw x,y[,z] coordinates; mutually exclusive with --group/--select")
}

// NewCommand builds the `vizb bar` cobra command.
func NewCommand() *cobra.Command {
	o := &Options{}
	cmd := &cobra.Command{
		Use:   "bar [target]",
		Short: "Generate a bar chart from data",
		Long:  "Generate an interactive bar chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			o.LinearOptions.Validate()
			cli.ValidateScale(&o.Scale)

			var threeDVisualMap *bool
			if cmd.Flags().Changed("3d-visualmap") {
				v := o.ThreeDVisualMap
				threeDVisualMap = &v
			}

			cfg := barchart.Materialise(barchart.Flags{
				Swap:            o.Swap,
				Scale:           o.Scale,
				Sort:            o.Sort,
				ShowLabels:      o.ShowLabels,
				ThreeDRotate:    o.ThreeDRotate,
				ThreeD:          o.ThreeD,
				ThreeDVisualMap: threeDVisualMap,
				Stat:            o.Stat,
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
