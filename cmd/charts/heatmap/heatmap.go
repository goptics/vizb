// Package heatmap registers the `vizb heatmap` subcommand: a heatmap chart.
// Heatmap folds z onto the legend and is always 2D, so --scale and --rotate are
// intentionally absent.
package heatmap

import (
	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	heatmapchart "github.com/goptics/vizb/config/charts/heatmap"
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

			cfg := heatmapchart.Materialise(heatmapchart.Flags{
				Swap:       o.Swap,
				Sort:       o.Sort,
				ShowLabels: o.ShowLabels,
			}, nil)

			cli.RunSingleChart(cmd, args, o.CommonOptions, []config_charts.ChartConfig{cfg})
		},
	}
	o.ChartOptions.Bind(cmd.Flags())
	return cmd
}
