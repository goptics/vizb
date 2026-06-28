// Package cli holds the shared command machinery used by the root command and
// every chart subcommand: the generic chart-command builder, the unified flag
// descriptor bag (replacing the former global shared.FlagState and the
// CommonOptions/LinearOptions/ChartOptions structs), and the one linear pipeline
// that turns input into a chart HTML/JSON file.
package cli

import (
	"slices"

	config_charts "github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

// ChartCommands builds one cobra subcommand per registered chart Spec. The set
// of subcommands and their flags is therefore derived entirely from the
// config/charts registry — adding a chart or a chart flag needs no change here.
func ChartCommands() []*cobra.Command {
	specs := config_charts.Specs()
	cmds := make([]*cobra.Command, 0, len(specs))
	for _, spec := range specs {
		cmds = append(cmds, newChartCommand(spec))
	}
	return cmds
}

// newChartCommand builds the `vizb <type>` command from a Spec. It binds the
// data flags plus the chart's own descriptors into one FlagBag, then on Run
// validates, builds the chart seed from the changed flags, materialises a single
// typed Config, and runs the linear pipeline.
func newChartCommand(spec config_charts.Spec) *cobra.Command {
	bag := NewFlagBag(append(slices.Clone(DataFlags), spec.Flags...))

	cmd := &cobra.Command{
		Use:   spec.Use,
		Short: spec.Short,
		Long:  spec.Long,
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			bag.Validate(cmd)

			seed := bag.ChartSeed(cmd)
			cfg, err := config_charts.Materialise(spec.Type, seed, nil)
			if err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			parseCfg := bag.ParseConfig()
			axes := parser.GroupAxes(parseCfg)
			if err := shared.ValidateSwap(cfg.SwapString(), axes); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			RunSingleChart(cmd, args, bag.Meta(), parseCfg, []config_charts.ChartConfig{cfg})
		},
	}

	bag.Bind(cmd.Flags())
	return cmd
}
