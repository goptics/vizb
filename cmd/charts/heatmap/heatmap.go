// Package heatmap registers the `vizb heatmap` subcommand: a heatmap chart.
// Heatmap folds z onto the legend and is always 2D, so --scale and --rotate are
// intentionally absent.
package heatmap

import (
	"github.com/goptics/vizb/cmd/cli"
	"github.com/spf13/cobra"
)

func init() { cli.Register(NewCommand) }

// Options carries only the shared chart flags; no --scale/--rotate.
type Options struct {
	cli.ChartOptions
}

// NewCommand builds the `vizb heatmap` cobra command.
func NewCommand() *cobra.Command {
	o := &Options{}
	cmd := &cobra.Command{
		Use:   "heatmap [target]",
		Short: "Generate a heatmap chart from data",
		Long:  "Generate an interactive heatmap chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			o.LinearOptions.Validate()

			cli.RunSingleChart(cmd, args, o.CommonOptions, cli.LinearDefaults{
				Sort:       o.Sort,
				ShowLabels: o.ShowLabels,
			}, "heatmap", o.Swap, nil)
		},
	}
	o.ChartOptions.Bind(cmd.Flags())
	return cmd
}
