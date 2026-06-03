package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goptics/vizb/pkg/style"
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
	mergeCmd.Flags().StringVarP(&shared.FlagState.TagAxis, "tag-axis", "A", "n",
		"Where to inject tag: n (name), x (xAxis), y (yAxis), z (zAxis)")
}

func runMerge(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		shared.ExitWithError("No input files provided", nil)
	}

	files := collectJSONFiles(args)
	if len(files) == 0 {
		shared.ExitWithError("No valid files found to merge", nil)
	}

	benches := readBenchmarks(files)
	if len(benches) == 0 {
		shared.ExitWithError("No valid benchmark files processed", nil)
	}

	merged := shared.MergeBenchmarks(benches, shared.Dimension(shared.FlagState.TagAxis))
	writeMergeOutput(merged)
}

func collectJSONFiles(args []string) []string {
	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			logWarn("cannot access %s: %v", arg, err)
			continue
		}

		if info.IsDir() {
			found, err := filepath.Glob(filepath.Join(arg, "*.json"))
			if err != nil {
				logWarn("error scanning directory %s: %v", arg, err)
				continue
			}
			files = append(files, found...)
		} else {
			files = append(files, arg)
		}
	}
	return files
}

func readBenchmarks(files []string) []shared.Benchmark {
	var benches []shared.Benchmark
	for _, file := range files {
		parsed, err := parseBenchmarkFile(file)
		if err != nil {
			logWarn("%s: %v", file, err)
			continue
		}
		benches = append(benches, parsed...)
	}
	return benches
}

func parseBenchmarkFile(file string) ([]shared.Benchmark, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	trimmed := bytes.TrimLeft(content, " \t\r\n")
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	switch trimmed[0] {
	case '[':
		var benches []shared.Benchmark
		if err := json.Unmarshal(content, &benches); err != nil {
			return nil, fmt.Errorf("invalid benchmark array: %w", err)
		}
		return benches, nil
	case '{':
		var bench shared.Benchmark
		if err := json.Unmarshal(content, &bench); err != nil {
			return nil, fmt.Errorf("invalid benchmark object: %w", err)
		}
		return []shared.Benchmark{bench}, nil
	default:
		return nil, fmt.Errorf("not valid JSON")
	}
}

func writeMergeOutput(benches []shared.Benchmark) {
	jsonData, err := json.Marshal(benches)
	if err != nil {
		shared.ExitWithError("Failed to marshal merged benchmark data: %v", err)
	}

	outFile := shared.FlagState.OutputFile
	if outFile == "" {
		outFile = shared.MustCreateTempFile(shared.TempBenchFilePrefix, "json")
		shared.TempFiles.Store(outFile)
	} else if filepath.Ext(outFile) == "" {
		outFile += ".json"
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()
	defer HandleOutputResult(f)

	if _, err := f.Write(jsonData); err != nil {
		shared.ExitWithError("Failed to write JSON output: %v", err)
	}
	fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated merged JSON successfully: %s", outFile)))
}

func logWarn(format string, args ...interface{}) {
	msg := fmt.Sprintf("Warning: "+format, args...)
	fmt.Fprintln(os.Stderr, style.Warning.Render(msg))
}
