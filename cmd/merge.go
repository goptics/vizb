package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
			fmt.Fprintf(os.Stderr, "Warning: cannot access %s: %v\n", arg, err)
			continue
		}

		if info.IsDir() {
			files, err := filepath.Glob(filepath.Join(arg, "*.json"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: error scanning directory %s: %v\n", arg, err)
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
			fmt.Fprintf(os.Stderr, "Warning: cannot read file %s: %v\n", file, err)
			continue
		}

		var bench shared.Benchmark
		if err := json.Unmarshal(content, &bench); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: file %s does not satisfy Benchmark struct (JSON parse error), skipping.\n", file)
			continue
		}

		// Basic validation: check if Data is present or if it looks like a benchmark
		// The user said: "if one of passes json file doen't satifies the struct it will shows just a warning"
		// Unmarshal might succeed even if fields are missing (zero values).
		// We can check if 'Data' is empty, but maybe that's valid?
		// Let's assume if Unmarshal succeeds, it's "satisfied" enough, unless we want to check specific fields.
		// Given the user's prompt, "satisfies the struct" usually implies type correctness which Unmarshal handles.

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

	htmlContent := template.GenerateHTMLBenchmarkUI(jsonData)
	if _, err := f.WriteString(htmlContent); err != nil {
		shared.ExitWithError("Failed to write output file: %v", err)
	}

	fmt.Printf("🎉 Generated merged chart successfully: %s\n", outFile)
}
