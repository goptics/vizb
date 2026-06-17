package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/goptics/vizb/cmd/cli"
	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	heatmapchart "github.com/goptics/vizb/config/charts/heatmap"
	linechart "github.com/goptics/vizb/config/charts/line"
	piechart "github.com/goptics/vizb/config/charts/pie"
	radarchart "github.com/goptics/vizb/config/charts/radar"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

// uiOptions holds the flags for the ui/html subcommand.
type uiOptions struct {
	OutputFile string
	Charts     []string
	ChartSpecs []string
	DataURL    string
	Enable3D   bool
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
	uiCmd.Flags().StringSliceVarP(&uiOpts.Charts, "charts", "c", shared.DefaultChartTypes, "Chart types to bundle (bar, line, pie, heatmap, radar)")
	uiCmd.Flags().StringArrayVar(&uiOpts.ChartSpecs, "chart", nil, "Per-chart type settings override: <type>:<key>=<val>,... (repeatable)")
	uiCmd.Flags().BoolVar(&uiOpts.Enable3D, "3d", false, "Bundle the 3D renderer for --data-url (remote data shape is unknown at build time)")
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
		charts := uiOpts.Charts
		needs3D := uiOpts.Enable3D && shared.ChartsHave3DCapable(charts)
		htmlContent := template.GenerateRemoteUI(
			uiOpts.DataURL, charts, needs3D, template.VizbHTMLTemplate,
		)
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

	benches, err := cli.ParseDatasetFile(args[0])
	if err != nil {
		shared.ExitWithError("Failed to parse DataSet file: %v", err)
	}

	if len(benches) == 0 {
		shared.ExitWithError("No DataSet data found in file", nil)
	}

	// Determine the effective chart selection that drives chunk pruning. When -c
	// is given it overrides (and is written back into each dataset so the embedded
	// VIZB_DATA tabs match the bundled chunks); otherwise honour each dataset's
	// baked-in chart types (extracted from Settings in the new model).
	var charts []string
	if cmd.Flags().Changed("charts") {
		charts = uiOpts.Charts
	} else {
		charts = unionCharts(benches)
	}

	if cmd.Flags().Changed("charts") {
		for i := range benches {
			benches[i].Settings = filterSettings(benches[i].Settings, charts)
		}
	}

	if len(uiOpts.ChartSpecs) > 0 {
		// Collect the union of every active chart type across the input
		// datasets so --chart overrides can be validated against the actual
		// selection (matches the same rule as the root command).
		active := unionCharts(benches)
		overrides, err := shared.ParseOverrides(uiOpts.ChartSpecs, active, nil)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		for i := range benches {
			applyOverrides(&benches[i].Settings, overrides)
		}
	}

	needs3D := shared.ChartsHave3DCapable(charts) && anyDatasetHasZAxis(benches)

	jsonData, err := json.Marshal(benches)
	if err != nil {
		shared.ExitWithError("Failed to marshal DataSet data: %v", err)
	}

	htmlContent := template.GenerateUI(jsonData, charts, needs3D, template.VizbHTMLTemplate)
	if _, err := f.WriteString(htmlContent); err != nil {
		shared.ExitWithError("Failed to write output file: %v", err)
	}
	fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated UI successfully: %s", outFile)))
}

