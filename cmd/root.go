package cmd

import (
	"strings"

	"github.com/goptics/vizb/cmd/cli"
	// Chart subcommands self-register into cli's registry via their init().
	_ "github.com/goptics/vizb/cmd/charts/bar"
	_ "github.com/goptics/vizb/cmd/charts/heatmap"
	_ "github.com/goptics/vizb/cmd/charts/line"
	_ "github.com/goptics/vizb/cmd/charts/pie"
	_ "github.com/goptics/vizb/cmd/charts/radar"
	"github.com/goptics/vizb/pkg/parser"

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
	Short: "Visualize datasets as interactive charts",
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

	cfg := rootOpts.ParseConfig()
	axes := parser.GroupAxes(cfg)

	specs, err := shared.ParseChartSpecs(rootOpts.ChartSpecs, rootOpts.Charts, axes)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	selections := cli.SelectionsFromCharts(rootOpts.Charts, specs)
	defaults := cli.LinearDefaults{
		Sort: rootOpts.Sort,
		// Scale is per-chart only (bar/line). The dataset-level default stays
		// "linear"; the UI falls back to it for charts without a per-chart override.
		Scale:      "linear",
		ShowLabels: rootOpts.ShowLabels,
	}

	// applyOnPassthrough is false: the root command preserves an existing Dataset
	// JSON as-is (matching historical behaviour).
	cli.RunLinear(cmd, args, rootOpts.CommonOptions, defaults, selections, false)
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
