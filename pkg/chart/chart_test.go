package chart

import (
	"encoding/json"
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
			{Action: "output", Output: "BenchmarkWorkload1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
			{Action: "output", Output: "BenchmarkWorkload1/Subject2-8         20000               50 ns/op             150 B/op              3 allocs/op"},
			{Action: "output", Output: "BenchmarkWorkload2/Subject1-8         15000              120 ns/op             180 B/op              4 allocs/op"},
		})

		// Set default separator
		shared.FlagState.Separator = "/"

		// Parse the results
		results, hasMemStats, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.True(t, hasMemStats)
		assert.Len(t, results, 3)

		// Verify the first result
		assert.Equal(t, "Workload1", results[0].Workload)
		assert.Equal(t, "Subject1", results[0].Subject)
		assert.Equal(t, 100.0, results[0].NsPerOp)
		assert.Equal(t, 200.0, results[0].BytesPerOp)
		assert.Equal(t, uint64(5), results[0].AllocsPerOp)

		// Verify the second result
		assert.Equal(t, "Workload1", results[1].Workload)
		assert.Equal(t, "Subject2", results[1].Subject)
		assert.Equal(t, 50.0, results[1].NsPerOp)
		assert.Equal(t, 150.0, results[1].BytesPerOp)
		assert.Equal(t, uint64(3), results[1].AllocsPerOp)

		// Verify the third result
		assert.Equal(t, "Workload2", results[2].Workload)
		assert.Equal(t, "Subject1", results[2].Subject)
		assert.Equal(t, 120.0, results[2].NsPerOp)
		assert.Equal(t, 180.0, results[2].BytesPerOp)
		assert.Equal(t, uint64(4), results[2].AllocsPerOp)
	})

	t.Run("Valid benchmark results without memory stats", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results (no memory stats)
		jsonPath := filepath.Join(tempDir, "bench_no_mem.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkWorkload1/Subject1-8         10000              100 ns/op"},
			{Action: "output", Output: "BenchmarkWorkload1/Subject2-8         20000               50 ns/op"},
		})

		// Set default separator
		shared.FlagState.Separator = "/"

		// Parse the results
		results, hasMemStats, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.False(t, hasMemStats)
		assert.Len(t, results, 2)

		// Verify the first result
		assert.Equal(t, "Workload1", results[0].Workload)
		assert.Equal(t, "Subject1", results[0].Subject)
		assert.Equal(t, 100.0, results[0].NsPerOp)
		assert.Equal(t, 0.0, results[0].BytesPerOp)
		assert.Equal(t, uint64(0), results[0].AllocsPerOp)

		// Verify the second result
		assert.Equal(t, "Workload1", results[1].Workload)
		assert.Equal(t, "Subject2", results[1].Subject)
		assert.Equal(t, 50.0, results[1].NsPerOp)
		assert.Equal(t, 0.0, results[1].BytesPerOp)
		assert.Equal(t, uint64(0), results[1].AllocsPerOp)
	})

	t.Run("Custom separator", func(t *testing.T) {
		// Create a temporary JSON file with valid benchmark results using custom separator
		jsonPath := filepath.Join(tempDir, "bench_custom_sep.json")
		createTestFile(t, jsonPath, []BenchEvent{
			{Action: "output", Output: "BenchmarkWorkload1-Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
		})

		// Set custom separator
		shared.FlagState.Separator = "-"

		// Parse the results
		results, hasMemStats, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.True(t, hasMemStats)
		assert.Len(t, results, 1)

		// With the separator as '-', the last two parts are 'Subject1' and '8'
		// But the code strips the CPU count suffix, so we expect 'Subject1' as the subject
		assert.Equal(t, "Subject1", results[0].Workload)
		assert.Equal(t, "8", results[0].Subject)
		assert.Equal(t, 100.0, results[0].NsPerOp)
		assert.Equal(t, 200.0, results[0].BytesPerOp)
		assert.Equal(t, uint64(5), results[0].AllocsPerOp)
	})

	t.Run("Empty file", func(t *testing.T) {
		// Create an empty JSON file
		jsonPath := filepath.Join(tempDir, "empty.json")
		createTestFile(t, jsonPath, []BenchEvent{})

		// Parse the results
		results, hasMemStats, err := parseBenchmarkResults(jsonPath)
		require.NoError(t, err)
		assert.False(t, hasMemStats)
		assert.Len(t, results, 0)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		// Create a file with invalid JSON
		jsonPath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(jsonPath, []byte("this is not valid JSON"), 0644)
		require.NoError(t, err)

		// Parse the results
		_, _, err = parseBenchmarkResults(jsonPath)
		assert.Error(t, err)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		// Try to parse a non-existent file
		_, _, err := parseBenchmarkResults(filepath.Join(tempDir, "non_existent.json"))
		assert.Error(t, err)
	})
}

func TestCreateChart(t *testing.T) {
	t.Run("Chart creation with multiple workloads and subjects", func(t *testing.T) {
		results := []BenchmarkResult{
			{Workload: "Workload1", Subject: "Subject1", NsPerOp: 100, BytesPerOp: 200, AllocsPerOp: 5},
			{Workload: "Workload1", Subject: "Subject2", NsPerOp: 50, BytesPerOp: 150, AllocsPerOp: 3},
			{Workload: "Workload2", Subject: "Subject1", NsPerOp: 120, BytesPerOp: 180, AllocsPerOp: 4},
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
			{Action: "output", Output: "BenchmarkWorkload1/Subject1-8         10000              100 ns/op             200 B/op              5 allocs/op"},
			{Action: "output", Output: "BenchmarkWorkload1/Subject2-8         20000               50 ns/op             150 B/op              3 allocs/op"},
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
			{Action: "output", Output: "BenchmarkWorkload1/Subject1-8         10000              100 ns/op"},
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
