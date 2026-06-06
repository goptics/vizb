package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goptics/vizb/pkg/parser"
	_ "github.com/goptics/vizb/pkg/parser/csv"
	goparser "github.com/goptics/vizb/pkg/parser/golang"
	_ "github.com/goptics/vizb/pkg/parser/javascript"
	_ "github.com/goptics/vizb/pkg/parser/json"
	_ "github.com/goptics/vizb/pkg/parser/rust"
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
	Short: "Visualize benchmarks or tabular CSV/JSON data as interactive 4D charts",
	Long: `A CLI tool that turns benchmark output (Go, Rust, JavaScript) or any tabular
CSV/JSON data into an interactive, self-contained HTML chart application.
It reads a file or piped stdin, auto-detects the input format (override with --parser),
and renders bar, line, and pie charts you can explore in the browser.`,
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
	rootCmd.PersistentFlags().StringVarP(&shared.FlagState.OutputFile, "output", "o", "", "Output file path/name")
	rootCmd.Flags().StringVarP(&shared.FlagState.MemUnit, "mem-unit", "M", "B", "Memory unit available: b, B, KB, MB, GB")
	rootCmd.Flags().StringVarP(&shared.FlagState.TimeUnit, "time-unit", "T", "ns", "Time unit available: ns, us, ms, s")
	rootCmd.Flags().StringVarP(&shared.FlagState.NumberUnit, "number-unit", "N", "", "Number unit available: K, M, B, T (default: as-is)")
	rootCmd.Flags().StringVarP(&shared.FlagState.GroupPattern, "group-pattern", "p", "x", "Pattern to extract grouping information from data labels / series names")
	rootCmd.Flags().StringVarP(&shared.FlagState.GroupRegex, "group-regex", "r", "", "Regex pattern to extract grouping information from data labels / series names")
	rootCmd.Flags().StringVarP(&shared.FlagState.Sort, "sort", "s", "", "Sort in asc or desc order (default: as-is)")
	rootCmd.Flags().StringSliceVarP(&shared.FlagState.Charts, "charts", "c", []string{"bar", "line", "pie"}, "Chart types to generate (bar, line, pie)")
	rootCmd.Flags().StringSliceVarP(&shared.FlagState.Group, "group", "g", nil, "Column/field names merged (in flag order, '/'-joined) into the group name; parsed by -p/-r (csv/json parsers)")
	rootCmd.Flags().BoolVarP(&shared.FlagState.ShowLabels, "show-labels", "l", false, "Show labels on charts")
	rootCmd.Flags().StringVarP(&shared.FlagState.FilterRegex, "filter", "f", "", "Regex pattern to include only matching data labels / series names")
	rootCmd.Flags().StringVarP(&shared.FlagState.Scale, "scale", "S", "linear", "Scale type (linear, log)")
	rootCmd.Flags().StringVarP(&shared.FlagState.Tag, "tag", "t", "", "Tag/identifier for the benchmark")
	rootCmd.Flags().StringVarP(&shared.FlagState.Parser, "parser", "P", "auto", "Benchmark parser to use; 'auto' detects from input content (one of: auto, "+strings.Join(parser.AvailableParsers(), ", ")+")")

	// Add a hook to validate flags after parsing
	cobra.OnInitialize(func() {
		utils.ApplyValidationRules(flagValidationRules)
	})
}

func runBenchmark(cmd *cobra.Command, args []string) {
	stat, _ := os.Stdin.Stat()
	isStdinPiped := (stat.Mode() & os.ModeCharDevice) == 0

	var target string

	if len(args) > 0 {
		target = args[0]
		checkTargetFile(target)
	} else if isStdinPiped {
		target = shared.MustCreateTempFile(shared.TempBenchFilePrefix, "out")
		shared.TempFiles.Store(target)

		writeStdinPipedInputs(target)
	} else {
		cmd.Help()
		shared.OsExit(0)
	}

	// In 'auto' mode (the default), sniff the input content and surface the
	// auto-selected parser so the choice is never silent.
	if shared.FlagState.Parser == "auto" {
		detected := parser.DetectParser(target)
		shared.FlagState.Parser = detected
		fmt.Println(style.Info.Render("✨ Auto-detected parser: " + detected))
	}

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

func convertToDataset(filePath string) (benchmark *shared.Dataset) {
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
	benchmark := convertToDataset(filePath)

	// if it fails, try to parse results from txt, or bench event json
	if benchmark == nil {
		filePath = preprocessInputFile(filePath)
		results := prepareData(filePath)
		benchmark = prepareDatasetFromResults(results)
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
	if shared.FlagState.Parser == "go" && utils.IsBenchJSONFile(filePath) {
		return goparser.ConvertGoJsonBenchToText(filePath)
	}

	return filePath
}

// prepareData parses benchmark results or exits on error
func prepareData(filePath string) []shared.DataPoint {
	parseFn, err := parser.GetParser(shared.FlagState.Parser)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	data := parseFn(filePath)

	if len(data) == 0 {
		shared.ExitWithError("No benchmark data found", nil)
	}

	return data
}

func prepareDatasetFromResults(results []shared.DataPoint) *shared.Dataset {
	benchmark := &shared.Dataset{
		Name:        shared.FlagState.Name,
		Description: shared.FlagState.Description,
		Data:        results,
	}
	enableSorting := shared.FlagState.Sort != ""

	benchmark.CPU.Cores = shared.CPUCount
	benchmark.CPU.Name = strings.TrimSpace(shared.CPU)
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
	benchmark.Settings.Scale = shared.FlagState.Scale

	benchmark.Tag = shared.FlagState.Tag
	benchmark.Timestamp = time.Now().UTC().Format(time.RFC3339)

	return benchmark
}

// writeOutput writes results to file in required format
func writeOutput(f *os.File, benchmark *shared.Dataset, format string) {
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
