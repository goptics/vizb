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
// the multi-chart selection (--charts), global --scale, and the per-chart
// --chart override specs.
type rootOptions struct {
	cli.LinearOptions
	Charts     []string
	Scale      string
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
	rootCmd.Flags().StringVarP(&rootOpts.Scale, "scale", "S", "linear", "Scale type (linear, log)")
	rootCmd.Flags().StringArrayVar(&rootOpts.ChartSpecs, "chart", nil,
		"Per-chart settings override: <type>:<key>=<val>(,<key>=<val>)* or bare flags (labels, rotate). "+
			"Keys: swap, sort, scale, labels, rotate. E.g. --chart bar:swap=yxn,sort=asc --chart pie:labels")

	// Register the chart subcommands (bar/line/pie/heatmap/radar) from the registry.
	rootCmd.AddCommand(cli.Commands()...)
}

func runBenchmark(cmd *cobra.Command, args []string) {
	validateRootOptions()

	// Build a per-chart Config for every active chart type. The switch is
	// required because each chart's Materialise is typed to that chart's Flags
	// struct. Per-chart --chart specs (rootOpts.ChartSpecs) are not yet
	// applied here; the typed override path lands in the root-command
	// refactor (which rewires ParseChartSpecs -> ParseOverrides). For now the
	// global flags seed each Materialise call with `nil` override, matching
	// the behaviour Task 4 will refine.
	configs := make([]config_charts.ChartConfig, 0, len(rootOpts.Charts))
	for _, chartType := range rootOpts.Charts {
		var cfg config_charts.ChartConfig
		switch chartType {
		case "bar":
			c := barchart.Materialise(barchart.Flags{
				Scale:      rootOpts.Scale,
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}, nil)
			cfg = c
		case "line":
			c := linechart.Materialise(linechart.Flags{
				Scale:      rootOpts.Scale,
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}, nil)
			cfg = c
		case "pie":
			c := piechart.Materialise(piechart.Flags{
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}, nil)
			cfg = c
		case "heatmap":
			c := heatmapchart.Materialise(heatmapchart.Flags{
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}, nil)
			cfg = c
		case "radar":
			c := radarchart.Materialise(radarchart.Flags{
				Sort:       rootOpts.Sort,
				ShowLabels: rootOpts.ShowLabels,
			}, nil)
			cfg = c
		}
		configs = append(configs, cfg)
	}

	// applyOnPassthrough is false: the root command preserves an existing Dataset
	// JSON as-is (matching historical behaviour).
	cli.RunLinear(cmd, args, rootOpts.CommonOptions, configs, false)
}

func validateRootOptions() {
	rootOpts.LinearOptions.Validate()
	cli.ValidateScale(&rootOpts.Scale)
	utils.ApplyValidationRules([]utils.ValidationRule{{
		Label:        "charts",
		SliceValue:   &rootOpts.Charts,
		ValidSet:     allChartTypes,
		Normalizer:   strings.ToLower,
		SliceDefault: allChartTypes,
	}})
}
