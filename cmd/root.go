// Package cmd wires the root cobra command and registers its subcommands.
package cmd

import (
	"strings"

	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	heatmapchart "github.com/goptics/vizb/config/charts/heatmap"
	linechart "github.com/goptics/vizb/config/charts/line"
	piechart "github.com/goptics/vizb/config/charts/pie"
	radarchart "github.com/goptics/vizb/config/charts/radar"
	// Chart subcommands self-register into cli's registry via their init().
	_ "github.com/goptics/vizb/cmd/charts/bar"
	_ "github.com/goptics/vizb/cmd/charts/heatmap"
	_ "github.com/goptics/vizb/cmd/charts/line"
	_ "github.com/goptics/vizb/cmd/charts/pie"
	_ "github.com/goptics/vizb/cmd/charts/radar"
	// Parsers self-register into pkg/parser via their init().
	_ "github.com/goptics/vizb/pkg/parser/csv"
	_ "github.com/goptics/vizb/pkg/parser/golang"
	_ "github.com/goptics/vizb/pkg/parser/javascript"
	_ "github.com/goptics/vizb/pkg/parser/json"
	_ "github.com/goptics/vizb/pkg/parser/rust"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/goptics/vizb/version"
	"github.com/spf13/cobra"
)

// allChartTypes is the default/validation set for the root --charts flag.
var allChartTypes = []string{"bar", "line", "pie", "heatmap", "radar"}

// rootOptions holds the root command's flags: the shared linear data flags plus
// the multi-chart selection (--charts) and the per-chart --chart override specs.
// Scale is per-chart only (bar/line); the root command has no global --scale.
type rootOptions struct {
	cli.LinearOptions
	Charts     []string
	ChartSpecs []string
}

var rootOpts rootOptions

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vizb [target]",
	Short: "Visualize dataSets or tabular CSV/JSON data as interactive 4D charts",
	Long: `A CLI tool that turns dataSet output (Go, Rust, JavaScript) or any tabular
CSV/JSON data into an interactive, self-contained HTML chart application.
It reads a file or piped stdin, auto-detects the input format (override with --parser),
and renders bar, line, pie, heatmap, and radar charts you can explore in the browser.`,
	Version: version.Version,
	Args:    cobra.ArbitraryArgs,
	Run:     runBenchmark,
}

// Execute runs the main command-line interface for vizb.
func Execute() {
	defer shared.TempFiles.RemoveAll()

	if err := rootCmd.Execute(); err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
}

func init() {
	rootOpts.LinearOptions.Bind(rootCmd.Flags())
	rootCmd.Flags().StringSliceVarP(&rootOpts.Charts, "charts", "c", allChartTypes, "Chart types to generate (bar, line, pie, heatmap, radar)")
	rootCmd.Flags().StringArrayVar(&rootOpts.ChartSpecs, "chart", nil,
		"Per-chart settings override: <type>:<key>=<val>(,<key>=<val>)* or bare flags (labels, rotate). "+
			"Keys: swap, sort, scale, labels, rotate. E.g. --chart bar:swap=yxn,sort=asc --chart pie:labels")

	// Register the chart subcommands (bar/line/pie/heatmap/radar) from the registry.
	rootCmd.AddCommand(cli.Commands()...)
}

func runBenchmark(cmd *cobra.Command, args []string) {
	validateRootOptions()

	// Parse the per-chart --chart overrides into a typed map. An unknown
	// chart type or out-of-range value short-circuits with a CLI error.
	overrides, err := shared.ParseOverrides(rootOpts.ChartSpecs, rootOpts.Charts, nil)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	// Build a per-chart Config for every active chart type. The switch is
	// required because each chart's Materialise is typed to that chart's Flags
	// struct. The override from `overrides[chartType]` is passed as the
	// second arg so per-chart values (e.g. --chart bar:swap=yxn) win over the
	// global -s/--sort/-l/--show-labels defaults seeded via flagsFromDefaults.
	configs := make([]config_charts.ChartConfig, 0, len(rootOpts.Charts))
	for _, chartType := range rootOpts.Charts {
		var cfg config_charts.ChartConfig
		switch chartType {
		case "bar":
			flags := barchart.Flags{
				Scale:      "linear",
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}
			var override *barchart.Config
			if o, ok := overrides["bar"]; ok {
				override = o.(*barchart.Config)
			}
			cfg = barchart.Materialise(flags, override)
		case "line":
			flags := linechart.Flags{
				Scale:      "linear",
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}
			var override *linechart.Config
			if o, ok := overrides["line"]; ok {
				override = o.(*linechart.Config)
			}
			cfg = linechart.Materialise(flags, override)
		case "pie":
			flags := piechart.Flags{
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}
			var override *piechart.Config
			if o, ok := overrides["pie"]; ok {
				override = o.(*piechart.Config)
			}
			cfg = piechart.Materialise(flags, override)
		case "heatmap":
			flags := heatmapchart.Flags{
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}
			var override *heatmapchart.Config
			if o, ok := overrides["heatmap"]; ok {
				override = o.(*heatmapchart.Config)
			}
			cfg = heatmapchart.Materialise(flags, override)
		case "radar":
			flags := radarchart.Flags{
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}
			var override *radarchart.Config
			if o, ok := overrides["radar"]; ok {
				override = o.(*radarchart.Config)
			}
			cfg = radarchart.Materialise(flags, override)
		}
		configs = append(configs, cfg)
	}

	// applyOnPassthrough is false: the root command preserves an existing Dataset
	// JSON as-is (matching historical behaviour).
	cli.RunLinear(cmd, args, rootOpts.CommonOptions, configs, false)
}

func validateRootOptions() {
	rootOpts.LinearOptions.Validate()
	utils.ApplyValidationRules([]utils.ValidationRule{{
		Label:        "charts",
		SliceValue:   &rootOpts.Charts,
		ValidSet:     allChartTypes,
		Normalizer:   strings.ToLower,
		SliceDefault: allChartTypes,
	}})
}
