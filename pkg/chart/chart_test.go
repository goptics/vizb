package chart

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareTitle(t *testing.T) {
	tests := []struct {
		name       string
		inputName  string
		chartTitle string
		expected   string
	}{
		{
			name:       "With empty name",
			inputName:  "",
			chartTitle: "Test Chart",
			expected:   "Test Chart",
		},
		{
			name:       "With non-empty name",
			inputName:  "BenchName",
			chartTitle: "Test Chart",
			expected:   "BenchName - Test Chart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepareTitle(tt.inputName, tt.chartTitle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "Integer value",
			input:    123.0,
			expected: "123",
		},
		{
			name:     "Decimal value",
			input:    123.45,
			expected: "123.45",
		},
		{
			name:     "Long decimal value",
			input:    123.456789,
			expected: "123.46",
		},
		{
			name:     "Zero",
			input:    0.0,
			expected: "0",
		},
		{
			name:     "Small decimal",
			input:    0.001,
			expected: "0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateFloat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupResultsByName(t *testing.T) {
	tests := []struct {
		name     string
		results  []shared.BenchmarkResult
		expected map[string][]shared.BenchmarkResult
	}{
		{
			name:     "Empty results",
			results:  []shared.BenchmarkResult{},
			expected: map[string][]shared.BenchmarkResult{},
		},
		{
			name: "Single result",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
			},
			expected: map[string][]shared.BenchmarkResult{
				"BenchmarkTest": {
					{
						Name:     "BenchmarkTest",
						Workload: "",
						Subject:  "Subject1",
						Stats: []shared.Stat{
							{Type: "Execution Time", Value: 100.0, Unit: "ns"},
						},
					},
				},
			},
		},
		{
			name: "Multiple results with same name",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
				{
					Name:     "BenchmarkTest",
					Workload: "",
					Subject:  "Subject2",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 200.0, Unit: "ns"},
					},
				},
			},
			expected: map[string][]shared.BenchmarkResult{
				"BenchmarkTest": {
					{
						Name:     "BenchmarkTest",
						Workload: "",
						Subject:  "Subject1",
						Stats: []shared.Stat{
							{Type: "Execution Time", Value: 100.0, Unit: "ns"},
						},
					},
					{
						Name:     "BenchmarkTest",
						Workload: "",
						Subject:  "Subject2",
						Stats: []shared.Stat{
							{Type: "Execution Time", Value: 200.0, Unit: "ns"},
						},
					},
				},
			},
		},
		{
			name: "Multiple results with different names",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest1",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
				{
					Name:     "BenchmarkTest2",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 200.0, Unit: "ns"},
					},
				},
			},
			expected: map[string][]shared.BenchmarkResult{
				"BenchmarkTest1": {
					{
						Name:     "BenchmarkTest1",
						Workload: "",
						Subject:  "Subject1",
						Stats: []shared.Stat{
							{Type: "Execution Time", Value: 100.0, Unit: "ns"},
						},
					},
				},
				"BenchmarkTest2": {
					{
						Name:     "BenchmarkTest2",
						Workload: "",
						Subject:  "Subject1",
						Stats: []shared.Stat{
							{Type: "Execution Time", Value: 200.0, Unit: "ns"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, groupNames := groupResultsByName(tt.results)

			// Check map lengths match
			assert.Equal(t, len(tt.expected), len(result))
			assert.Equal(t, len(tt.expected), len(groupNames))

			// Check each key and its values
			for name, expectedResults := range tt.expected {
				actualResults, ok := result[name]
				assert.True(t, ok, "Expected key %s not found in result", name)
				assert.Equal(t, len(expectedResults), len(actualResults))

				// Compare each result in the slice
				for i, expectedResult := range expectedResults {
					assert.Equal(t, expectedResult.Name, actualResults[i].Name)
					assert.Equal(t, expectedResult.Workload, actualResults[i].Workload)
					assert.Equal(t, expectedResult.Subject, actualResults[i].Subject)
					assert.Equal(t, len(expectedResult.Stats), len(actualResults[i].Stats))

					for j, expectedStat := range expectedResult.Stats {
						assert.Equal(t, expectedStat.Type, actualResults[i].Stats[j].Type)
						assert.Equal(t, expectedStat.Value, actualResults[i].Stats[j].Value)
						assert.Equal(t, expectedStat.Unit, actualResults[i].Stats[j].Unit)
					}
				}
			}
		})
	}
}

func TestCreateChart(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		results   []shared.BenchmarkResult
		statIndex int
		expectNil bool
	}{
		{
			name:      "Empty results",
			title:     "Test Chart",
			results:   []shared.BenchmarkResult{},
			statIndex: 0,
			expectNil: true,
		},
		{
			name:  "Single result",
			title: "Test Chart",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
			},
			statIndex: 0,
			expectNil: false,
		},
		{
			name:  "Multiple results with same workload",
			title: "Test Chart",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "Workload1",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
				{
					Name:     "BenchmarkTest",
					Workload: "Workload1",
					Subject:  "Subject2",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 200.0, Unit: "ns"},
					},
				},
			},
			statIndex: 0,
			expectNil: false,
		},
		{
			name:  "Multiple results with different workloads",
			title: "Test Chart",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "Workload1",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
				{
					Name:     "BenchmarkTest",
					Workload: "Workload2",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 200.0, Unit: "ns"},
					},
				},
			},
			statIndex: 0,
			expectNil: false,
		},
		{
			name:  "Multiple stats",
			title: "Test Chart",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "Workload1",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
						{Type: "Memory Usage", Value: 200.0, Unit: "B"},
						{Type: "Allocations", Value: 5.0, Unit: ""},
					},
				},
			},
			statIndex: 1, // Test with memory usage
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart := createChart(tt.title, tt.results, tt.statIndex)

			if tt.expectNil {
				return
			}

			// Verify chart was created
			assert.NotNil(t, chart)

			// Verify chart title
			assert.Equal(t, tt.title, chart.Title.Title)

			// We can't directly access SeriesList or other internal properties,
			// but we can verify the chart was created with the correct type
			assert.IsType(t, &charts.Bar{}, chart)
		})
	}
}

