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
	Short: "Visualize dataSets or tabular CSV/JSON data as interactive 4D charts",
	Long: `A CLI tool that turns dataSet output (Go, Rust, JavaScript) or any tabular
CSV/JSON data into an interactive, self-contained HTML chart application.
It reads a file or piped stdin, auto-detects the input format (override with --parser),
and renders bar, line, pie, and heatmap charts you can explore in the browser.`,
	Version: version.Version,
	Args:    cobra.ArbitraryArgs,
	Run:     runBenchmark,
}

// Execute runs the main command-line interface for vizb.
// It processes the command line arguments and executes the dataSet visualization workflow.
// This function is the main entry point called from main.go and handles cleanup of temporary files.
func Execute() {
	defer shared.TempFiles.RemoveAll()

	if err := rootCmd.Execute(); err != nil {
		shared.ExitWithError(err.Error(), nil)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&shared.FlagState.Name, "name", "n", "Comparisons", "Name of the comparison")
	rootCmd.Flags().StringVarP(&shared.FlagState.Description, "description", "d", "", "Description of the comparison")
	rootCmd.PersistentFlags().StringVarP(&shared.FlagState.OutputFile, "output", "o", "", "Output file path/name")
	rootCmd.Flags().StringVarP(&shared.FlagState.MemUnit, "mem-unit", "M", "B", "Memory unit available: b, B, KB, MB, GB")
	rootCmd.Flags().StringVarP(&shared.FlagState.TimeUnit, "time-unit", "T", "ns", "Time unit available: ns, us, ms, s")
	rootCmd.Flags().StringVarP(&shared.FlagState.NumberUnit, "number-unit", "N", "", "Number unit available: K, M, B, T (default: as-is)")
	rootCmd.Flags().StringVarP(&shared.FlagState.GroupPattern, "group-pattern", "p", "x", "Pattern to extract grouping information from data labels / series names")
	rootCmd.Flags().StringVarP(&shared.FlagState.GroupRegex, "group-regex", "r", "", "Regex pattern to extract grouping information from data labels / series names")
	rootCmd.Flags().StringVarP(&shared.FlagState.Sort, "sort", "s", "", "Sort in asc or desc order (default: as-is)")
	rootCmd.Flags().StringSliceVarP(&shared.FlagState.Charts, "charts", "c", []string{"bar", "line", "pie", "heatmap", "radar"}, "Chart types to generate (bar, line, pie, heatmap, radar)")
	rootCmd.Flags().StringSliceVarP(&shared.FlagState.Group, "group", "g", nil, "Names each dimension in --group-pattern/regex order. csv/json: column/field names whose values feed the dimensions; benchmark parsers: human-readable labels for the name/x/y/z axes")
	rootCmd.Flags().BoolVarP(&shared.FlagState.ShowLabels, "show-labels", "l", false, "Show labels on charts")
	rootCmd.Flags().StringVarP(&shared.FlagState.FilterRegex, "filter", "f", "", "Regex pattern to include only matching data labels / series names")
	rootCmd.Flags().StringVarP(&shared.FlagState.Scale, "scale", "S", "linear", "Scale type (linear, log)")
	rootCmd.Flags().StringVarP(&shared.FlagState.Tag, "tag", "t", "", "Tag/identifier for the comparison")
	rootCmd.Flags().StringVarP(&shared.FlagState.Parser, "parser", "P", "auto", "Benchmark parser to use; 'auto' detects from input content (one of: auto, "+strings.Join(parser.AvailableParsers(), ", ")+")")
	rootCmd.Flags().StringArrayVar(&shared.FlagState.ChartSpecs, "chart", nil,
		"Per-chart settings override: <type>:<key>=<val>(,<key>=<val>)* or bare flags (labels, rotate). "+
			"Keys: swap, sort, scale, labels, rotate. E.g. --chart bar:swap=yxn,sort=asc --chart pie:labels")

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
	dataSetProgressManager := NewBenchmarkProgressManager(
		progressbar.NewOptions(-1,
			progressbar.OptionSetDescription(style.Info.Render("Processing data sets")),
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

		dataSetProgressManager.ProcessLine(line)
	}

	dataSetProgressManager.Finish()

	writer.Flush()
	inputTempFile.Sync()
}

func checkTargetFile(filePath string) {
	fmt.Println(style.Info.Render(fmt.Sprintf("🔎 Reading data from file: %s", filePath)))

	// Check if the target file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		shared.ExitWithError(fmt.Sprintf("Error: File '%s' does not exist", filePath), nil)
	}
}

