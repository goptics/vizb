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
	Short: "Merge multiple data set JSON files",
	Long: `Merge multiple data set JSON files into a single data set report.
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

	dataSets := readFiles(files)
	if len(dataSets) == 0 {
		shared.ExitWithError("No valid data set files processed", nil)
	}

	merged := shared.MergeDatasets(dataSets, shared.Dimension(shared.FlagState.TagAxis))
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

func readFiles(files []string) []shared.Dataset {
	var dataSets []shared.Dataset
	for _, file := range files {
		parsed, err := parseInputFile(file)
		if err != nil {
			logWarn("%s: %v", file, err)
			continue
		}
		dataSets = append(dataSets, parsed...)
	}
	return dataSets
}

func parseInputFile(file string) ([]shared.Dataset, error) {
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
		// Two-pass: decode each element from its own raw bytes so MigrateDataset
		// can recover the legacy top-level axisLabels field (lost after Unmarshal).
		var rawElems []json.RawMessage
		if err := json.Unmarshal(content, &rawElems); err != nil {
			return nil, fmt.Errorf("invalid data set array: %w", err)
		}
		dataSets := make([]shared.Dataset, 0, len(rawElems))
		for _, rawElem := range rawElems {
			var ds shared.Dataset
			if err := json.Unmarshal(rawElem, &ds); err != nil {
				return nil, fmt.Errorf("invalid data set array: %w", err)
			}
			shared.MigrateDataset(&ds, rawElem)
			dataSets = append(dataSets, ds)
		}
		return dataSets, nil
	case '{':
		var bench shared.Dataset
		if err := json.Unmarshal(content, &bench); err != nil {
			return nil, fmt.Errorf("invalid data set object: %w", err)
		}
		shared.MigrateDataset(&bench, content)
		return []shared.Dataset{bench}, nil
	default:
		return nil, fmt.Errorf("not valid JSON")
	}
}

func writeMergeOutput(dataSets []shared.Dataset) {
	jsonData, err := json.Marshal(dataSets)
	if err != nil {
		shared.ExitWithError("Failed to marshal merged data set data: %v", err)
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

func logWarn(format string, args ...any) {
	msg := fmt.Sprintf("Warning: "+format, args...)
	fmt.Fprintln(os.Stderr, style.Warning.Render(msg))
}