func TestGenerateHTMLCharts(t *testing.T) {
	tests := []struct {
		name               string
		results            []shared.BenchmarkResult
		expectedChartCount int
		expectedChartNames []string
		expectedStatCounts map[string]int // name -> number of charts per benchmark (should match number of stats types)
	}{
		{
			name:               "Empty results",
			results:            []shared.BenchmarkResult{},
			expectedChartCount: 0,
			expectedChartNames: []string{},
			expectedStatCounts: map[string]int{},
		},
		{
			name: "Single result with one stat",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
			},
			expectedChartCount: 1,
			expectedChartNames: []string{"BenchmarkTest"},
			expectedStatCounts: map[string]int{
				"BenchmarkTest": 1,
			},
		},
		{
			name: "Single result with multiple stats",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
						{Type: "Memory Usage", Value: 200.0, Unit: "B"},
						{Type: "Allocations", Value: 5.0, Unit: ""},
					},
				},
			},
			expectedChartCount: 1,
			expectedChartNames: []string{"BenchmarkTest"},
			expectedStatCounts: map[string]int{
				"BenchmarkTest": 3, // One chart per stat type (time, memory, allocations)
			},
		},
		{
			name: "Multiple results with different names",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest1",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
				{
					Name:     "BenchmarkTest2",
					Workload: "",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 200.0, Unit: "ns"},
						{Type: "Memory Usage", Value: 300.0, Unit: "B"},
					},
				},
			},
			expectedChartCount: 2,
			expectedChartNames: []string{"BenchmarkTest1", "BenchmarkTest2"},
			expectedStatCounts: map[string]int{
				"BenchmarkTest1": 1, // One chart for Execution Time
				"BenchmarkTest2": 2, // One chart each for Execution Time and Memory Usage
			},
		},
		{
			name: "Multiple results with same name",
			results: []shared.BenchmarkResult{
				{
					Name:     "BenchmarkTest",
					Workload: "Workload1",
					Subject:  "Subject1",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.0, Unit: "ns"},
					},
				},
				{
					Name:     "BenchmarkTest",
					Workload: "Workload2",
					Subject:  "Subject2",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 200.0, Unit: "ns"},
						{Type: "Memory Usage", Value: 300.0, Unit: "B"},
					},
				},
			},
			expectedChartCount: 1,
			expectedChartNames: []string{"BenchmarkTest"},
			expectedStatCounts: map[string]int{
				"BenchmarkTest": 1, // Only the first result's stats are used (Execution Time)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			benchCharts := GenerateHTMLCharts(tt.results)

			// Verify number of BenchCharts
			assert.Equal(t, tt.expectedChartCount, len(benchCharts), "Unexpected number of BenchCharts")

			// Verify chart names
			actualNames := make([]string, 0, len(benchCharts))
			for _, bc := range benchCharts {
				actualNames = append(actualNames, bc.Name)
			}

			// Sort slices to ensure consistent comparison
			assert.ElementsMatch(t, tt.expectedChartNames, actualNames, "Chart names don't match")

			// Verify each BenchCharts entry
			for _, bc := range benchCharts {
				expectedStatCount, ok := tt.expectedStatCounts[bc.Name]
				assert.True(t, ok, "Unexpected benchmark name: %s", bc.Name)
				assert.Equal(t, expectedStatCount, len(bc.Charts), "Wrong number of charts for %s", bc.Name)

				// Verify each chart is a Bar chart
				for _, chart := range bc.Charts {
					assert.IsType(t, &charts.Bar{}, chart)
				}
			}
		})
	}
}

