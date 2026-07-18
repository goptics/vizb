package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/pkg/core"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/spf13/cobra"
)

// mergeOptions holds the flags for the merge subcommand.
type mergeOptions struct {
	OutputFile string
	TagAxis    string
}

var mergeOpts mergeOptions

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
	mergeCmd.Flags().StringVarP(&mergeOpts.OutputFile, "output", "o", "", "Output file path/name")
	mergeCmd.Flags().StringVarP(&mergeOpts.TagAxis, "tag-axis", "A", "n",
		"Where to inject tag: n (name), x (xAxis), y (yAxis), z (zAxis)")
}

func runMerge(cmd *cobra.Command, args []string) {
	utils.ApplyValidationRules([]utils.ValidationRule{{
		Label:    "inject dimension",
		Value:    &mergeOpts.TagAxis,
		ValidSet: []string{"n", "x", "y", "z"},
		Default:  "n",
	}})

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

	writeMergeOutput(mergeDatasets(dataSets, shared.Dimension(mergeOpts.TagAxis)))
}

func mergeDatasets(dataSets []shared.Dataset, dimension shared.Dimension) []shared.Dataset {
	merged, err := core.Merge(dataSets, dimension)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
	return merged
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
		parsed, err := cli.ParseDatasetFile(file)
		if err != nil {
			logWarn("%s: %v", file, err)
			continue
		}
		dataSets = append(dataSets, parsed...)
	}
	return dataSets
}

func writeMergeOutput(dataSets []shared.Dataset) {
	jsonData, err := json.Marshal(dataSets)
	if err != nil {
		shared.ExitWithError("Failed to marshal merged data set data: %v", err)
	}

	outFile := mergeOpts.OutputFile
	if outFile == "" {
		outFile = shared.MustCreateTempFile(shared.TempBenchFilePrefix, "json")
		shared.TempFiles.Store(outFile)
	} else if filepath.Ext(outFile) == "" {
		outFile += ".json"
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()
	defer cli.HandleOutputResult(f, mergeOpts.OutputFile)

	if _, err := f.Write(jsonData); err != nil {
		shared.ExitWithError("Failed to write JSON output: %v", err)
	}
	fmt.Println(style.Success.Render(fmt.Sprintf("🎉 Generated merged JSON successfully: %s", outFile)))
}

func logWarn(format string, args ...any) {
	msg := fmt.Sprintf("Warning: "+format, args...)
	fmt.Fprintln(os.Stderr, style.Warning.Render(msg))
}
