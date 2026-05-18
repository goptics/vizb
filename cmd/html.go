package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

var htmlCmd = &cobra.Command{
	Use:   "html [file]",
	Short: "Generate HTML visualization from a benchmark JSON file",
	Long: `Generate an interactive HTML chart from a benchmark JSON file.
The input file must be a valid vizb benchmark JSON (single object or array).`,
	Args: cobra.ExactArgs(1),
	Run:  runHTML,
}

func init() {
	rootCmd.AddCommand(htmlCmd)
}

func runHTML(cmd *cobra.Command, args []string) {
	benches, err := parseBenchmarkFile(args[0])
	if err != nil {
		shared.ExitWithError("Failed to parse benchmark file: %v", err)
	}

	if len(benches) == 0 {
		shared.ExitWithError("No benchmark data found in file", nil)
	}

	jsonData, err := json.Marshal(benches)
	if err != nil {
		shared.ExitWithError("Failed to marshal benchmark data: %v", err)
	}

	outFile := shared.FlagState.OutputFile
	if outFile == "" {
		outFile = resolveOutputFileName(outFile)
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()
	defer HandleOutputResult(f)

	htmlContent := template.GenerateHTMLBenchmarkUI(jsonData, template.VizbHTMLTemplate)
	if _, err := f.WriteString(htmlContent); err != nil {
		shared.ExitWithError("Failed to write output file: %v", err)
	}
	fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated HTML chart successfully: %s", outFile)))
}
