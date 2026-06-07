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
		htmlContent := template.GenerateRemoteUI(shared.FlagState.DataURL, template.VizbHTMLTemplate)
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

	jsonData, err := json.Marshal(benches)
	if err != nil {
		shared.ExitWithError("Failed to marshal DataSet data: %v", err)
	}

	htmlContent := template.GenerateUI(jsonData, template.VizbHTMLTemplate)
	if _, err := f.WriteString(htmlContent); err != nil {
		shared.ExitWithError("Failed to write output file: %v", err)
	}
	fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated HTML chart successfully: %s", outFile)))
}
