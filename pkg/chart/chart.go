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
	"#5470C6", // Blue
	"#3BA272", // Green
	"#FC8452", // Orange
	"#73C0DE", // Light blue
	"#EE6666", // Red
	"#FAC858", // Yellow
	"#9A60B4", // Purple
	"#EA7CCC", // Pink
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
