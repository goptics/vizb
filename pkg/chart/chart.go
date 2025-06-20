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
	Name        string
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

func prepareChartTitle(Name, chartTitle string) string {
	if Name != "" {
		return Name + " - " + chartTitle
	}

	return chartTitle
}

func GenerateChartsFromFile(jsonPath string) (string, error) {
	results, err := parseBenchmarkResults(jsonPath)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no benchmark results found")
	}

	// Group results by task name
	taskGroups := groupResultsByTask(results)
	BenchCharts := make([]shared.BenchCharts, 0, len(taskGroups))

	for Name, taskResults := range taskGroups {
		var nsPerOpChart, bytesPerOpChart, allocsPerOpChart *charts.Bar

		nsPerOpChart = createChart(prepareChartTitle(Name, fmt.Sprintf("Execution Time (%s/op)", shared.FlagState.TimeUnit)), taskResults, func(r BenchmarkResult) string {
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

		if shared.HasMemStats {
			bytesPerOpChart = createChart(prepareChartTitle(Name, fmt.Sprintf("Memory Usage (%s/op)", shared.FlagState.MemUnit)), taskResults, func(r BenchmarkResult) string {
				// Convert B to KB and truncate to 2 decimal
				var memory float64
				switch strings.ToLower(shared.FlagState.MemUnit) {
				case "b":
					memory = r.BytesPerOp * 8
				case "kb":
					memory = r.BytesPerOp / 1024
				case "mb":
					memory = r.BytesPerOp / (1024 * 1024)
				case "gb":
					memory = r.BytesPerOp / (1024 * 1024 * 1024)
				default:
					memory = r.BytesPerOp // default B
				}

				// Format without decimal places if it's a whole number
				if memory == float64(int64(memory)) {
					return fmt.Sprintf("%.0f", memory)
				}

				return fmt.Sprintf("%.2f", memory)
			})

			// Prepare title with unit if specified
			allocTitle := "Allocations (allocs/op)"

			if shared.FlagState.AllocUnit != "" {
				shared.FlagState.AllocUnit = strings.ToUpper(shared.FlagState.AllocUnit)
				allocTitle = fmt.Sprintf("Allocations (%s/op)", shared.FlagState.AllocUnit)
			}

			allocsPerOpChart = createChart(prepareChartTitle(Name, allocTitle), taskResults, func(r BenchmarkResult) string {
				var allocs float64 = float64(r.AllocsPerOp)

				// Convert based on the allocation unit flag
				switch shared.FlagState.AllocUnit {
				case "K":
					allocs = allocs / 1000
				case "M":
					allocs = allocs / 1000000
				case "B":
					allocs = allocs / 1000000000
				case "T":
					allocs = allocs / 1000000000000
				default:
					// Default: as-is, no conversion needed
					return fmt.Sprintf("%d", r.AllocsPerOp)
				}

				// Format without decimal places if it's a whole number
				if allocs == float64(int64(allocs)) {
					return fmt.Sprintf("%.0f", allocs)
				}

				return fmt.Sprintf("%.2f", allocs)
			})
		}

		// Add charts for this task
		BenchCharts = append(BenchCharts, shared.BenchCharts{
			Name:             Name,
			NsPerOpChart:     nsPerOpChart,
			BytesPerOpChart:  bytesPerOpChart,
			AllocsPerOpChart: allocsPerOpChart,
		})
	}

	// Write all charts to HTML file using quicktemplate
	f, err := os.Create(shared.FlagState.OutputFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Set initialization options for all charts
	initOpts := opts.Initialization{}
	for i := range BenchCharts {
		BenchCharts[i].NsPerOpChart.SetGlobalOptions(charts.WithInitializationOpts(initOpts))
		if shared.HasMemStats {
			BenchCharts[i].BytesPerOpChart.SetGlobalOptions(charts.WithInitializationOpts(initOpts))
			BenchCharts[i].AllocsPerOpChart.SetGlobalOptions(charts.WithInitializationOpts(initOpts))
		}
	}

	// Write the template to the file with all charts
	templates.WriteBenchmarkChart(f, BenchCharts)

	return shared.FlagState.OutputFile, nil
}

// Group results by task name
func groupResultsByTask(results []BenchmarkResult) map[string][]BenchmarkResult {
	taskGroups := make(map[string][]BenchmarkResult)

	// Group results by task name
	for _, result := range results {
		Name := result.Name

		taskGroups[Name] = append(taskGroups[Name], result)
	}

	return taskGroups
}

func parseBenchmarkResults(jsonPath string) (results []BenchmarkResult, e error) {
	f, err := os.Open(jsonPath)
	if err != nil {
		e = err
		return
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	for {
		var ev BenchEvent
		if err := dec.Decode(&ev); err != nil {
			if err == io.EOF {
				break
			}
			e = err
			return
		}

		// We're looking for output lines that contain benchmark results
		if ev.Action == "output" && strings.Contains(ev.Output, "ns/op") {
			var stats []string

			// First try to match with memory stats
			if memMatch := benchMemLineRe.FindStringSubmatch(ev.Output); memMatch != nil {
				stats = memMatch
				shared.HasMemStats = true
			} else if basicMatch := benchLineRe.FindStringSubmatch(ev.Output); basicMatch != nil {
				stats = basicMatch
			} else {
				continue
			}

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

			var workload, subject string
			partsLen := len(nameParts)

			switch {
			case partsLen == 1:
				subject = nameParts[0]
			case partsLen == 2:
				benchName, subject = nameParts[0], nameParts[1]
			default:
				benchName, workload, subject = nameParts[partsLen-3], nameParts[partsLen-2], nameParts[partsLen-1]
			}

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
			nsPerOp, _ := strconv.ParseFloat(stats[1], 64)

			// Default values for memory stats
			var bytesPerOp float64
			var allocsPerOp uint64

			// If we have memory stats, parse them
			if shared.HasMemStats && len(stats) >= 4 {
				bytesPerOp, _ = strconv.ParseFloat(stats[2], 64)
				allocsPerOp, _ = strconv.ParseUint(stats[3], 10, 64)
			}

			results = append(results, BenchmarkResult{
				Name:        benchName,
				Workload:    workload,
				Subject:     subject,
				NsPerOp:     nsPerOp,
				BytesPerOp:  bytesPerOp,
				AllocsPerOp: allocsPerOp,
			})
		}
	}

	return
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
			Right:        "3%",
			Bottom:       "3%",
			Top:          "15%",
			ContainLabel: opts.Bool(true),
		}),
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:    opts.Bool(true),
			Padding: []int{30, 0, 0, 0},
			Right:   "3%",
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
		)
	}

	return bar
}
