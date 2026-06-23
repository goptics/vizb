// Package cmd wires the root cobra command and registers its subcommands.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	heatmapchart "github.com/goptics/vizb/config/charts/heatmap"
	linechart "github.com/goptics/vizb/config/charts/line"
	piechart "github.com/goptics/vizb/config/charts/pie"
	radarchart "github.com/goptics/vizb/config/charts/radar"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/pkg/style"
	// Chart subcommands self-register into cli's registry via their init().
	_ "github.com/goptics/vizb/cmd/charts/bar"
	_ "github.com/goptics/vizb/cmd/charts/heatmap"
	_ "github.com/goptics/vizb/cmd/charts/line"
	_ "github.com/goptics/vizb/cmd/charts/pie"
	_ "github.com/goptics/vizb/cmd/charts/radar"
	_ "github.com/goptics/vizb/cmd/charts/scatter"
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

// defaultChartTypes and validChartTypes alias shared constants for the root
// --charts flag: bar/line/pie by default, all five accepted when explicit.
var (
	defaultChartTypes = shared.DefaultChartTypes
	validChartTypes   = shared.ValidChartTypes
)

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
	Short: "Tabular visualization engine — charts and stats from CSV, JSON, and benchmarks",
	Long: `A tabular visualization engine for CSV, JSON, and benchmark output.
Turns numeric rows into interactive charts and descriptive statistics in one
self-contained HTML file. Reads a file or piped stdin, auto-detects the input
format (override with --parser), and renders bar, line, scatter, pie, heatmap,
and radar charts you can explore in the browser.`,
	Version: version.Version,
	Args:    cobra.ArbitraryArgs,
	Run:     runBenchmark,
}

// Execute runs the main command-line interface for vizb.
func Execute() {
	defer shared.TempFiles.RemoveAll()
	os.Args = cli.RewriteStatArg(os.Args)

	if err := rootCmd.Execute(); err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
}

func init() {
	rootOpts.LinearOptions.Bind(rootCmd.Flags())
	rootCmd.Flags().StringSliceVarP(&rootOpts.Charts, "charts", "c", defaultChartTypes, "Chart types to generate (bar, line, scatter, pie, heatmap, radar)")
	rootCmd.Flags().StringArrayVar(&rootOpts.ChartSpecs, "chart", nil,
		"Per-chart settings override: <type>:<key>=<val>(,<key>=<val>)* or bare flags (labels, 3d-rotate, 3d). "+
			"Keys: swap, sort, scale, labels, 3d-rotate, 3d. E.g. --chart bar:swap=yxn,sort=asc --chart pie:labels")

	// Register the chart subcommands (bar/line/pie/heatmap/radar) from the registry.
	rootCmd.AddCommand(cli.Commands()...)
}

func runBenchmark(cmd *cobra.Command, args []string) {
	warnDeprecatedRootFlags(cmd)
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
				Stat:       rootOpts.Stat,
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
				Stat:       rootOpts.Stat,
			}
			var override *linechart.Config
			if o, ok := overrides["line"]; ok {
				override = o.(*linechart.Config)
			}
			cfg = linechart.Materialise(flags, override)
		case "scatter":
			flags := scatterchart.Flags{
				Scale:      "linear",
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
				Stat:       rootOpts.Stat,
			}
			var override *scatterchart.Config
			if o, ok := overrides["scatter"]; ok {
				override = o.(*scatterchart.Config)
			}
			cfg = scatterchart.Materialise(flags, override)
		case "pie":
			flags := piechart.Flags{
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
				Stat:       rootOpts.Stat,
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
				Stat:       rootOpts.Stat,
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
				Stat:       rootOpts.Stat,
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
		ValidSet:     validChartTypes,
		Normalizer:   strings.ToLower,
		SliceDefault: defaultChartTypes,
	}})
}

// warnDeprecatedRootFlags emits a stderr warning when the root command's global
// --sort or --show-labels flag is explicitly set, recommending the per-chart
// --chart override equivalent. The flags remain functional — this is a
// deprecation notice only, not a removal. Chart subcommands (vizb bar, etc.)
// have their own per-chart --sort/--show-labels and are NOT deprecated.
func warnDeprecatedRootFlags(cmd *cobra.Command) {
	if cmd.Flags().Changed("sort") {
		fmt.Fprintln(os.Stderr, style.Warning.Render(
			"Warning: --sort is deprecated on the root command; use --chart <type>:sort=<asc|desc> instead (e.g. --chart bar:sort=asc)"))
	}
	if cmd.Flags().Changed("show-labels") {
		fmt.Fprintln(os.Stderr, style.Warning.Render(
			"Warning: --show-labels is deprecated on the root command; use --chart <type>:labels instead (e.g. --chart pie:labels)"))
	}
}
