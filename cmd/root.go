// Package cmd wires the root cobra command and registers its subcommands.
package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/goptics/vizb/cmd/cli"
	internal_charts "github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/pkg/style"

	// Chart configs self-register into the charts registry and cli metadata
	// via init() in cmd/charts/<c>; blank-importing them makes the registry
	// (and thus the subcommands and --chart key set) complete.
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

// rootFlags are the descriptors the root command binds: every data flag plus the
// shared chart-seed flags (sort/labels/stat) that seed every selected chart.
// Scale is per-chart only (bar/line); the root command has no global --scale.
func rootFlags() []flags.Flag {
	return append(slices.Clone(cli.DataFlags),
		internal_charts.SortFlag, internal_charts.LabelsFlag, internal_charts.StatFlag)
}

// rootBag binds and validates the root flags; rootCharts/rootChartSpecs are the
// root-only chart selection (--charts) and per-chart override specs (--chart).
var (
	rootBag        = cli.NewFlagBag(rootFlags())
	rootCharts     []string
	rootChartSpecs []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vizb [target]",
	Short: "Tabular visualization engine — charts and stats from CSV, JSON, and benchmarks",
	Long: `A tabular visualization engine for CSV, JSON, and benchmark output.
Turns numeric columns into interactive charts and descriptive statistics in one
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
	rootBag.Bind(rootCmd.Flags())
	rootCmd.Flags().StringSliceVarP(&rootCharts, "charts", "c", defaultChartTypes, "Chart types to generate (bar, line, scatter, pie, heatmap, radar)")
	rootCmd.Flags().StringArrayVar(&rootChartSpecs, "chart", nil,
		"Per-chart settings override: <type>:<key>=<val>(,<key>=<val>)* or bare flags (labels, 3d-rotate, 3d). "+
			"Keys: swap, sort, scale, labels, 3d-rotate, 3d, symbol, symbol-size. E.g. --chart bar:swap=yxn,sort=asc --chart scatter:symbol=diamond,symbol-size=12")

	// Build the chart subcommands (bar/line/pie/heatmap/radar/scatter) from the
	// config/charts registry.
	rootCmd.AddCommand(cli.ChartCommands()...)
}

func runBenchmark(cmd *cobra.Command, args []string) {
	warnDeprecatedRootFlags(cmd)
	validateRootOptions(cmd)

	// Parse the per-chart --chart overrides into typed Configs. Unknown chart
	// types or out-of-range values are CLI errors; keys valid for another chart
	// type are dropped with a warning.
	overrides, warnings, err := shared.ParseOverrides(rootChartSpecs, rootCharts, nil)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, style.Warning.Render(w))
	}

	// The root command's chart-seed flags (sort/labels/stat) seed every chart;
	// the per-chart --chart override (when present) wins over the seed. Scale is
	// per-chart only, so it is not seeded here — Materialise applies the "linear"
	// default from the chart's own ScaleFlag.
	seed := rootBag.ChartSeed(cmd)

	configs := make([]internal_charts.ChartConfig, 0, len(rootCharts))
	for _, chartType := range rootCharts {
		cfg, err := internal_charts.Materialise(chartType, seed, overrides[chartType])
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		configs = append(configs, cfg)
	}

	// applyOnPassthrough is false: the root command preserves an existing Dataset
	// JSON as-is (matching historical behaviour).
	cli.RunLinear(cmd, args, rootBag.Meta(), rootBag.ParseConfig(), configs, false)
}

func validateRootOptions(cmd *cobra.Command) {
	rootBag.Validate(cmd)
	utils.ApplyValidationRules([]utils.ValidationRule{{
		Label:        "charts",
		SliceValue:   &rootCharts,
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
