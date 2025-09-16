package chart

// This package contains the chart generation functionality for vizb

import (
	"fmt"
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/goptics/vizb/shared"
)

var colorList = []string{

	"#E74C3C", // Red
	"#3498DB", // Blue
	"#2ECC71", // Green
	"#F39C12", // Orange
	"#9B59B6", // Purple
	"#1ABC9C", // Teal
	"#E67E22", // Dark Orange
	"#34495E", // Dark Blue Gray
	"#F1C40F", // Yellow
	"#E91E63", // Pink
	"#00BCD4", // Cyan
	"#8BC34A", // Light Green
	"#FF9800", // Amber
	"#673AB7", // Deep Purple
	"#009688", // Teal Green
	"#795548", // Brown
	"#607D8B", // Blue Gray
	"#FFC107", // Gold
	"#FF5722", // Deep Orange
	"#4CAF50", // Material Green
	"#2196F3", // Material Blue
	"#FFEB3B", // Bright Yellow
	"#9C27B0", // Material Purple
	"#00E676", // Neon Green
	"#FF1744", // Bright Red
	"#00B0FF", // Light Blue
	"#FFAB00", // Deep Amber
	"#AA00FF", // Electric Purple
	"#76FF03", // Lime
	"#FF6D00", // Vivid Orange
	"#18FFFF", // Aqua Cyan
	"#C6FF00", // Electric Lime
	"#FF3D00", // Red Orange
	"#651FFF", // Indigo
	"#00E5FF", // Light Cyan
	"#AEEA00", // Yellow Green
	"#DD2C00", // Deep Red
	"#3F51B5", // Indigo Blue
	"#4DB6AC", // Medium Aquamarine
	"#8D6E63", // Light Brown
	"#A1887F", // Warm Gray
	"#90A4AE", // Cool Gray
	"#BCAAA4", // Beige
	"#D7CCC8", // Light Beige
	"#F8BBD9", // Light Pink
	"#C8E6C9", // Mint Green
	"#DCEDC1", // Pale Green
	"#F0F4C3", // Pale Yellow
	"#FFF9C4", // Cream
	"#FFCCBC", // Peach
	"#D1C4E9", // Lavender
	"#C5CAE9", // Periwinkle
	"#BBDEFB", // Alice Blue
	"#B3E5FC", // Powder Blue
	"#B2EBF2", // Pale Turquoise
}

func prepareTitle(name, chartTitle string) string {
	if name != "" {
		return name + " - " + chartTitle
	}

	return chartTitle
}

func truncateFloat(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%.0f", f)
	}

	return fmt.Sprintf("%.2f", f)
}

// Group results by task name
func groupResultsByName(results []shared.BenchmarkResult) (map[string][]shared.BenchmarkResult, []string) {
	benchGroups := make(map[string][]shared.BenchmarkResult)
	groupNames := make([]string, 0)

	for _, result := range results {
		name := result.Name

		if _, has := benchGroups[name]; !has {
			groupNames = append(groupNames, name)
		}

		benchGroups[name] = append(benchGroups[name], result)
	}

	return benchGroups, groupNames
}

func countTotalSubjects(results []shared.BenchmarkResult) int {
	counter := make(map[string]bool)

	for _, r := range results {
		counter[r.Subject] = true
	}

	return len(counter)
}

func calculateLegendSpace(numItems int) string {
	itemsPerColumn := 15
	columns := (numItems + itemsPerColumn - 1) / itemsPerColumn

	return fmt.Sprintf("%d%%", min(15+(columns-1)*4, 35))
}

func createChart(title string, results []shared.BenchmarkResult, statIndex int) *charts.Bar {
	// Group data by workload and subject
	data := make(map[string]map[string]string)
	for _, r := range results {
		if data[r.Workload] == nil {
			data[r.Workload] = make(map[string]string)
		}
		data[r.Workload][r.Subject] = truncateFloat(r.Stats[statIndex].Value)
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
			Top:          calculateLegendSpace(countTotalSubjects(results)),
			ContainLabel: opts.Bool(true),
		}),
		charts.WithTitleOpts(opts.Title{
			Title: title,
			Top:   "1%",
			Left:  "2%",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:       opts.Bool(true),
			Top:        "7%", // Position for legend start
			Left:       "center",
			ItemWidth:  10,
			ItemHeight: 10,
			TextStyle: &opts.TextStyle{
				FontSize: 12,
			},
		}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show:  opts.Bool(true),
			Right: "2%",
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  opts.Bool(true),
					Type:  "png",
					Title: "Save",
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

// GenerateHTMLCharts creates interactive HTML charts from benchmark results.
// It groups benchmark results by name and generates charts for each stat type
// (execution time, memory usage, allocations) found in the results.
// Returns a slice of BenchCharts containing the generated chart data.
func GenerateHTMLCharts(results []shared.BenchmarkResult) []shared.BenchCharts {
	// Group results by task name
	benchGroups, groupNames := groupResultsByName(results)
	benchCharts := make([]shared.BenchCharts, 0, len(benchGroups))

	for _, name := range groupNames {
		benchResults := benchGroups[name]

		if len(benchResults) == 0 {
			continue
		}

		// Use the structure from the first result as a template
		firstResult := benchResults[0]
		charts := make([]*charts.Bar, 0, len(firstResult.Stats))

		// Create a chart for each stat type in the original order
		for idx, stat := range firstResult.Stats {
			// Determine chart title based on stat type and unit
			var chartTitle string
			if stat.Unit != "" {
				chartTitle = fmt.Sprintf("%s (%s/op)", stat.Type, stat.Unit)
			} else {
				chartTitle = fmt.Sprintf("%s/op", stat.Type)
			}

			// Create chart for this stat type
			chartBar := createChart(
				prepareTitle(name, chartTitle),
				benchResults,
				idx,
			)
			charts = append(charts, chartBar)
		}

		// Create only one BenchCharts entry per name group
		benchCharts = append(benchCharts, shared.BenchCharts{
			Name:   name,
			Charts: charts,
		})
	}

	return benchCharts
}
