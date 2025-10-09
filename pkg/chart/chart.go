package chart

// This package contains the chart generation functionality for vizb

import (
	"fmt"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/goptics/vizb/shared"
)

var colorList = []string{
	"#5470C6", // Blue
	"#3BA272", // Green
	"#FC8452", // Orange
	"#73C0DE", // Light blue
	"#EE6666", // Red
	"#FAC858", // Yellow
	"#9A60B4", // Purple
	"#EA7CCC", // Pink
	"#91CC75", // Lime
	"#FF9F7F", // Coral

	"#3E5A9E", // Navy Depth (Blue)
	"#2E7D32", // Forest Canopy (Green)
	"#EF6C00", // Burnt Sienna (Orange)
	"#7E57C2", // Amethyst Veil (Purple)
	"#F9A825", // Goldenrod Shine (Yellow)
	"#6A8ACF", // Steel Blue (Blue)
	"#4CAF50", // Emerald Leaf (Green)
	"#FF8F00", // Amber Flame (Orange)
	"#AB47BC", // Fuchsia Bloom (Pink)
	"#FFEB3B", // Lemon Zest (Yellow)

	"#2B4E72", // Midnight Slate (Blue)
	"#1B5E20", // Pine Shadow (Green)
	"#D84315", // Rust Ember (Red)
	"#512DA8", // Violet Shadow (Purple)
	"#F57F17", // Saffron Warmth (Yellow)
	"#4A90E2", // Cobalt Glow (Blue)
	"#66BB6A", // Verdant Bloom (Green)
	"#FF5722", // Coral Fire (Orange)
	"#BA68C8", // Orchid Haze (Purple)
	"#FFF176", // Banana Glow (Yellow)

	"#1E3A5F", // Deep Ocean (Blue)
	"#00695C", // Jade Depth (Green)
	"#BF360C", // Crimson Ember (Red)
	"#673AB7", // Plum Depth (Purple)
	"#C0CA33", // Chartreuse Edge (Lime)
	"#7FB3D5", // Azure Mist (Blue)
	"#81C784", // Moss Glow (Green)
	"#FFAB91", // Peach Sunset (Orange)
	"#E040FB", // Magenta Spark (Pink)
	"#DCE775", // Lime Radiance (Lime)

	"#335C8A", // Indigo Wave (Blue)
	"#388E3C", // Olive Ridge (Green)
	"#E64A19", // Tangerine Blaze (Orange)
	"#9575CD", // Lavender Dusk (Purple)
	"#78909C", // Slate Gray-Blue (Neutral)
	"#5C9EAD", // Teal Horizon (Blue-Green)
	"#AED581", // Sage Whisper (Green)
	"#FF7043", // Salmon Glow (Orange)
	"#F06292", // Rose Quartz (Pink)
	"#A1887F", // Taupe Earth (Neutral)
}

// to map subject with color index
var subjectColorMap = make(map[string]int)

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
	// Group data by workload and subject, preserving insertion order
	data := make(map[string]map[string]string)
	workloads := make([]string, 0)
	workloadSeen := make(map[string]bool)

	subjectList := make([]string, 0)
	subjectSeen := make(map[string]bool)

	for _, r := range results {
		// Track workload order
		if !workloadSeen[r.Workload] {
			workloads = append(workloads, r.Workload)
			workloadSeen[r.Workload] = true
		}

		// Track subject order
		if !subjectSeen[r.Subject] {
			subjectList = append(subjectList, r.Subject)
			subjectSeen[r.Subject] = true
		}

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
	bar.SetXAxis(workloads)

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
		color := colorList[subjectColorMap[subject]]

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
