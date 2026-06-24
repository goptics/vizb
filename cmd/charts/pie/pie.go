// Package pie registers the `vizb pie` subcommand: a pie chart. Pie data is
// non-linear, so --scale and --3d-rotate are intentionally absent (the Options
// struct simply doesn't carry them).
package pie

import (
	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	piechart "github.com/goptics/vizb/config/charts/pie"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

func init() { cli.Register(NewCommand) }

// Options carries only the shared chart flags; no --scale/--3d-rotate.
type Options struct {
	cli.ChartOptions
}

// NewCommand builds the `vizb pie` cobra command.
func NewCommand() *cobra.Command {
	o := &Options{}
	cmd := &cobra.Command{
		Use:   "pie [target]",
		Short: "Generate a pie chart from data",
		Long:  "Generate an interactive pie chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			o.Validate()

			cfg := piechart.Materialise(piechart.Flags{
				Swap:       o.Swap,
				Sort:       o.Sort,
				ShowLabels: o.ShowLabels,
				Stat:       o.Stat,
			}, nil)

			axes := parser.GroupAxes(o.ParseConfig())
			if err := shared.ValidateSwap(cfg.Swap, axes); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			cli.RunSingleChart(cmd, args, o.CommonOptions, []config_charts.ChartConfig{cfg})
		},
	}
	o.Bind(cmd.Flags())
	return cmd
}
