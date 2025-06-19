package chart

// This package contains the chart generation functionality for vizb

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/goptics/vizb/pkg/templates"
	"github.com/goptics/vizb/shared"
)

// BenchmarkResult represents a parsed benchmark result
type BenchmarkResult struct {
	Workload    string
	Subject     string
	NsPerOp     float64
	BytesPerOp  float64
	AllocsPerOp uint64
}

// BenchEvent represents the structure of a JSON event from 'go test -bench . -json'
type BenchEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test,omitempty"`
	Output string `json:"Output,omitempty"`
}

var (
	// Match benchmark output lines with memory stats (when -benchmem is used)
	benchMemLineRe = regexp.MustCompile(`Benchmark[^\s]+\s+\d+\s+([\d\.]+)\s+ns/op\s+([\d\.]+)\s+B/op\s+([\d\.]+)\s+allocs/op`)
	// Match benchmark output lines without memory stats
	benchLineRe = regexp.MustCompile(`Benchmark[^\s]+\s+\d+\s+([\d\.]+)\s+ns/op`)
	// Define a list of colors to use for the subjects
	colorList = []string{
		"#5470C6", // Blue
		"#3BA272", // Green
		"#FC8452", // Orange
		"#73C0DE", // Light blue
		"#EE6666", // Red
		"#FAC858", // Yellow
		"#9A60B4", // Purple
		"#EA7CCC", // Pink
	}
)

func GenerateChartsFromFile(jsonPath string) (string, error) {
	results, hasMemStats, err := parseBenchmarkResults(jsonPath)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no benchmark results found")
	}

	// Create charts for each metric with converted units and truncated to 2 decimal places
	// Convert ns to ms (divide by 1,000,000) and truncate to 2 decimal places
	nsPerOpChart := createChart(fmt.Sprintf("Execution Time (%s/op)", shared.FlagState.TimeUnit), results, func(r BenchmarkResult) string {
		// Convert ns to ms and truncate to 2 decimal
		var time float64
		switch shared.FlagState.TimeUnit {
		case "s":
			time = r.NsPerOp / 1000000000
		case "ms":
			time = r.NsPerOp / 1000000
		case "us":
			time = r.NsPerOp / 1000
		default:
			time = r.NsPerOp
		}

		// Format without decimal places if it's a whole number
		if time == float64(int64(time)) {
			return fmt.Sprintf("%.0f", time)
		}

		return fmt.Sprintf("%.2f", time)
	})

	// Convert B to KB (divide by 1024) and truncate to 2 decimal places
	bytesPerOpChart := createChart(fmt.Sprintf("Memory Usage (%s/op)", shared.FlagState.MemUnit), results, func(r BenchmarkResult) string {
		// Convert B to KB and truncate to 2 decimal
		var memory float64
		switch shared.FlagState.MemUnit {
		case "b":
			memory = r.BytesPerOp * 8
		case "KB":
			memory = r.BytesPerOp / 1024 / 1024
		case "MB":
			memory = r.BytesPerOp / 1024 / 1024 / 1024
		case "GB":
			memory = r.BytesPerOp / 1024 / 1024 / 1024 / 1024
		default:
			memory = r.BytesPerOp // default B
		}

		// Format without decimal places if it's a whole number
		if memory == float64(int64(memory)) {
			return fmt.Sprintf("%.0f", memory)
		}

		return fmt.Sprintf("%.2f", memory)
	})

	// Truncate allocations to 2 decimal places
	allocsPerOpChart := createChart("Allocations (allocs/op)", results, func(r BenchmarkResult) string {
		return fmt.Sprintf("%d", r.AllocsPerOp)
	})

	// Write all charts to HTML file using quicktemplate
	f, err := os.Create(shared.FlagState.OutputFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Set initialization options for all charts
	initOpts := opts.Initialization{Theme: "light"}
	nsPerOpChart.SetGlobalOptions(charts.WithInitializationOpts(initOpts))

	// Only include memory charts if we have memory stats
	if hasMemStats {
		bytesPerOpChart.SetGlobalOptions(charts.WithInitializationOpts(initOpts))
		allocsPerOpChart.SetGlobalOptions(charts.WithInitializationOpts(initOpts))

		// Write the template to the file with all charts
		templates.WriteBenchmarkChart(f, nsPerOpChart, bytesPerOpChart, allocsPerOpChart)
	} else {
		// When no memory stats are available, only pass the time chart
		// The nil values will tell the template not to render those sections
		templates.WriteBenchmarkChart(f, nsPerOpChart, nil, nil)
	}

	return shared.FlagState.OutputFile, nil
}

