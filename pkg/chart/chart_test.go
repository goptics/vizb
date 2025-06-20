package chart

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBenchmarkResults(t *testing.T) {
	// Setup test data directory
	tempDir := t.TempDir()

	t.Run("Valid benchmark results with memory stats", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results
		jsonPath := filepath.Join(tempDir, "bench_with_mem.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
			{Action: "output", Output: "BenchmarkBenchName1/Subject2-8         20000               50 ns/op             150 B/op              3 allocs/op"},
			{Action: "output", Output: "BenchmarkBenchName2/Subject1-8         15000              120 ns/op             180 B/op              4 allocs/op"},
			{Action: "output", Output: "BenchmarkBenchName/Workload/Subject-8   30000               40 ns/op             100 B/op              2 allocs/op"},
		})

		// Set default separator
		shared.FlagState.Separator = "/"

		// Parse the results
		results, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.Len(t, results, 4)

		// Verify the first result - case 2 in switch (two parts)
		assert.Equal(t, "BenchName1", results[0].Name)
		assert.Equal(t, "Subject1", results[0].Subject)
		assert.Equal(t, "", results[0].Workload)
		assert.Equal(t, 100.0, results[0].NsPerOp)
		assert.Equal(t, 200.0, results[0].BytesPerOp)
		assert.Equal(t, uint64(5), results[0].AllocsPerOp)

		// Verify the second result - case 2 in switch (two parts)
		assert.Equal(t, "BenchName1", results[1].Name)
		assert.Equal(t, "Subject2", results[1].Subject)
		assert.Equal(t, "", results[1].Workload)
		assert.Equal(t, 50.0, results[1].NsPerOp)
		assert.Equal(t, 150.0, results[1].BytesPerOp)
		assert.Equal(t, uint64(3), results[1].AllocsPerOp)

		// Verify the third result - case 2 in switch (two parts)
		assert.Equal(t, "BenchName2", results[2].Name)
		assert.Equal(t, "Subject1", results[2].Subject)
		assert.Equal(t, "", results[2].Workload)
		assert.Equal(t, 120.0, results[2].NsPerOp)
		assert.Equal(t, 180.0, results[2].BytesPerOp)
		assert.Equal(t, uint64(4), results[2].AllocsPerOp)

		// Verify the fourth result - default case in switch (three parts)
		// According to the docs, for n > 2:
		// - Name = first part (n-3)
		// - Workload = second-to-last part (n-2)
		// - Subject = last part (n-1)
		assert.Equal(t, "BenchName", results[3].Name)
		assert.Equal(t, "Subject", results[3].Subject)
		assert.Equal(t, "Workload", results[3].Workload)
		assert.Equal(t, 40.0, results[3].NsPerOp)
		assert.Equal(t, 100.0, results[3].BytesPerOp)
		assert.Equal(t, uint64(2), results[3].AllocsPerOp)
	})

	t.Run("Valid benchmark results without memory stats", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results (no memory stats)
		jsonPath := filepath.Join(tempDir, "bench_no_mem.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op"},
			{Action: "output", Output: "BenchmarkBenchName1/Subject2-8         20000               50 ns/op"},
		})

		// Set default separator
		shared.FlagState.Separator = "/"

		// Parse the results
		results, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		// Verify the first result - case 2 in switch (two parts)
		assert.Equal(t, "BenchName1", results[0].Name)
		assert.Equal(t, "Subject1", results[0].Subject)
		assert.Equal(t, "", results[0].Workload)
		assert.Equal(t, 100.0, results[0].NsPerOp)
		assert.Equal(t, 0.0, results[0].BytesPerOp)
		assert.Equal(t, uint64(0), results[0].AllocsPerOp)

		// Verify the second result - case 2 in switch (two parts)
		assert.Equal(t, "BenchName1", results[1].Name)
		assert.Equal(t, "Subject2", results[1].Subject)
		assert.Equal(t, "", results[1].Workload)
		assert.Equal(t, 50.0, results[1].NsPerOp)
		assert.Equal(t, 0.0, results[1].BytesPerOp)
		assert.Equal(t, uint64(0), results[1].AllocsPerOp)
	})

	t.Run("Custom separator", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results using custom separator
		jsonPath := filepath.Join(tempDir, "bench_custom_sep.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1_Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
		})

		// Set custom separator
		shared.FlagState.Separator = "_"

		// Parse the results
		results, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.Len(t, results, 1)

		// With the separator as '_', we expect BenchName1 as Name and Subject1 as Subject
		// The '8' part is stripped as CPU count
		assert.Equal(t, "BenchName1", results[0].Name)
		assert.Equal(t, "Subject1", results[0].Subject)
		assert.Equal(t, "", results[0].Workload)
		assert.Equal(t, 100.0, results[0].NsPerOp)
		assert.Equal(t, 200.0, results[0].BytesPerOp)
		assert.Equal(t, uint64(5), results[0].AllocsPerOp)
	})

	t.Run("Empty file", func(t *testing.T) {
		// Create an empty JSON file
		jsonPath := filepath.Join(tempDir, "empty.json")
		createTestFile(t, jsonPath, []BenchEvent{})

		// Parse the results
		results, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Create a file with invalid JSON
		jsonPath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(jsonPath, []byte("this is not valid JSON"), 0644)
		require.NoError(t, err)

		// Parse the results
		_, err = parseBenchmarkResults(jsonPath)
		assert.Error(t, err)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		// Try to parse a non-existent file
		_, err := parseBenchmarkResults(filepath.Join(tempDir, "non_existent.json"))
		assert.Error(t, err)
	})
}

