package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/goptics/vizb/pkg/chart"
	"github.com/goptics/vizb/pkg/chart/templates"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// Variables to allow mocking for testing
var osExit = os.Exit
var osTempCreate = os.CreateTemp

const tempBenchFilePrefix = "vizb-benchmark-"
const outputFilePrefix = "vizb-"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vizb [target]",
	Short: "Generate benchmark charts from Go test benchmarks",
	Long: `A CLI tool that extends the functionality of 'go test -bench' with chart generation.
It runs the benchmark command internally, captures the JSON output, and generates
an interactive HTML chart based on the results.`,
	Version: "v0.1.1",
	Args:    cobra.ArbitraryArgs,
	Run:     runBenchmark,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		osExit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&shared.FlagState.Name, "name", "n", "Benchmarks", "Name of the chart")
	rootCmd.Flags().StringVarP(&shared.FlagState.Description, "description", "d", "", "Description of the benchmark")
	rootCmd.Flags().StringVarP(&shared.FlagState.Separator, "separator", "s", "/", "Separator for grouping benchmark names")
	rootCmd.Flags().StringVarP(&shared.FlagState.OutputFile, "output", "o", "", "Output HTML file name")
	rootCmd.Flags().StringVarP(&shared.FlagState.Format, "format", "f", "html", "Output format (html, json)")
	rootCmd.Flags().StringVarP(&shared.FlagState.MemUnit, "mem-unit", "m", "B", "Memory unit available: b, B, KB, MB, GB")
	rootCmd.Flags().StringVarP(&shared.FlagState.TimeUnit, "time-unit", "t", "ns", "Time unit available: ns, us, ms, s")
	rootCmd.Flags().StringVarP(&shared.FlagState.AllocUnit, "alloc-unit", "a", "", "Allocation unit available: K, M, B, T (default: as-is)")

	// Add a hook to validate flags after parsing
	cobra.OnInitialize(validateFlags)
}

// validateFlags validates the flag values and sets defaults for invalid values
func validateFlags() {
	// Validate memory unit
	validMemUnits := []string{"b", "B", "kb", "mb", "gb"}

	if !slices.Contains(validMemUnits, strings.ToLower(shared.FlagState.MemUnit)) {
		fmt.Fprintf(os.Stderr, "Warning: Invalid memory unit '%s'. Using default 'B'\n", shared.FlagState.MemUnit)
		shared.FlagState.MemUnit = "B"
	}

	// Validate time unit
	validTimeUnits := map[string]bool{"ns": true, "us": true, "ms": true, "s": true}
	if _, valid := validTimeUnits[shared.FlagState.TimeUnit]; !valid {
		fmt.Fprintf(os.Stderr, "Warning: Invalid time unit '%s'. Using default 'ns'\n", shared.FlagState.TimeUnit)
		shared.FlagState.TimeUnit = "ns"
	}

	// Validate allocation unit
	if shared.FlagState.AllocUnit != "" {
		shared.FlagState.AllocUnit = strings.ToUpper(shared.FlagState.AllocUnit)

		validAllocUnits := []string{"K", "M", "B", "T"}
		if !slices.Contains(validAllocUnits, shared.FlagState.AllocUnit) {
			fmt.Fprintf(os.Stderr, "Warning: Invalid allocation unit '%s'. Using default (as-is)\n", shared.FlagState.AllocUnit)
			shared.FlagState.AllocUnit = ""
		}
	}

	// Validate format
	validFormats := []string{"html", "json"}
	shared.FlagState.Format = strings.ToLower(shared.FlagState.Format)

	if !slices.Contains(validFormats, shared.FlagState.Format) {
		fmt.Fprintf(os.Stderr, "Warning: Invalid format '%s'. Using default 'html'\n", shared.FlagState.Format)
		shared.FlagState.Format = "html"
	}
}

func runBenchmark(cmd *cobra.Command, args []string) {
	// Check if we're receiving data from stdin (piped input)
	stat, _ := os.Stdin.Stat()
	isStdinPiped := (stat.Mode() & os.ModeCharDevice) == 0

	// If no args provided and no piped input, show error
	if len(args) == 0 && !isStdinPiped {
		fmt.Fprintln(os.Stderr, "Error: no target provided and no piped input detected")
		cmd.Help()
		osExit(1)
	}

	// Default target name for piped input
	target := "stdin"

	if len(args) > 0 {
		target = args[0]
	}

	// Process the benchmark data
	if isStdinPiped {
		target = createTempFile(tempBenchFilePrefix, "json")
		defer os.Remove(target)
		writeStdinPipedInputs(target)
	} else {
		readTargetedJsonFile(target)
	}

	// Generate the output file with charts or JSON
	generateOutputFile(target)
}

func createTempFile(prefix, extension string) string {
	temp, err := osTempCreate("", fmt.Sprintf("%s*.%s", prefix, extension))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temporary file: %v\n", err)
		osExit(1)
	}
	defer temp.Close()

	return temp.Name()
}

