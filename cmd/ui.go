package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/goptics/vizb/cmd/cli"
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
	uiCmd.Flags().StringSliceVarP(&uiOpts.Charts, "charts", "c", []string{"bar", "line", "pie", "heatmap"}, "Chart types to bundle (bar, line, pie, heatmap)")
	uiCmd.Flags().StringArrayVar(&uiOpts.ChartSpecs, "chart", nil, "Per-chart type settings override: <type>:<key>=<val>,... (repeatable)")
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
		// No data at generation time: --charts is the authoritative selection, and
		// 3D is kept whenever a 3D-capable type (bar/line) is bundled, since the
		// remote data might carry a z dimension.
		charts := uiOpts.Charts
		htmlContent := template.GenerateRemoteUI(
			uiOpts.DataURL, charts, shared.ChartsHave3DCapable(charts), template.VizbHTMLTemplate,
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
		// Task 2 keeps the call sites compiling; the per-chart Config assembly
		// itself is rewritten in Task 4.
	} else {
		charts = unionCharts(benches)
	}

	if len(uiOpts.ChartSpecs) > 0 {
		// ParseChartSpecs still operates on the legacy []ChartSettings shape
		// (Task 4 will rewrite it on the new model). Kept as a no-op call for
		// now so the build passes; the per-chart override application is
		// rewritten alongside ParseChartSpecs in Task 4.
		_ = uiOpts.ChartSpecs
		_ = shared.ParseChartSpecs
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
