package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge [files/directories...]",
	Short: "Merge multiple benchmark JSON files",
	Long: `Merge multiple benchmark JSON files into a single benchmark report.
You can provide individual JSON files or directories containing JSON files.`,
	Run: runMerge,
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		shared.ExitWithError("No input files provided", nil)
	}

	var allFiles []string

	// Expand directories and collect files
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Warning.Render(fmt.Sprintf("Warning: cannot access %s: %v", arg, err)))
			continue
		}

		if info.IsDir() {
			files, err := filepath.Glob(filepath.Join(arg, "*.json"))
			if err != nil {
				fmt.Fprintln(os.Stderr, style.Warning.Render(fmt.Sprintf("Warning: error scanning directory %s: %v", arg, err)))
				continue
			}
			allFiles = append(allFiles, files...)
		} else {
			allFiles = append(allFiles, arg)
		}
	}

	if len(allFiles) == 0 {
		shared.ExitWithError("No valid files found to merge", nil)
	}

	var mergedBench []shared.Benchmark
	validFilesCount := 0

	for _, file := range allFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, style.Warning.Render(fmt.Sprintf("Warning: cannot read file %s: %v", file, err)))
			continue
		}

		var bench shared.Benchmark
		if err := json.Unmarshal(content, &bench); err != nil {
			fmt.Fprintln(os.Stderr, style.Warning.Render(fmt.Sprintf("Warning: file %s does not satisfy Benchmark struct (JSON parse error), skipping.", file)))
			continue
		}

		mergedBench = append(mergedBench, bench)
		validFilesCount++
	}

	if validFilesCount == 0 {
		shared.ExitWithError("No valid benchmark files processed", nil)
	}

	// Generate Output
	jsonData, err := json.Marshal(mergedBench)
	if err != nil {
		shared.ExitWithError("Failed to marshal merged benchmark data: %v", err)
	}

	// Determine output file
	outFile := shared.FlagState.OutputFile
	if outFile == "" {
		outFile = resolveOutputFileName(outFile, "html")
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()
	defer HandleOutputResult(f)

	htmlContent := template.GenerateHTMLBenchmarkUI(jsonData, template.VizbHTMLTemplate)
	if _, err := f.WriteString(htmlContent); err != nil {
		shared.ExitWithError("Failed to write output file: %v", err)
	}

	fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated merged chart successfully: %s", outFile)))
}
