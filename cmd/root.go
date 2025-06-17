package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goptics/vizb/pkg/chart"
	"github.com/goptics/vizb/shared"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// Version information that will be populated at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	CommitSHA = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vizb [target]",
	Short: "Generate benchmark charts from Go test benchmarks",
	Long: `A CLI tool that extends the functionality of 'go test -bench' with chart generation.
It runs the benchmark command internally, captures the JSON output, and generates
an interactive HTML chart based on the results.`,
	Args: cobra.ArbitraryArgs,
	Run:  runBenchmark,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&shared.FlagState.Name, "name", "n", "Benchmarks", "Name of the chart")
	rootCmd.Flags().StringVarP(&shared.FlagState.Separator, "separator", "s", "/", "Separator for grouping benchmark names")
	rootCmd.Flags().StringVarP(&shared.FlagState.OutputFile, "output", "o", "", "Output HTML file name")
	rootCmd.Flags().StringVarP(&shared.FlagState.MemUnit, "mem-unit", "m", "B", "Memory unit available: b, B, KB, MB, GB")
	rootCmd.Flags().StringVarP(&shared.FlagState.TimeUnit, "time-unit", "t", "ns", "Time unit available: ns, us, ms, s")
	rootCmd.Flags().StringVarP(&shared.FlagState.Description, "description", "d", "", "Description of the benchmark")
	rootCmd.Flags().BoolVarP(&shared.FlagState.ShowVersion, "version", "v", false, "Show version information")
}

// BenchEvent represents the structure of a JSON event from 'go test -json'
type BenchEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test,omitempty"`
	Output string `json:"Output,omitempty"`
}

func runBenchmark(cmd *cobra.Command, args []string) {
	// Check if version flag is set
	if shared.FlagState.ShowVersion {
		fmt.Printf("vizb version %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", CommitSHA)
		return
	}

	// Check if we're receiving data from stdin (piped input)
	stat, _ := os.Stdin.Stat()
	isStdinPiped := (stat.Mode() & os.ModeCharDevice) == 0

	// If no args provided and no piped input, show error
	if len(args) == 0 && !isStdinPiped {
		fmt.Fprintln(os.Stderr, "Error: no target provided and no piped input detected")
		cmd.Help()
		os.Exit(1)
	}

	// Default target name for piped input
	target := "stdin"
	if len(args) > 0 {
		target = args[0]
	}

	// Variable to store the path to the JSON data file (either temp file or input file)
	var jsonFilePath string

	// Process the benchmark data
	if isStdinPiped {
		// Create a temporary file for the benchmark data
		tempFile, err := os.CreateTemp("", "benchmark-*.json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temporary file: %v\n", err)
			os.Exit(1)
		}

		jsonFilePath = tempFile.Name()
		defer os.Remove(jsonFilePath) // Clean up the temp file when done

		// Close the temp file first
		tempFile.Close()

		// Open the JSON file for writing
		jsonFile, err := os.Create(jsonFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening JSON file for writing: %v\n", err)
			os.Exit(1)
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
				os.Exit(1)
			}

			// Write the line to the file
			if _, err := writer.WriteString(line); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
				os.Exit(1)
			}

			var ev BenchEvent
			if err := json.Unmarshal([]byte(line), &ev); err == nil {
				// If this is a benchmark test event, update the current benchmark name
				if ev.Test != "" && strings.HasPrefix(ev.Test, "Benchmark") {
					currentBenchName = ev.Test
					// Only count sub-benchmarks (with a slash in the name)
					if strings.Contains(ev.Test, shared.FlagState.Separator) && strings.Contains(line, "=== RUN") {
						benchmarkCount++
					}
				}

				// Update progress bar description with benchmark count and current benchmark name
				description := fmt.Sprintf("Processing benchmarks [%s] (%d tests)",
					currentBenchName, benchmarkCount)

				bar.Describe(description)
			} else {
				fmt.Fprintf(os.Stderr, "\n❌ Error: Input is not in proper JSON format.\n")
				os.Exit(1)
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
	} else {
		// Using the target as a JSON file path
		jsonFilePath = target

		// Check if the target file exists
		if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist\n", jsonFilePath)
			os.Exit(1)
		}

		// Check if the file contains valid JSON
		srcFile, err := os.Open(jsonFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening source file: %v\n", err)
			os.Exit(1)
		}
		defer srcFile.Close()

		// Read a small portion of the file to check if it's valid JSON
		scanner := bufio.NewScanner(srcFile)

		if scanner.Scan() {
			firstLine := scanner.Text()
			var ev BenchEvent
			if err := json.Unmarshal([]byte(firstLine), &ev); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error: Input file is not in proper JSON format.\n")
				os.Exit(1)
			}
		}

		srcFile.Seek(0, 0)

		fmt.Printf("📊 Reading benchmark data from file: %s\n", target)
	}

	// Generate the chart using the chart package functionality
	fmt.Println("\n🔄 Generating chart...")

	// If no output file is specified, create a temporary one
	var tempOutputFile string
	if shared.FlagState.OutputFile == "" {
		// Create a temporary HTML file
		tempOutFile, err := os.CreateTemp("", "chart-*.html")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temporary output file: %v\n", err)
			os.Exit(1)
		}
		tempOutputFile = tempOutFile.Name()
		tempOutFile.Close() // Close it now, it will be reopened by the chart generator

		// Clear the terminal screen
		shared.FlagState.OutputFile = tempOutputFile
	} else if !strings.HasSuffix(strings.ToLower(shared.FlagState.OutputFile), ".html") {
		shared.FlagState.OutputFile = shared.FlagState.OutputFile + ".html"
	}

	// Import the chart package from the new location
	actualFilename, err := chart.GenerateChartsFromFile(jsonFilePath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error generating chart: %v\n", err)
		os.Exit(1)
	}

	// Handle different output scenarios
	if tempOutputFile != "" {
		// We created a temporary file, so read its contents and display them
		htmlContent, err := os.ReadFile(actualFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading chart HTML: %v\n", err)
			os.Exit(1)
		}
		fmt.Print("\033[H\033[2J") // ANSI escape sequence to clear screen and move cursor to home position
		// Print the HTML content to stdout
		fmt.Println(string(htmlContent))

		// Clean up the temporary file
		os.Remove(actualFilename)
	} else {
		// Normal file output, print success messages
		fmt.Printf("\n🎉 Chart generated successfully!\n")
		fmt.Printf("📄 Output file: %s\n", actualFilename)
		fmt.Printf("\nOpen the HTML file in your browser to view the benchmark results.\n")
	}
}
