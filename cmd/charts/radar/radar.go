// Package radar registers the `vizb radar` subcommand: a radar chart. Radar data
// is non-linear, so --scale and --3d-rotate are intentionally absent.
package radar

import (
	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	radarchart "github.com/goptics/vizb/config/charts/radar"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

func init() { cli.Register(NewCommand) }

// Options carries only the shared chart flags; no --scale/--3d-rotate.
type Options struct {
	cli.ChartOptions
}

// NewCommand builds the `vizb radar` cobra command.
func NewCommand() *cobra.Command {
	o := &Options{}
	cmd := &cobra.Command{
		Use:   "radar [target]",
		Short: "Generate a radar chart from data",
		Long:  "Generate an interactive radar chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			o.LinearOptions.Validate()

			cfg := radarchart.Materialise(radarchart.Flags{
				Swap:       o.Swap,
				Sort:       o.Sort,
				ShowLabels: o.ShowLabels,
				Stat:       o.Stat,
			}, nil)

			axes := parser.GroupAxes(o.CommonOptions.ParseConfig())
			if err := shared.ValidateSwap(cfg.Swap, axes); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			cli.RunSingleChart(cmd, args, o.CommonOptions, []config_charts.ChartConfig{cfg})
		},
	}
	o.ChartOptions.Bind(cmd.Flags())
	return cmd
}
