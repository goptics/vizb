package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/goptics/vizb/version"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vizb [target]",
	Short: "Generate benchmark charts from Go test benchmarks",
	Long: `A CLI tool that extends the functionality of 'go test -bench' with chart generation.
It takes benchmark text or json file as input and generates an interactive HTML chart application.`,
	Version: version.Version,
	Args:    cobra.ArbitraryArgs,
	Run:     runBenchmark,
}

// Execute runs the main command-line interface for vizb.
// It processes the command line arguments and executes the benchmark visualization workflow.
// This function is the main entry point called from main.go and handles cleanup of temporary files.
func Execute() {
	defer shared.TempFiles.RemoveAll()

	if err := rootCmd.Execute(); err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&shared.FlagState.Name, "name", "n", "Benchmarks", "Name of the benchmark")
	rootCmd.Flags().StringVarP(&shared.FlagState.Description, "description", "d", "", "Description of the benchmark")
	rootCmd.PersistentFlags().StringVarP(&shared.FlagState.OutputFile, "output", "o", "", "Output file name (.json for JSON, .html or other for HTML)")
	rootCmd.Flags().StringVarP(&shared.FlagState.MemUnit, "mem-unit", "m", "B", "Memory unit available: b, B, KB, MB, GB")
	rootCmd.Flags().StringVarP(&shared.FlagState.TimeUnit, "time-unit", "t", "ns", "Time unit available: ns, us, ms, s")
	rootCmd.Flags().StringVarP(&shared.FlagState.NumberUnit, "number-unit", "u", "", "Number unit available: K, M, B, T (default: as-is)")
	rootCmd.Flags().StringVarP(&shared.FlagState.GroupPattern, "group-pattern", "p", "x", "Pattern to extract grouping information from benchmark names")
	rootCmd.Flags().StringVarP(&shared.FlagState.GroupRegex, "group-regex", "r", "", "Regex pattern to extract grouping information from benchmark names")
	rootCmd.Flags().StringVarP(&shared.FlagState.Sort, "sort", "s", "", "Sort in asc or desc order (default: as-is)")
	rootCmd.Flags().StringSliceVarP(&shared.FlagState.Charts, "charts", "c", []string{"bar", "line", "pie"}, "Chart types to generate (bar, line, pie)")
	rootCmd.Flags().BoolVarP(&shared.FlagState.ShowLabels, "show-labels", "l", false, "Show labels on charts")
	rootCmd.Flags().StringVarP(&shared.FlagState.FilterRegex, "filter", "f", "", "Regex pattern to include only matching benchmark names")

	// Add a hook to validate flags after parsing
	cobra.OnInitialize(func() {
		utils.ApplyValidationRules(flagValidationRules)
	})
}

func runBenchmark(cmd *cobra.Command, args []string) {
	// Check if we're receiving data from stdin (piped input)
	stat, _ := os.Stdin.Stat()
	isStdinPiped := (stat.Mode() & os.ModeCharDevice) == 0

	// If no args provided and no piped input, show error
	if len(args) == 0 && !isStdinPiped {
		fmt.Fprintln(os.Stderr, "Error: no target provided and no piped input detected")
		cmd.Help()
		shared.OsExit(1)
	}

	// Default target name for piped input
	target := "stdin"

	if len(args) > 0 {
		target = args[0]
	}

	// Process the benchmark data
	if isStdinPiped {
		target = shared.MustCreateTempFile(shared.TempBenchFilePrefix, "out")
		shared.TempFiles.Store(target)

		writeStdinPipedInputs(target)
	} else {
		checkTargetFile(target)
	}

	// Generate the output file with charts or JSON
	generateOutputFile(target)
}

func writeStdinPipedInputs(tempfilePath string) {
	inputTempFile := shared.MustCreateFile(tempfilePath)
	defer inputTempFile.Close()

	// Use bufio to read stdin line by line in real-time
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(inputTempFile)

	// Create a progress bar manager
	benchmarkProgressManager := NewBenchmarkProgressManager(
		progressbar.NewOptions(-1,
			progressbar.OptionSetDescription(style.Info.Render("Processing benchmarks")),
			progressbar.OptionSetWidth(50),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionOnCompletion(func() { fmt.Println() }),
		),
	)

	// Process each line as it comes in
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// End of input
				break
			}
			shared.ExitWithError("Error reading from stdin", err)
		}

		// Write the line to the file
		if _, err := writer.WriteString(line); err != nil {
			shared.ExitWithError("Error writing to file", err)
		}

		benchmarkProgressManager.ProcessLine(line)
	}

	benchmarkProgressManager.Finish()

	writer.Flush()
	inputTempFile.Sync()
}

func checkTargetFile(filePath string) {
	fmt.Println(style.Info.Render(fmt.Sprintf("📊 Reading benchmark data from file: %s", filePath)))

	// Check if the target file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		shared.ExitWithError(fmt.Sprintf("Error: File '%s' does not exist", filePath), nil)
	}
}