func TestCreateChart(t *testing.T) {
	t.Run("Chart creation with multiple BenchNames, workloads and subjects", func(t *testing.T) {
		results := []BenchmarkResult{
			{Name: "BenchName1", Subject: "Subject1", Workload: "", NsPerOp: 100, BytesPerOp: 200, AllocsPerOp: 5},
			{Name: "BenchName1", Subject: "Subject2", Workload: "", NsPerOp: 50, BytesPerOp: 150, AllocsPerOp: 3},
			{Name: "BenchName2", Subject: "Subject1", Workload: "", NsPerOp: 120, BytesPerOp: 180, AllocsPerOp: 4},
			{Name: "BenchName3", Subject: "Subject1", Workload: "WorkloadA", NsPerOp: 80, BytesPerOp: 160, AllocsPerOp: 3},
		}

		// Create a chart for NsPerOp
		chart := createChart("Test Chart", results, func(r BenchmarkResult) string {
			return "100"
		})

		// Verify chart properties
		assert.NotNil(t, chart)
		assert.Equal(t, "Test Chart", chart.Title.Title)

		// Verify chart was created successfully
		assert.NotNil(t, chart)

		// We can't directly access SeriesList, but we can verify the chart was created
		assert.NotNil(t, chart)
	})

	t.Run("Empty results", func(t *testing.T) {
		// Create a chart with no results
		chart := createChart("Empty Chart", []BenchmarkResult{}, func(r BenchmarkResult) string {
			return "0"
		})

		// Verify chart properties
		assert.NotNil(t, chart)
		assert.Equal(t, "Empty Chart", chart.Title.Title)

		// Verify chart was created successfully
		assert.NotNil(t, chart)

		// We can't directly access SeriesList, but we can verify the chart was created
		assert.NotNil(t, chart)
	})
}

