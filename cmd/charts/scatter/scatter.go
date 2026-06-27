// Package scatter registers the `vizb scatter` subcommand: a scatter chart with the full
// set of linear flags plus --scale and --3d-rotate (3D-only).
package scatter

import (
	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() { cli.Register(NewCommand) }

// Options adds the scatter-specific flags (--scale, --3d-rotate) on top of the shared
// chart options. pie/heatmap/radar omit these, enforced at compile time.
type Options struct {
	cli.ChartOptions
	Scale           string
	Symbol          string
	SymbolSize      float64
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
	fs.StringVar(&o.Symbol, "symbol", "", "Marker symbol (ECharts built-in: circle, rect, roundRect, triangle, diamond, pin, arrow, none; or path:// / image:// / SVG path)")
	fs.Float64Var(&o.SymbolSize, "symbol-size", 0, "Marker size in pixels (overrides default sizing)")
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
			o.Validate()
			cli.ValidateScale(&o.Scale)

			var threeDVisualMap *bool
			if cmd.Flags().Changed("3d-visualmap") {
				v := o.ThreeDVisualMap
				threeDVisualMap = &v
			}
			var symbolSize *float64
			if cmd.Flags().Changed("symbol-size") {
				v := o.SymbolSize
				symbolSize = &v
			}
			cli.ValidateSymbolFlags(o.Symbol, symbolSize)

			cfg := scatterchart.Materialise(scatterchart.Flags{
				Swap:            o.Swap,
				Scale:           o.Scale,
				Sort:            o.Sort,
				ShowLabels:      o.ShowLabels,
				Symbol:          o.Symbol,
				SymbolSize:      symbolSize,
				ThreeDRotate:    o.ThreeDRotate,
				ThreeD:          o.ThreeD,
				ThreeDVisualMap: threeDVisualMap,
				Stat:            o.Stat,
			}, nil)

			parserCfg := o.ParseConfig()
			cli.RunSingleChartWithConfig(cmd, args, o.CommonOptions, parserCfg, []config_charts.ChartConfig{cfg})
		},
	}
	o.Bind(cmd.Flags())
	return cmd
}
