package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goptics/vizb/pkg/chart"
	"github.com/goptics/vizb/pkg/chart/templates"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/goptics/vizb/version"
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
	Version: version.Version,
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
	rootCmd.Flags().StringVarP(&shared.FlagState.OutputFile, "output", "o", "", "Output HTML file name")
	rootCmd.Flags().StringVarP(&shared.FlagState.Format, "format", "f", "html", "Output format (html, json)")
	rootCmd.Flags().StringVarP(&shared.FlagState.MemUnit, "mem-unit", "m", "B", "Memory unit available: b, B, KB, MB, GB")
	rootCmd.Flags().StringVarP(&shared.FlagState.TimeUnit, "time-unit", "t", "ns", "Time unit available: ns, us, ms, s")
	rootCmd.Flags().StringVarP(&shared.FlagState.AllocUnit, "alloc-unit", "a", "", "Allocation unit available: K, M, B, T (default: as-is)")

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
		osExit(1)
	}

	// Default target name for piped input
	target := "stdin"

	if len(args) > 0 {
		target = args[0]
	}

	// Process the benchmark data
	if isStdinPiped {
		target = createTempFile(tempBenchFilePrefix, "out")
		defer os.Remove(target)
		writeStdinPipedInputs(target)
	} else {
		checkTargetFile(target)
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

func writeStdinPipedInputs(tempfilePath string) {
	// Open the JSON file for writing
	outFile, err := os.Create(tempfilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening JSON file for writing: %v\n", err)
		osExit(1)
	}
	defer outFile.Close()

	// Use bufio to read stdin line by line in real-time
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(outFile)

	// Create a progress bar manager
	benchmarkProgressManager := NewBenchmarkProgressManager(
		progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("Processing benchmarks"),
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
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			osExit(1)
		}

		// Write the line to the file
		if _, err := writer.WriteString(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			osExit(1)
		}

		benchmarkProgressManager.ProcessLine(line)
	}

	writer.Flush()
	outFile.Sync()
	fmt.Println()
}

func checkTargetFile(filePath string) {
	// Check if the target file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist\n", filePath)
		osExit(1)
	}

	// Check if the file contains valid JSON
	srcFile, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening source file: %v\n", err)
		osExit(1)
	}
	defer srcFile.Close()

	fmt.Printf("📊 Reading benchmark data from file: %s\n", filePath)

	ext := filepath.Ext(filePath)
	// Read a small portion of the file to check if it's valid JSON
	scanner := bufio.NewScanner(srcFile)

	if scanner.Scan() {
		firstLine := scanner.Text()

		switch ext {
		case "json":
			var ev shared.BenchEvent
			if err := json.Unmarshal([]byte(firstLine), &ev); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error: Input file is not in proper JSON format.\n")
				osExit(1)
			}
		}
	}

	srcFile.Seek(0, 0)
}

// generateOutputFile handles the parsing of benchmark results, output file creation,
// and generation of charts or JSON based on the specified format.
func generateOutputFile(filePath string) {
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

	// if a json file is provided then covert it to a new text file
	if utils.IsBenchJSONFile(filePath) {
		newBenchTxtPath, err := parser.ConvertJsonBenchToText(filePath)
		defer os.Remove(newBenchTxtPath)

		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error parsing json to txt: %v\n", err)
			osExit(1)
		}

		filePath = newBenchTxtPath
	}

	// Parse benchmark results
	results, err := parser.ParseBenchmarkResults(filePath)

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
		fmt.Fprintf(os.Stdout, "🔄 Generating Chart...\n")
		// Write all charts to HTML file using quicktemplate
		templates.WriteBenchmarkChart(f, chart.GenerateHTMLCharts(results))
		fmt.Fprintf(os.Stdout, "🎉 Generated HTML chart successfully!\n")

	case "json":
		fmt.Fprintf(os.Stdout, "🔄 Generating JSON...\n")
		// Write all charts to JSON file
		bytes, err := json.Marshal(results)

		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error marshaling results: %v\n", err)
			osExit(1)
		}

		f.Write(bytes)
		fmt.Fprintf(os.Stdout, "🎉 Generated JSON successfully!\n")
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