func TestGenerateChartsFromFile(t *testing.T) {
	// Save the original flag state to restore after the test
	originalFlagState := shared.FlagState
	defer func() {
		shared.FlagState = originalFlagState
	}()

	// Setup test data directory
	tempDir := t.TempDir()

	t.Run("Generate charts with memory stats", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results
		jsonPath := filepath.Join(tempDir, "bench_with_mem.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
			{Action: "output", Output: "BenchmarkBenchName1/Subject2-8         20000               50 ns/op             150 B/op              3 allocs/op"},
		})

		// Set flag state for the test
		shared.FlagState.Separator = "/"
		shared.FlagState.TimeUnit = "ns"
		shared.FlagState.MemUnit = "B"
		shared.FlagState.OutputFile = filepath.Join(tempDir, "output.html")

		// Generate charts
		outputPath, err := GenerateChartsFromFile(jsonPath)
		require.NoError(t, err)

		// Verify output file exists
		assert.Equal(t, shared.FlagState.OutputFile, outputPath)
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)

		// Verify file content (just check it's not empty)
		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.NotEmpty(t, content)
	})

	t.Run("Generate charts without memory stats", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results (no memory stats)
		jsonPath := filepath.Join(tempDir, "bench_no_mem.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op"},
		})

		// Set flag state for the test
		shared.FlagState.Separator = "/"
		shared.FlagState.TimeUnit = "ms"
		shared.FlagState.OutputFile = filepath.Join(tempDir, "output_no_mem.html")

		// Generate charts
		outputPath, err := GenerateChartsFromFile(jsonPath)
		require.NoError(t, err)

		// Verify output file exists
		assert.Equal(t, shared.FlagState.OutputFile, outputPath)
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)

		// Verify file content (just check it's not empty)
		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.NotEmpty(t, content)
	})

	t.Run("Different time units", func(t *testing.T) {
		// Test with different time units
		jsonPath := filepath.Join(tempDir, "bench_time_units.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
		})

		// Test different time units
		timeUnits := []string{"ns", "us", "ms", "s"}
		for _, unit := range timeUnits {
			t.Run(fmt.Sprintf("TimeUnit_%s", unit), func(t *testing.T) {
				shared.FlagState.TimeUnit = unit
				shared.FlagState.OutputFile = filepath.Join(tempDir, fmt.Sprintf("output_%s.html", unit))
				
				outputPath, err := GenerateChartsFromFile(jsonPath)
				require.NoError(t, err)
				
				// Verify output file exists
				_, err = os.Stat(outputPath)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("Different memory units", func(t *testing.T) {
		// Test with different memory units
		jsonPath := filepath.Join(tempDir, "bench_mem_units.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
		})

		// Test different memory units
		memUnits := []string{"b", "B", "kb", "mb", "gb"}
		for _, unit := range memUnits {
			t.Run(fmt.Sprintf("MemUnit_%s", unit), func(t *testing.T) {
				shared.FlagState.MemUnit = unit
				shared.FlagState.OutputFile = filepath.Join(tempDir, fmt.Sprintf("output_mem_%s.html", unit))
				
				outputPath, err := GenerateChartsFromFile(jsonPath)
				require.NoError(t, err)
				
				// Verify output file exists
				_, err = os.Stat(outputPath)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("Different allocation units", func(t *testing.T) {
		// Test with different allocation units
		jsonPath := filepath.Join(tempDir, "bench_alloc_units.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
		})

		// Test different allocation units
		allocUnits := []string{"", "K", "M", "B", "T"}
		for _, unit := range allocUnits {
			t.Run(fmt.Sprintf("AllocUnit_%s", unit), func(t *testing.T) {
				shared.FlagState.AllocUnit = unit
				shared.FlagState.OutputFile = filepath.Join(tempDir, fmt.Sprintf("output_alloc_%s.html", unit))
				
				outputPath, err := GenerateChartsFromFile(jsonPath)
				require.NoError(t, err)
				
				// Verify output file exists
				_, err = os.Stat(outputPath)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("No benchmark results", func(t *testing.T) {
		// Create a temporary JSON file with no benchmark results
		jsonPath := filepath.Join(tempDir, "no_results.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "Some other output"},
		})

		// Generate charts
		_, err := GenerateChartsFromFile(jsonPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no benchmark results found")
	})

	t.Run("Non-existent file", func(t *testing.T) {
		// Try to generate charts from a non-existent file
		_, err := GenerateChartsFromFile(filepath.Join(tempDir, "non_existent.json"))
		assert.Error(t, err)
	})

	t.Run("Error creating output file", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results
		jsonPath := filepath.Join(tempDir, "bench_valid.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkBenchName1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
		})

		// Set output file to a non-writable location
		shared.FlagState.OutputFile = "/non/existent/directory/output.html"

		// Generate charts - should fail due to invalid output path
		_, err := GenerateChartsFromFile(jsonPath)
		assert.Error(t, err)
	})
}

func TestPrepareChartTitle(t *testing.T) {
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
			result := prepareChartTitle(tt.inputName, tt.chartTitle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a test file with benchmark events
func createTestFile(t *testing.T, path string, events []BenchEvent) {
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, event := range events {
		err := encoder.Encode(event)
		require.NoError(t, err)
	}
}