// Helper function to create a test file with benchmark events
func createTestFile(t *testing.T, path string, events []string) {
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, event := range events {
		writer.WriteString(event + "\n")
	}

	writer.Flush()
}

func TestIntegrationWithParser(t *testing.T) {
	// Setup test data directory
	tempDir := t.TempDir()

	// Save original flag state to restore after tests
	origTimeUnit := shared.FlagState.TimeUnit
	origMemUnit := shared.FlagState.MemUnit
	origAllocUnit := shared.FlagState.AllocUnit

	// Restore flag state after tests
	defer func() {
		shared.FlagState.TimeUnit = origTimeUnit
		shared.FlagState.MemUnit = origMemUnit
		shared.FlagState.AllocUnit = origAllocUnit
		shared.HasMemStats = false
		shared.CPUCount = 0
	}()

	t.Run("Integration with parser", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "bench.txt")
		createTestFile(t, filePath, []string{
			"BenchmarkGroup/Task/SubjectA 100 123.45 ns/op 64.0 B/op 2 allocs/op",
			"BenchmarkGroup/Task/SubjectB 100 234.56 ns/op 128.0 B/op 4 allocs/op",
		})

		// Set flag state for the test
		shared.FlagState.TimeUnit = "ns"
		shared.FlagState.MemUnit = "B"
		shared.FlagState.AllocUnit = ""
		shared.FlagState.GroupPattern = "subject"

		// Parse the results
		results := parser.ParseBenchmarkResults(filePath)
		assert.Len(t, results, 2)

		// Generate charts
		benchCharts := GenerateHTMLCharts(results)

		// Verify charts were generated
		assert.GreaterOrEqual(t, len(benchCharts), 1, "Should have at least 1 benchmark chart group")

		// Print chart names for debugging
		for _, bc := range benchCharts {
			t.Logf("Found chart group: %s with %d charts", bc.Name, len(bc.Charts))
		}

		// Verify that charts exist for each benchmark output
		chartNames := make(map[string]bool)
		for _, bc := range benchCharts {
			chartNames[bc.Name] = true
		}

		// We're just checking that we have at least one chart group
		// and that charts were generated successfully
		assert.GreaterOrEqual(t, len(chartNames), 1, "Should have at least one chart group")

		// Check that each chart group has the right number of charts based on its stats
		for _, bc := range benchCharts {
			// We should at least have a chart for execution time
			assert.GreaterOrEqual(t, len(bc.Charts), 1,
				"Each chart group should have at least 1 chart")

			// Each chart should be a Bar chart
			for _, chart := range bc.Charts {
				assert.IsType(t, &charts.Bar{}, chart)
			}
		}
	})
}
