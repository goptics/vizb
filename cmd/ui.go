package cmd

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/goptics/vizb/cmd/cli"
	internal_charts "github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
	heatmapchart "github.com/goptics/vizb/internal/charts/heatmap"
	linechart "github.com/goptics/vizb/internal/charts/line"
	piechart "github.com/goptics/vizb/internal/charts/pie"
	radarchart "github.com/goptics/vizb/internal/charts/radar"
	"github.com/goptics/vizb/pkg/core"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/spf13/cobra"
)

// uiOptions holds the flags for the ui/html subcommand.
type uiOptions struct {
	OutputFile string
	Charts     []string
	ChartSpecs []string
	DataURL    string
	Enable3D   bool
	Stat       []string
}

var uiOpts uiOptions

var uiCmd = &cobra.Command{
	Use:     "ui [file]",
	Aliases: []string{"html"},
	Short:   "Generate the interactive HTML UI from a DataSet JSON file",
	Long: `Generate an interactive HTML chart from a DataSet JSON file.
The input file must be a valid vizb DataSet JSON (single object or array).

When --data-url is set, no input file is needed. The generated HTML will fetch
DataSet JSON from the provided URL at runtime instead of embedding it.
Note: the JSON host must serve Access-Control-Allow-Origin: * for file:// access.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runUI,
}

func init() {
	rootCmd.AddCommand(uiCmd)
	uiCmd.Flags().StringVarP(&uiOpts.OutputFile, "output", "o", "", "Output file path/name")
	uiCmd.Flags().StringVarP(&uiOpts.DataURL, "data-url", "U", "", "URL to fetch DataSet JSON from at runtime (no input file needed)")
	// --charts lets `vizb ui` prune chart chunks (incl. --data-url, where it's the
	// only source of the selection since the data is fetched at runtime).
	uiCmd.Flags().StringSliceVarP(&uiOpts.Charts, "charts", "c", shared.DefaultChartTypes, "Chart types to bundle (bar, line, pie, heatmap, radar, scatter)")
	uiCmd.Flags().StringArrayVar(&uiOpts.ChartSpecs, "chart", nil, "Per-chart type settings override: <type>:<key>=<val>,... (repeatable)")
	uiCmd.Flags().BoolVar(&uiOpts.Enable3D, "3d", false, "Bundle the 3D renderer for --data-url (remote data shape is unknown at build time)")
	cli.BindStatFlag(uiCmd.Flags(), &uiOpts.Stat)
}

func runUI(cmd *cobra.Command, args []string) {
	if uiOpts.DataURL != "" {
		if err := validateAPIURL(uiOpts.DataURL); err != nil {
			shared.ExitWithError(fmt.Sprintf("invalid data url '%s': %v", uiOpts.DataURL, err), nil)
		}
	}

	outFile := uiOpts.OutputFile
	if outFile == "" {
		outFile = cli.ResolveOutputFileName(outFile)
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()
	defer cli.HandleOutputResult(f, uiOpts.OutputFile)

	if uiOpts.DataURL != "" {
		// No data at generation time: --charts is the authoritative selection.
		// 3D is opt-in via --3d because the remote data shape is unknown.
		// Heatmap chunk ships by default (remote stat config unknown); pruned
		// only when --stat is given and excludes correlations.
		charts := uiOpts.Charts
		needs3D := uiOpts.Enable3D && shared.ChartsHave3DCapable(charts)
		statChanged := cmd.Flags().Changed("stat")
		if statChanged {
			validateStat(&uiOpts.Stat)
		}
		needsHeatmapChunk := !statChanged || shared.StatNeedsCorrelation(uiOpts.Stat)
		htmlContent, err := template.GenerateRemoteUI(
			uiOpts.DataURL, charts, needs3D, needsHeatmapChunk, template.VizbHTMLTemplate,
		)
		if err != nil {
			shared.ExitWithError("Failed to generate UI: %v", err)
		}
		if _, err := f.WriteString(htmlContent); err != nil {
			shared.ExitWithError("Failed to write output file: %v", err)
		}
		fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated HTML chart successfully: %s", outFile)))
		return
	}

	if len(args) == 0 {
		shared.ExitWithError("provide a DataSet JSON file or use --data-url <url>", nil)
		return
	}

	datasets, err := cli.ParseDatasetFile(args[0])
	if err != nil {
		shared.ExitWithError("Failed to parse DataSet file: %v", err)
	}

	if len(datasets) == 0 {
		shared.ExitWithError("No dataset found in file", nil)
	}

	// Determine the effective chart selection that drives chunk pruning. When -c
	// is given it overrides (and is written back into each dataset so the embedded
	// VIZB_DATA tabs match the bundled chunks); otherwise honour each dataset's
	// baked-in chart types (extracted from Settings in the new model).
	var charts []string
	if cmd.Flags().Changed("charts") {
		charts = uiOpts.Charts
	} else {
		charts = unionCharts(datasets)
	}

	if cmd.Flags().Changed("charts") {
		for i := range datasets {
			datasets[i].Settings = filterSettings(datasets[i].Settings, charts)
		}
	}

	if len(uiOpts.ChartSpecs) > 0 {
		// Collect the union of every active chart type across the input
		// datasets so --chart overrides can be validated against the actual
		// selection (matches the same rule as the root command).
		active := unionCharts(datasets)
		overrides, warnings, err := shared.ParseOverrides(uiOpts.ChartSpecs, active, nil)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, style.Warning.Render(w))
		}
		for i := range datasets {
			applyOverrides(&datasets[i].Settings, overrides)
		}
	}

	if cmd.Flags().Changed("stat") {
		validateStat(&uiOpts.Stat)
		statCfg := &shared.StatConfig{Enabled: true, Math: uiOpts.Stat}
		for i := range datasets {
			applyStatToSettings(datasets[i].Settings, statCfg)
		}
	}

	htmlContent := generateEmbeddedUI(datasets, charts)
	if _, err := f.WriteString(htmlContent); err != nil {
		shared.ExitWithError("Failed to write output file: %v", err)
	}
	fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated UI successfully: %s", outFile)))
}

func generateEmbeddedUI(datasets []shared.Dataset, charts []string) string {
	htmlContent, err := core.GenerateUI(datasets, charts)
	if err != nil {
		shared.ExitWithError("Failed to generate UI: %v", err)
	}
	return htmlContent
}

// filterSettings keeps only configs whose chart type is in the allowed list,
// preserving the original settings order.
func filterSettings(settings []internal_charts.ChartConfig, allowed []string) []internal_charts.ChartConfig {
	if len(allowed) == 0 {
		return settings
	}
	filtered := make([]internal_charts.ChartConfig, 0, len(settings))
	for _, s := range settings {
		if slices.Contains(allowed, s.ChartType()) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// unionCharts collects the distinct chart types across datasets, preserving first
// appearance order, so every chart any input requests stays bundled.
func unionCharts(datasets []shared.Dataset) []string {
	seen := make(map[string]bool)
	var charts []string
	for i := range datasets {
		for _, c := range datasets[i].Settings {
			if !seen[c.ChartType()] {
				seen[c.ChartType()] = true
				charts = append(charts, c.ChartType())
			}
		}
	}
	return charts
}

// applyOverrides merges sparse --chart overrides into existing settings via the
// same JSON omitempty round-trip Materialise uses. Overrides whose chart type
// is not already in the input file are silently dropped.
func applyOverrides(settings *[]internal_charts.ChartConfig, overrides map[string]internal_charts.ChartConfig) {
	for i, cfg := range *settings {
		ov, ok := overrides[cfg.ChartType()]
		if !ok {
			continue
		}
		raw, err := json.Marshal(cfg)
		if err != nil {
			shared.ExitWithError("Failed to marshal chart config: %v", err)
		}
		var merged map[string]any
		if err := json.Unmarshal(raw, &merged); err != nil {
			shared.ExitWithError("Failed to decode chart config: %v", err)
		}
		ob, err := json.Marshal(ov)
		if err != nil {
			shared.ExitWithError("Failed to marshal chart override: %v", err)
		}
		var om map[string]any
		if err := json.Unmarshal(ob, &om); err != nil {
			shared.ExitWithError("Failed to decode chart override: %v", err)
		}
		maps.Copy(merged, om)
		out, err := json.Marshal(merged)
		if err != nil {
			shared.ExitWithError("Failed to marshal merged chart config: %v", err)
		}
		decoded, err := internal_charts.Decode(cfg.ChartType(), out)
		if err != nil {
			shared.ExitWithError("Failed to decode merged chart config: %v", err)
		}
		(*settings)[i] = decoded
	}
}

// validateAPIURL ensures a string is a valid http(s) URL.
func validateAPIURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("must be a valid http:// or https:// URL")
	}
	return nil
}

func validateStat(stat *[]string) {
	utils.ApplyValidationRules([]utils.ValidationRule{{
		Label:        "stat",
		SliceValue:   stat,
		ValidSet:     shared.ValidStatMath,
		Normalizer:   strings.ToLower,
		SliceDefault: nil,
	}})
}

// applyStatToSettings sets the same stat config on every chart config in settings.
func applyStatToSettings(settings []internal_charts.ChartConfig, stat *shared.StatConfig) {
	for _, cfg := range settings {
		switch c := cfg.(type) {
		case *barchart.Config:
			c.Stat = stat
		case *linechart.Config:
			c.Stat = stat
		case *piechart.Config:
			c.Stat = stat
		case *heatmapchart.Config:
			c.Stat = stat
		case *radarchart.Config:
			c.Stat = stat
		default:
			// ponytail: new chart type — add a case above and wire c.Stat = stat
			panic(fmt.Sprintf("applyStatToSettings: unhandled chart type %T", c))
		}
	}
}
