// Package cli holds the shared command machinery used by the root command and
// every chart subcommand: the generic chart-command builder, the unified flag
// descriptor bag (replacing the former global shared.FlagState and the
// CommonOptions/LinearOptions/ChartOptions structs), and the one linear pipeline
// that turns input into a chart HTML/JSON file.
package cli

import (
	"fmt"
	"slices"

	internal_charts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

// ChartMeta holds the cobra command metadata for a chart type. Set by
// cmd/charts/<c> init() via SetChartMeta, consumed by ChartCommands.
type ChartMeta struct {
	Type  string
	Use   string
	Short string
	Long  string
}

// ContinuousSelectLong is shared Long help for bar/line/scatter: value axes need
// solo --select; all-numeric no-flag input uses auto col-axis x.
func ContinuousSelectLong(chart string) string {
	return fmt.Sprintf(`Generate an interactive %s chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.

Continuous coordinates (value axes) require solo --select, e.g.:
  vizb %s data.csv --select x,y,z -o out.html
  vizb %s data.csv --select x,y,z,value --3d-visualmap -o out.html

All-numeric files with no flags use auto col-axis x (column names as series),
not continuous coordinates. Use -A / --col-axis for series-on-axis without --select.`, chart, chart, chart)
}

// chartMetas maps chart type to its cobra metadata. Populated by
// cmd/charts/<c> init() (blank-imported from cmd/root.go).
var chartMetas = map[string]ChartMeta{}

// SetChartMeta registers cobra command metadata for a chart type. Called
// from cmd/charts/<c> init().
func SetChartMeta(m ChartMeta) {
	chartMetas[m.Type] = m
}

// ChartCommands builds one cobra subcommand per registered chart Spec. The set
// of subcommands and their flags is derived from the config/charts registry
// and the cmd-side chart metadata — adding a chart needs no change here.
func ChartCommands() []*cobra.Command {
	specs := internal_charts.Specs()
	cmds := make([]*cobra.Command, 0, len(specs))
	for _, spec := range specs {
		meta, ok := chartMetas[spec.Type]
		if !ok {
			continue
		}
		cmds = append(cmds, newChartCommand(spec, meta))
	}
	return cmds
}

// newChartCommand builds the `vizb <type>` command from a Spec and ChartMeta.
// It binds the data flags plus the chart's own flag descriptors into one
// FlagBag, then on Run validates, builds the chart seed from the changed
// flags, materialises a single typed Config, and runs the linear pipeline.
func newChartCommand(spec internal_charts.Spec, meta ChartMeta) *cobra.Command {
	bag := NewFlagBag(append(slices.Clone(DataFlags), internal_charts.FlagsFor(meta.Type)...))

	cmd := &cobra.Command{
		Use:   meta.Use,
		Short: meta.Short,
		Long:  meta.Long,
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			bag.Validate(cmd)

			seed := bag.ChartSeed(cmd)
			cfg, err := internal_charts.Materialise(spec.Type, seed, nil)
			if err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			parseCfg := bag.ParseConfig()
			axes := parser.GroupAxes(parseCfg)
			if err := shared.ValidateSwap(cfg.SwapString(), axes); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}

			RunSingleChart(cmd, args, bag.Meta(), parseCfg, []internal_charts.ChartConfig{cfg})
		},
	}

	bag.Bind(cmd.Flags())
	return cmd
}