func convertToBenchmark(filePath string) (benchmark *shared.Benchmark) {
	f := shared.MustOpenFile(filePath)
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		shared.ExitWithError("Failed to read file: %v", err)
	}

	if err := json.Unmarshal(content, &benchmark); err != nil {
		return nil
	}

	return benchmark
}

func generateOutputFile(filePath string) {
	outFile := resolveOutputFileName(shared.FlagState.OutputFile)
	// first try to convert to benchmark
	benchmark := convertToBenchmark(filePath)

	// if it fails, try to parse results from txt, or bench event json
	if benchmark == nil {
		filePath = preprocessInputFile(filePath)
		results := prepareBenchmarkData(filePath)
		benchmark = prepareBenchmarkFromParsedResults(results)
	}

	f := shared.MustCreateFile(outFile)

	defer f.Close()

	writeOutput(f, benchmark, inferFormatFromExtension(outFile))

	HandleOutputResult(f)
}

// inferFormatFromExtension returns the output format based on file extension
func inferFormatFromExtension(outFile string) string {
	switch ext := strings.ToLower(filepath.Ext(outFile)); ext {
	case ".json":
		return "json"
	default:
		return "html"
	}
}

// resolveOutputFileName decides final output file name and infers format from extension
func resolveOutputFileName(outFile string) string {
	// create a temp file if we need to print the output inside stdout (default to html)
	if outFile == "" {
		tmpFilePath := shared.MustCreateTempFile(shared.TempBenchFilePrefix, "html")
		shared.TempFiles.Store(tmpFilePath)
		return tmpFilePath
	}

	// if file doesn't have an extension, add the default html extension
	if filepath.Ext(outFile) == "" {
		outFile += ".html"
	}

	return outFile
}

// preprocessInputFile handles JSON → TXT conversion if needed
func preprocessInputFile(filePath string) string {
	if utils.IsBenchJSONFile(filePath) {
		return parser.ConvertJsonBenchToText(filePath)
	}

	return filePath
}

// prepareBenchmarkData parses benchmark results or exits on error
func prepareBenchmarkData(filePath string) []shared.BenchmarkData {
	data := parser.ParseBenchmarkData(filePath)

	if len(data) == 0 {
		shared.ExitWithError("No benchmark data found", nil)
	}

	return data
}

func prepareBenchmarkFromParsedResults(results []shared.BenchmarkData) *shared.Benchmark {
	benchmark := &shared.Benchmark{
		Name:        shared.FlagState.Name,
		Description: shared.FlagState.Description,
		Data:        results,
	}
	enableSorting := shared.FlagState.Sort != ""

	benchmark.CPU.Cores = shared.CPUCount
	benchmark.CPU.Name = shared.CPU
	benchmark.Arch = shared.Arch
	benchmark.OS = shared.OS
	benchmark.Pkg = shared.Pkg
	benchmark.Settings.Charts = shared.FlagState.Charts
	benchmark.Settings.Sort.Enabled = enableSorting

	if enableSorting {
		benchmark.Settings.Sort.Order = shared.FlagState.Sort
	} else {
		benchmark.Settings.Sort.Order = "asc"
	}

	benchmark.Settings.ShowLabels = shared.FlagState.ShowLabels

	return benchmark
}

// writeOutput writes results to file in required format
func writeOutput(f *os.File, benchmark *shared.Benchmark, format string) {
	switch format {
	case "html":
		fmt.Println(style.Info.Render("🔄 Generating Chart..."))

		jsonData, err := json.Marshal(benchmark)
		if err != nil {
			shared.ExitWithError("Failed to marshal benchmark data: %v", err)
		}

		htmlContent := template.GenerateHTMLBenchmarkUI(jsonData, template.VizbHTMLTemplate)
		if _, err := f.WriteString(htmlContent); err != nil {
			shared.ExitWithError("Failed to write output file: %v", err)
		}

		fmt.Println(style.Success.Render("🎉 Generated HTML chart successfully!"))

	case "json":
		fmt.Println(style.Info.Render("🔄 Generating JSON..."))
		bytes, err := json.Marshal(benchmark)
		if err != nil {
			shared.ExitWithError("Error marshaling benchmark data", err)
		}
		f.Write(bytes)
		fmt.Println(style.Success.Render("🎉 Generated JSON successfully!"))
	}
}

// HandleOutputResult manages printing or showing final result
func HandleOutputResult(f *os.File) {
	if shared.FlagState.OutputFile != "" {
		fmt.Println(style.Info.Render(fmt.Sprintf("📄 Output file: %s", f.Name())))
		return
	}

	content, err := os.ReadFile(f.Name())
	if err != nil {
		shared.ExitWithError("Error reading output file", err)
	}
	fmt.Print("\033[H\033[2J") // clear screen
	fmt.Println(string(content))
}