func convertToDataset(filePath string) (dataSet *shared.Dataset) {
	f := shared.MustOpenFile(filePath)
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		shared.ExitWithError("Failed to read file: %v", err)
	}

	if err := json.Unmarshal(content, &dataSet); err != nil {
		return nil
	}
	shared.MigrateDataset(dataSet, content)
	return dataSet
}

func generateOutputFile(filePath string) {
	outFile := resolveOutputFileName(shared.FlagState.OutputFile)
	// first try to convert to dataSet
	dataSet := convertToDataset(filePath)

	// if it fails, try to parse results from txt, or bench event json
	if dataSet == nil {
		filePath = preprocessInputFile(filePath)
		results := prepareData(filePath)
		dataSet = prepareDatasetFromResults(results)
	}

	f := shared.MustCreateFile(outFile)

	defer f.Close()

	writeOutput(f, dataSet, inferFormatFromExtension(outFile))

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

// prepareData parses dataSet results or exits on error
func prepareData(filePath string) []shared.DataPoint {
	parseFn, err := parser.GetParser(shared.FlagState.Parser)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	fmt.Println(style.Info.Render("⚙️  Parsing data..."))
	data := parseFn(filePath)

	// CSV/JSON emit one DataPoint per row; when grouping is active, multiple rows
	// can share the same (name, xAxis, yAxis, zAxis) key. Collapse them by summing
	// so the output isn't a row-per-record dump (200k rows → a few thousand points).
	// Benchmark parsers are excluded: their count=N repeats share a key but must NOT
	// be summed (the UI averages those instead).
	if (shared.FlagState.Parser == "csv" || shared.FlagState.Parser == "json") && len(shared.FlagState.Group) > 0 {
		before := len(data)
		fmt.Println(style.Info.Render(fmt.Sprintf("🧮 Aggregating %d rows...", before)))
		data = shared.AggregateDataPoints(data)
		fmt.Println(style.Info.Render(fmt.Sprintf("✅ Aggregated into %d grouped data points", len(data))))
	}

	if len(data) == 0 {
		shared.ExitWithError("No dataSet data found", nil)
	}

	return data
}

func prepareDatasetFromResults(results []shared.DataPoint) *shared.Dataset {
	dataSet := &shared.Dataset{
		Name:        shared.FlagState.Name,
		Description: shared.FlagState.Description,
		Data:        results,
	}
	enableSorting := shared.FlagState.Sort != ""

	meta := shared.Meta{OS: shared.OS, Arch: shared.Arch, Pkg: shared.Pkg}
	if cpuName := strings.TrimSpace(shared.CPU); cpuName != "" || shared.CPUCount != 0 {
		meta.CPU = &shared.CPUInfo{Name: cpuName, Cores: shared.CPUCount}
	}
	if meta != (shared.Meta{}) {
		dataSet.Meta = &meta
	}
	dataSet.Settings.Charts = shared.FlagState.Charts
	dataSet.Settings.Sort.Enabled = enableSorting

	if enableSorting {
		dataSet.Settings.Sort.Order = shared.FlagState.Sort
	} else {
		dataSet.Settings.Sort.Order = "asc"
	}

	dataSet.Settings.ShowLabels = shared.FlagState.ShowLabels
	dataSet.Settings.Scale = shared.FlagState.Scale

	dataSet.Tag = shared.FlagState.Tag
	dataSet.Timestamp = time.Now().UTC().Format(time.RFC3339)

	dataSet.Settings.Axes = parser.GroupAxes()

	if len(shared.FlagState.ChartSpecs) > 0 {
		chartSettings, err := shared.ParseChartSpecs(
			shared.FlagState.ChartSpecs,
			shared.FlagState.Charts,
			dataSet.Settings.Axes,
		)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		dataSet.Settings.ChartSettings = chartSettings
	}

	return dataSet
}

// writeOutput writes results to file in required format
func writeOutput(f *os.File, dataSet *shared.Dataset, format string) {
	switch format {
	case "html":
		fmt.Println(style.Info.Render("🔄 Generating UI..."))

		jsonData, err := json.Marshal(dataSet)
		if err != nil {
			shared.ExitWithError("Failed to marshal dataSet data: %v", err)
		}

		htmlContent := template.GenerateUI(jsonData, dataSet.Settings.Charts, shared.DatasetNeeds3D(dataSet), template.VizbHTMLTemplate)
		if _, err := f.WriteString(htmlContent); err != nil {
			shared.ExitWithError("Failed to write output file: %v", err)
		}

		fmt.Println(style.Success.Render("🎉 Generated HTML UI successfully!"))

	case "json":
		fmt.Println(style.Info.Render("🔄 Generating JSON..."))
		bytes, err := json.Marshal(dataSet)
		if err != nil {
			shared.ExitWithError("Error marshaling dataSet data", err)
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
