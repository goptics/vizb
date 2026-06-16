// Package radar registers the `vizb radar` subcommand: a radar chart. Radar data
// is non-linear, so --scale and --rotate are intentionally absent.
package radar

import (
	"github.com/goptics/vizb/cmd/cli"
	"github.com/spf13/cobra"
)

func init() { cli.Register(NewCommand) }

// Options carries only the shared chart flags; no --scale/--rotate.
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

			cli.RunSingleChart(cmd, args, o.CommonOptions, cli.LinearDefaults{
				Sort:       o.Sort,
				ShowLabels: o.ShowLabels,
			}, "radar", o.Swap, nil)
		},
	}
	o.ChartOptions.Bind(cmd.Flags())
	return cmd
}