// filterSettings keeps only configs whose chart type is in the allowed list,
// preserving the original settings order.
func filterSettings(settings []config_charts.ChartConfig, allowed []string) []config_charts.ChartConfig {
	if len(allowed) == 0 {
		return settings
	}
	permitted := make(map[string]bool, len(allowed))
	for _, c := range allowed {
		permitted[c] = true
	}
	filtered := make([]config_charts.ChartConfig, 0, len(settings))
	for _, s := range settings {
		if permitted[s.ChartType()] {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// unionCharts collects the distinct chart types across datasets, preserving first
// appearance order, so every chart any input requests stays bundled.
func unionCharts(benches []shared.Dataset) []string {
	seen := make(map[string]bool)
	var charts []string
	for i := range benches {
		for _, c := range benches[i].Settings {
			if !seen[c.ChartType()] {
				seen[c.ChartType()] = true
				charts = append(charts, c.ChartType())
			}
		}
	}
	return charts
}

// applyOverrides mutates settings in place, applying any matching override
// from the per-chart map. The per-chart-type switch is required because each
// Config struct has a distinct concrete type, and the override values only
// apply when both sides are the same type. An override whose key doesn't match
// any existing setting is silently dropped (the user can't add a chart type
// via --chart in the ui subcommand — the chart list is locked to what's
// already in the input file).
func applyOverrides(settings *[]config_charts.ChartConfig, overrides map[string]config_charts.ChartConfig) {
	if len(overrides) == 0 {
		return
	}
	for _, s := range *settings {
		ov, ok := overrides[s.ChartType()]
		if !ok {
			continue
		}
		switch s := s.(type) {
		case *barchart.Config:
			if o, ok := ov.(*barchart.Config); ok {
				mergeBarConfig(s, o)
			}
		case *linechart.Config:
			if o, ok := ov.(*linechart.Config); ok {
				mergeLineConfig(s, o)
			}
		case *piechart.Config:
			if o, ok := ov.(*piechart.Config); ok {
				mergePieConfig(s, o)
			}
		case *heatmapchart.Config:
			if o, ok := ov.(*heatmapchart.Config); ok {
				mergeHeatmapConfig(s, o)
			}
		case *radarchart.Config:
			if o, ok := ov.(*radarchart.Config); ok {
				mergeRadarConfig(s, o)
			}
		}
	}
}

// mergeBarConfig copies the non-zero fields of `from` into `to`. Mirrors the
// override-merge logic in barchart.Materialise.
func mergeBarConfig(to, from *barchart.Config) {
	if from.Swap != "" {
		to.Swap = from.Swap
	}
	if from.Sort != nil {
		to.Sort = from.Sort
	}
	if from.Scale != "" {
		to.Scale = from.Scale
	}
	if from.ShowLabels != nil {
		to.ShowLabels = from.ShowLabels
	}
	if from.AutoRotate != nil {
		to.AutoRotate = from.AutoRotate
	}
}

// mergeLineConfig copies the non-zero fields of `from` into `to`. Mirrors the
// override-merge logic in linechart.Materialise.
func mergeLineConfig(to, from *linechart.Config) {
	if from.Swap != "" {
		to.Swap = from.Swap
	}
	if from.Sort != nil {
		to.Sort = from.Sort
	}
	if from.Scale != "" {
		to.Scale = from.Scale
	}
	if from.ShowLabels != nil {
		to.ShowLabels = from.ShowLabels
	}
	if from.AutoRotate != nil {
		to.AutoRotate = from.AutoRotate
	}
}

// mergePieConfig copies the non-zero fields of `from` into `to`. pie has no
// Scale / AutoRotate.
func mergePieConfig(to, from *piechart.Config) {
	if from.Swap != "" {
		to.Swap = from.Swap
	}
	if from.Sort != nil {
		to.Sort = from.Sort
	}
	if from.ShowLabels != nil {
		to.ShowLabels = from.ShowLabels
	}
}

// mergeHeatmapConfig copies the non-zero fields of `from` into `to`. heatmap
// has no Scale / AutoRotate.
func mergeHeatmapConfig(to, from *heatmapchart.Config) {
	if from.Swap != "" {
		to.Swap = from.Swap
	}
	if from.Sort != nil {
		to.Sort = from.Sort
	}
	if from.ShowLabels != nil {
		to.ShowLabels = from.ShowLabels
	}
}

// mergeRadarConfig copies the non-zero fields of `from` into `to`. radar has
// no Scale / AutoRotate.
func mergeRadarConfig(to, from *radarchart.Config) {
	if from.Swap != "" {
		to.Swap = from.Swap
	}
	if from.Sort != nil {
		to.Sort = from.Sort
	}
	if from.ShowLabels != nil {
		to.ShowLabels = from.ShowLabels
	}
}

// anyDatasetHasZAxis reports whether any dataset carries a z dimension.
func anyDatasetHasZAxis(benches []shared.Dataset) bool {
	for i := range benches {
		if shared.DatasetHasZAxis(&benches[i]) {
			return true
		}
	}
	return false
}

// validateAPIURL ensures a string is a valid http(s) URL.
func validateAPIURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("must be a valid http:// or https:// URL")
	}
	return nil
}