func writeStdinPipedInputs(tempJsonPath string) {
	// Open the JSON file for writing
	jsonFile, err := os.Create(tempJsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening JSON file for writing: %v\n", err)
		osExit(1)
	}
	defer jsonFile.Close()

	// Use bufio to read stdin line by line in real-time
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(jsonFile)
	lineCount := 0
	benchmarkCount := 0
	var currentBenchName string

	// Create a progress bar
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Processing benchmarks"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionOnCompletion(func() { fmt.Println() }),
	)

	// Process each line as it comes in
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// End of input
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			osExit(1)
		}

		// Write the line to the file
		if _, err := writer.WriteString(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			osExit(1)
		}

		var ev shared.BenchEvent
		if err := json.Unmarshal([]byte(line), &ev); err == nil {
			// If this is a benchmark test event, update the current benchmark name
			if ev.Test != "" && strings.HasPrefix(ev.Test, "Benchmark") {
				currentBenchName = ev.Test
				// Only count sub-benchmarks (with a slash in the name)
				if strings.Contains(ev.Test, shared.FlagState.Separator) && strings.Contains(line, "ns/op") {
					benchmarkCount++
				}
			}

			// Update progress bar description with benchmark count and current benchmark name
			description := fmt.Sprintf("Running Benchmarks [%s] (%d completed)",
				currentBenchName, benchmarkCount)

			bar.Describe(description)
		} else {
			fmt.Fprintf(os.Stderr, "\n❌ Error: Input is not in proper JSON format.\n")
			osExit(1)
		}

		// Flush periodically to ensure data is written
		if lineCount%100 == 0 {
			writer.Flush()
		}

		lineCount++
	}

	// Final flush to ensure all data is written
	writer.Flush()
	jsonFile.Sync()
	fmt.Println()
}

func readTargetedJsonFile(jsonPath string) {
	// Check if the target file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist\n", jsonPath)
		osExit(1)
	}

	// Check if the file contains valid JSON
	srcFile, err := os.Open(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening source file: %v\n", err)
		osExit(1)
	}
	defer srcFile.Close()

	fmt.Printf("📊 Reading benchmark data from file: %s\n", jsonPath)

	// Read a small portion of the file to check if it's valid JSON
	scanner := bufio.NewScanner(srcFile)

	if scanner.Scan() {
		firstLine := scanner.Text()
		var ev shared.BenchEvent
		if err := json.Unmarshal([]byte(firstLine), &ev); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: Input file is not in proper JSON format.\n")
			osExit(1)
		}
	}

	srcFile.Seek(0, 0)
}

// generateOutputFile handles the parsing of benchmark results, output file creation,
// and generation of charts or JSON based on the specified format.
func generateOutputFile(jsonPath string) {
	// Determine output file name
	outFile := shared.FlagState.OutputFile

	if outFile == "" {
		tempOutFile, err := os.CreateTemp("", fmt.Sprintf("%s*.%s", outputFilePrefix, shared.FlagState.Format))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temporary output file: %v\n", err)
			osExit(1)
		}
		defer os.Remove(tempOutFile.Name())
		outFile = tempOutFile.Name()

		tempOutFile.Close() // Close it now, it will be reopened below
	} else if !strings.HasSuffix(strings.ToLower(outFile), fmt.Sprintf(".%s", shared.FlagState.Format)) {
		outFile = outFile + fmt.Sprintf(".%s", shared.FlagState.Format)
	}

	// Parse benchmark results
	results, err := parser.ParseBenchmarkResults(jsonPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing benchmark results: %v\n", err)
		osExit(1)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "❌ No benchmark results found\n")
		osExit(1)
	}

	// Create output file
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating output file: %v\n", err)
		osExit(1)
	}
	defer f.Close()

	// Generate content based on format
	switch shared.FlagState.Format {
	case "html":
		fmt.Println("🔄 Generating Chart...")
		// Write all charts to HTML file using quicktemplate
		charts := chart.GenerateHTMLCharts(results)
		templates.WriteBenchmarkChart(f, charts)
		fmt.Printf("🎉 Generated HTML chart successfully!\n")

	case "json":
		fmt.Println("🔄 Generating JSON...")
		// Write all charts to JSON file
		bytes, err := json.Marshal(results)

		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error marshaling results: %v\n", err)
			osExit(1)
		}

		f.Write(bytes)
		fmt.Printf("🎉 Generated JSON successfully!\n")
	}

	// Handle different output scenarios
	if shared.FlagState.OutputFile == "" {
		// We created a temporary file, so read its contents and display them
		content, err := os.ReadFile(f.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading output file: %v\n", err)
			osExit(1)
		}
		fmt.Print("\033[H\033[2J") // ANSI escape sequence to clear screen and move cursor to home position
		// Print the HTML content to stdout
		fmt.Println(string(content))
	} else {
		// Normal file output, print success messages
		fmt.Printf("📄 Output file: %s\n", f.Name())
	}
}
