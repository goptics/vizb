package templates

import (
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func TestBenchmarkChart(t *testing.T) {
	// Save original state
	originalName := shared.FlagState.Name
	originalDesc := shared.FlagState.Description
	originalCPUCount := shared.CPUCount

	// Reset after test
	defer func() {
		shared.FlagState.Name = originalName
		shared.FlagState.Description = originalDesc
		shared.CPUCount = originalCPUCount
	}()

	// Set test values
	shared.FlagState.Name = "Test Benchmark"
	shared.FlagState.Description = "Test Description"
	shared.CPUCount = 8

	tests := []struct {
		name            string
		benchCharts     []shared.BenchCharts
		wantContains    []string
		notWantContains []string
	}{
		{
			name: "Basic chart with name and description",
			benchCharts: []shared.BenchCharts{
				{
					Name:   "Test Chart",
					Charts: []*charts.Bar{createMockChart()},
				},
			},
			wantContains: []string{
				"<title>Test Benchmark</title>",
				"<h1>Test Benchmark (CPU: 8)</h1>",
				"<p style=\"text-align: center; margin-bottom: 20px;\">Test Description</p>",
				"Test Chart",     // Chart name should appear
				"class=\"item\"", // Chart container should be included
			},
			notWantContains: []string{
				"id=\"bench-sidebar\"", // Sidebar should not appear for single chart
			},
		},
		{
			name: "Multiple charts should show sidebar",
			benchCharts: []shared.BenchCharts{
				{
					Name:   "Chart 1",
					Charts: []*charts.Bar{createMockChart(), createMockChart()},
				},
				{
					Name:   "Chart 2",
					Charts: []*charts.Bar{createMockChart()},
				},
			},
			wantContains: []string{
				"id=\"bench-sidebar\"",             // Sidebar should appear for multiple charts
				"Chart 1",                          // First chart name
				"Chart 2",                          // Second chart name
				"class=\"bench-indicator active\"", // First indicator should be active
				"bench-section-0",                  // First section ID
				"bench-section-1",                  // Second section ID
			},
		},
		{
			name: "No chart name should not show in sidebar",
			benchCharts: []shared.BenchCharts{
				{
					Name:   "", // Empty name
					Charts: []*charts.Bar{createMockChart()},
				},
				{
					Name:   "Chart with Name",
					Charts: []*charts.Bar{createMockChart(), createMockChart()}, // Add an extra chart to exceed threshold
				},
			},
			wantContains: []string{
				"Chart with Name", // This should appear in sidebar
			},
			notWantContains: []string{
				"class=\"bench-indicator\"> </div>", // No empty indicators
			},
		},
		{
			name:        "Empty charts list",
			benchCharts: []shared.BenchCharts{},
			wantContains: []string{
				"<h1>Test Benchmark (CPU: 8)</h1>",
			},
			notWantContains: []string{
				"id=\"bench-sidebar\"",  // Sidebar should not appear for empty charts
				"<div class=\"chart\">", // No chart divs for empty charts list
			},
		},
		{
			name: "No description",
			benchCharts: []shared.BenchCharts{
				{
					Name:   "Test Chart",
					Charts: []*charts.Bar{createMockChart()},
				},
			},
			wantContains: []string{
				"<title>Test Benchmark</title>", // Title should still appear
			},
			notWantContains: []string{
				"<p style=\"text-align: center; margin-bottom: 20px;\">", // Description paragraph should not appear
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only set description for tests that need it
			if tt.name == "No description" {
				shared.FlagState.Description = ""
			} else {
				shared.FlagState.Description = "Test Description"
			}

			// Render template
			output := BenchmarkChart(tt.benchCharts)

			// Check for expected content
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain: "+want)
			}

			// Check for content that should not be present
			for _, notWant := range tt.notWantContains {
				assert.NotContains(t, output, notWant, "Output should not contain: "+notWant)
			}

			// Additional test for no description
			if tt.name == "No description" {
				assert.NotContains(t, output, "<p style=\"text-align: center; margin-bottom: 20px;\">",
					"Description paragraph should not appear when description is empty")
			}
		})
	}
}

func TestRenderChart(t *testing.T) {
	tests := []struct {
		name         string
		chart        *charts.Bar
		wantContains []string
	}{
		{
			name:  "Basic chart rendering",
			chart: createMockChart(),
			wantContains: []string{
				"<div class=\"container\">",
				"<div class=\"item\" id=\"",          // Chart has an ID
				"style=\"width:100%;height:500px;\"", // Responsive sizing
				"echarts.init",                       // ECharts initialization
			},
		},
		{
			name:  "Chart with data rendering",
			chart: createMockChartWithData(),
			wantContains: []string{
				"series",                   // Data series
				"xAxis",                    // X-axis configuration
				"\"data\":[{\"value\":10}", // Bar data value format
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderChart(tt.chart)

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain: "+want)
			}

			// Verify fixed width/height are replaced with responsive values
			assert.NotContains(t, output, "width:600px", "Fixed width should be replaced")
			assert.NotContains(t, output, "height:400px", "Fixed height should be replaced")
		})
	}
}

// Helper function to create a mock chart for testing
func createMockChart() *charts.Bar {
	// Create a new bar chart
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Test Chart",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "600px",
			Height: "400px", // These should be replaced with responsive values
			Theme:  "light",
		}),
	)

	// Let go-echarts generate its own ID
	return bar
}

// Helper function to create a chart with data
func createMockChartWithData() *charts.Bar {
	bar := createMockChart()

	// Add X-axis data
	bar.SetXAxis([]string{"A", "B", "C", "D"})

	// Add series data
	bar.AddSeries("Series1", []opts.BarData{
		{Value: 10},
		{Value: 20},
		{Value: 30},
		{Value: 40},
	})

	return bar
}
