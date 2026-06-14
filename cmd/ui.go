package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

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
	uiCmd.Flags().StringVarP(&shared.FlagState.DataURL, "data-url", "U", "", "URL to fetch DataSet JSON from at runtime (no input file needed)")
	// --charts is registered on rootCmd (not persistent), so register it here too
	// to let `vizb html` prune chart chunks (incl. --data-url, where it's the only
	// source of the selection since the data is fetched at runtime).
	uiCmd.Flags().StringSliceVarP(&shared.FlagState.Charts, "charts", "c", []string{"bar", "line", "pie", "heatmap"}, "Chart types to bundle (bar, line, pie, heatmap)")
}

func runUI(cmd *cobra.Command, args []string) {
	outFile := shared.FlagState.OutputFile
	if outFile == "" {
		outFile = resolveOutputFileName(outFile)
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()
	defer HandleOutputResult(f)

	if shared.FlagState.DataURL != "" {
		// No data at generation time: --charts is the authoritative selection, and
		// 3D is kept whenever a 3D-capable type (bar/line) is bundled, since the
		// remote data might carry a z dimension.
		charts := shared.FlagState.Charts
		htmlContent := template.GenerateRemoteUI(
			shared.FlagState.DataURL, charts, shared.ChartsHave3DCapable(charts), template.VizbHTMLTemplate,
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

	benches, err := parseInputFile(args[0])
	if err != nil {
		shared.ExitWithError("Failed to parse DataSet file: %v", err)
	}

	if len(benches) == 0 {
		shared.ExitWithError("No DataSet data found in file", nil)
	}

	// Determine the effective chart selection that drives chunk pruning. When -c
	// is given it overrides (and is written back into each dataset so the embedded
	// VIZB_DATA tabs match the bundled chunks); otherwise honour each dataset's
	// baked-in Settings.Charts.
	var charts []string
	if cmd.Flags().Changed("charts") {
		charts = shared.FlagState.Charts
		for i := range benches {
			benches[i].Settings.Charts = charts
		}
	} else {
		charts = unionCharts(benches)
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
		for _, c := range benches[i].Settings.Charts {
			if !seen[c] {
				seen[c] = true
				charts = append(charts, c)
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
