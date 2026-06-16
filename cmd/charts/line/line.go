// Package line registers the `vizb line` subcommand: a line chart with the full
// set of linear flags plus --scale and --rotate (3D-only).
package line

import (
	"github.com/goptics/vizb/cmd/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() { cli.Register(NewCommand) }

// Options adds the line-specific flags (--scale, --rotate) on top of the shared
// chart options. pie/heatmap/radar omit these, enforced at compile time.
type Options struct {
	cli.ChartOptions
	Scale      string
	AutoRotate bool
}

// Bind registers the shared chart flags plus --scale and --rotate.
func (o *Options) Bind(fs *pflag.FlagSet) {
	o.ChartOptions.Bind(fs)
	fs.StringVarP(&o.Scale, "scale", "S", "linear", "Scale type (linear, log)")
	fs.BoolVar(&o.AutoRotate, "rotate", false, "Auto-rotate the 3D chart")
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

			var rotate *bool
			if o.AutoRotate {
				t := true
				rotate = &t
			}

			cli.RunSingleChart(cmd, args, o.CommonOptions, cli.LinearDefaults{
				Sort:       o.Sort,
				Scale:      o.Scale,
				ShowLabels: o.ShowLabels,
			}, "line", o.Swap, rotate)
		},
	}
	o.Bind(cmd.Flags())
	return cmd
}