func parseBenchmarkResults(jsonPath string) ([]BenchmarkResult, bool, error) {
	f, err := os.Open(jsonPath)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	var results []BenchmarkResult
	dec := json.NewDecoder(f)

	for {
		var ev BenchEvent
		if err := dec.Decode(&ev); err != nil {
			if err == io.EOF {
				break
			}
			return nil, false, err
		}

		// We're looking for output lines that contain benchmark results
		if ev.Action == "output" && strings.Contains(ev.Output, "ns/op") {
			// if output doesn't contains Benchmark throw error regarding that this is not a benchmark output and exit the process
			if !strings.Contains(ev.Output, "Benchmark") {
				fmt.Fprintln(os.Stderr, "Error: Invalid benchmark format. Output does not contain 'Benchmark'")
				fmt.Fprintln(os.Stderr, "Hint: Run 'go test -bench . -json' to generate the JSON output file")
				os.Exit(1)
			}
			// First try to match with memory stats
			hasMemStats := false
			var m []string

			if memMatch := benchMemLineRe.FindStringSubmatch(ev.Output); memMatch != nil {
				m = memMatch
				hasMemStats = true
			} else if basicMatch := benchLineRe.FindStringSubmatch(ev.Output); basicMatch != nil {
				m = basicMatch
			}

			if m != nil {
				// Extract the benchmark name from the output
				parts := strings.Fields(ev.Output)
				if len(parts) == 0 {
					continue
				}

				benchName := parts[0]

				nameParts := strings.Split(
					strings.TrimPrefix(benchName, "Benchmark"),
					shared.FlagState.Separator,
				)

				if len(nameParts) < 2 {
					// Exit with a clear error message about the separator issue
					fmt.Fprintf(os.Stderr, "Error: Invalid benchmark format. Could not extract workload and subject using separator '%s'\n", shared.FlagState.Separator)
					fmt.Fprintf(os.Stderr, "Hint: Use the -s/--separator flag to specify a different separator if your benchmark names use a different format\n")
					os.Exit(1)
				}

				workload, subject := nameParts[len(nameParts)-2], nameParts[len(nameParts)-1]

				// Remove CPU suffix from subject (e.g., "Subject-8" -> "Subject")
				if idx := strings.LastIndex(subject, "-"); idx > 0 {
					// Check if everything after the dash is a number
					if cpuCount, err := strconv.Atoi(subject[idx+1:]); err == nil {
						// Store CPU count in global bench state
						subject = subject[:idx]

						if shared.CPUCount == 0 {
							shared.CPUCount = cpuCount
						}
					}
				}

				// Parse metrics
				nsPerOp, _ := strconv.ParseFloat(m[1], 64)

				// Default values for memory stats
				bytesPerOp := 0.0
				var allocsPerOp uint64

				// If we have memory stats, parse them
				if hasMemStats && len(m) >= 4 {
					bytesPerOp, _ = strconv.ParseFloat(m[2], 64)
					allocsPerOp, _ = strconv.ParseUint(m[3], 10, 64)
				}

				// Debug print removed
				results = append(results, BenchmarkResult{
					Workload:    workload,
					Subject:     subject,
					NsPerOp:     nsPerOp,
					BytesPerOp:  bytesPerOp,
					AllocsPerOp: allocsPerOp,
				})
			}
		}
	}

	// Check if any result has memory stats
	hasMemoryStats := false
	for _, r := range results {
		if r.BytesPerOp > 0 || r.AllocsPerOp > 0 {
			hasMemoryStats = true
			break
		}
	}

	return results, hasMemoryStats, nil
}

func createChart(title string, results []BenchmarkResult, metricFn func(BenchmarkResult) string) *charts.Bar {
	// Group data by workload and subject
	data := make(map[string]map[string]string)
	for _, r := range results {
		if data[r.Workload] == nil {
			data[r.Workload] = make(map[string]string)
		}
		data[r.Workload][r.Subject] = metricFn(r)
	}

	// Create a new bar chart
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithColorsOpts(colorList),
		charts.WithYAxisOpts(opts.YAxis{
			SplitLine: &opts.SplitLine{Show: opts.Bool(true)},
			AxisLabel: &opts.AxisLabel{Show: opts.Bool(true)},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:         "3%",
			Right:        "5%",
			Bottom:       "3%",
			Top:          "10%",
			ContainLabel: opts.Bool(true),
		}),
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:  opts.Bool(true),
			Right: "10%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: opts.Bool(true),
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  opts.Bool(true),
					Type:  "png",
					Title: "Save as PNG",
				},
			},
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: "light",
		}),
	)

	// Get unique workloads and sort them for X-axis
	workloads := make([]string, 0, len(data))
	for w := range data {
		workloads = append(workloads, w)
	}
	sort.Strings(workloads)
	bar.SetXAxis(workloads)

	// Get unique subjects for series
	subjects := make(map[string]struct{})
	for _, m := range data {
		for s := range m {
			subjects[s] = struct{}{}
		}
	}
	subjectList := make([]string, 0, len(subjects))
	for s := range subjects {
		subjectList = append(subjectList, s)
	}
	sort.Strings(subjectList)

	// Assign colors to subjects dynamically
	colors := make(map[string]string)
	for i, subject := range subjectList {
		colorIndex := i % len(colorList) // Cycle through colors if we have more subjects than colors
		colors[subject] = colorList[colorIndex]
	}

	// Add a series for each subject
	for _, subject := range subjectList {
		values := make([]opts.BarData, len(workloads))
		for j, workload := range workloads {
			value, exists := data[workload][subject]
			if !exists {
				value = fmt.Sprintf("%d", 0)
			}
			values[j] = opts.BarData{Value: value}
		}

		// Get color for this subject - we should always have a color assigned
		color := colors[subject]

		// Add the series to the chart
		bar.AddSeries(subject, values,
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color: color,
			}),
			charts.WithLabelOpts(opts.Label{
				Show:      opts.Bool(true),
				Position:  "top",
				Formatter: "{c}",
				FontSize:  10,
			}),
			charts.WithBarChartOpts(opts.BarChart{
				Stack: "",
			}),
		)
	}

	return bar
}
