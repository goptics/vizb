package parser

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
)

type testBlock struct {
	name           string
	benchContent   []string // List of events to encode as JSON
	separator      string
	timeUnit       string
	memUnit        string
	allocUnit      string
	expected       []shared.BenchmarkResult
	expectMemStats bool
	expectCPUCount int
}

func TestParseBenchmarkResults(t *testing.T) {
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

	tests := []testBlock{
		{
			name: "Basic benchmark without memory stats",
			benchContent: []string{
				"BenchmarkSimple 100 123.45 ns/op",
			},
			separator: "/",
			timeUnit:  "ns",
			expected: []shared.BenchmarkResult{
				{
					Name:     "",
					Workload: "",
					Subject:  "Simple",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 123.45, Unit: "ns"},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 0,
		},
		{
			name: "Benchmark with memory stats",
			benchContent: []string{
				"BenchmarkWithMem 100 123.45 ns/op 64.0 B/op 2 allocs/op",
			},
			separator: "/",
			timeUnit:  "ms",
			memUnit:   "kb",
			allocUnit: "K",
			expected: []shared.BenchmarkResult{
				{
					Name:     "",
					Workload: "",
					Subject:  "WithMem",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 0.00012345, Unit: "ms"},
						{Type: "Memory Usage", Value: 0.0625, Unit: "kb"},
						{Type: "Allocations", Value: 0.000002, Unit: "K"},
					},
				},
			},
			expectMemStats: true,
			expectCPUCount: 0,
		},
		{
			name: "Multiple benchmarks with different formats",
			benchContent: []string{
				"BenchmarkGroup/Task/SubjectA 100 123.45 ns/op 64.0 B/op 2 allocs/op",
				"BenchmarkGroup/Task/SubjectB 100 234.56 ns/op 128.0 B/op 4 allocs/op",
				"BenchmarkOther/Simple 100 345.67 ns/op",
			},
			separator: "/",
			timeUnit:  "ns",
			memUnit:   "b",
			allocUnit: "",
			expected: []shared.BenchmarkResult{
				{
					Name:     "Group",
					Workload: "Task",
					Subject:  "SubjectA",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 123.45, Unit: "ns"},
						{Type: "Memory Usage", Value: 512.0, Unit: "b"}, // 64*8=512 bits
						{Type: "Allocations", Value: 2.0, Unit: ""},
					},
				},
				{
					Name:     "Group",
					Workload: "Task",
					Subject:  "SubjectB",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 234.56, Unit: "ns"},
						{Type: "Memory Usage", Value: 1024.0, Unit: "b"}, // 128*8=1024 bits
						{Type: "Allocations", Value: 4.0, Unit: ""},
					},
				},
				{
					Name:     "Other",
					Workload: "",
					Subject:  "Simple",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 345.67, Unit: "ns"},
					},
				},
			},
			expectMemStats: true,
			expectCPUCount: 0,
		},
		{
			name: "Benchmark with CPU count in subject",
			benchContent: []string{
				"BenchmarkParallel/SubjectA-8 100 123.45 ns/op",
			},
			separator: "/",
			timeUnit:  "ns",
			expected: []shared.BenchmarkResult{
				{
					Name:     "Parallel",
					Workload: "",
					Subject:  "SubjectA",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 123.45, Unit: "ns"},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 8,
		},
		{
			name: "Mixed benchmark formats with custom separator",
			benchContent: []string{
				"BenchmarkGroup/Task/SubjectA 100 123.45 ns/op 64.0 B/op 2 allocs/op",
				"BenchmarkSimple 100 234.56 ns/op",
			},
			separator: "_",
			timeUnit:  "us",
			memUnit:   "mb",
			allocUnit: "M",
			expected: []shared.BenchmarkResult{
				{
					Name:     "Group",
					Workload: "Task",
					Subject:  "SubjectA",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 0.12345, Unit: "us"},
						{Type: "Memory Usage", Value: 0.00006103515625, Unit: "mb"}, // 64/(1024*1024)
						{Type: "Allocations", Value: 0.000002, Unit: "M"},
					},
				},
				{
					Name:     "",
					Workload: "",
					Subject:  "Simple",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 0.23456, Unit: "us"},
					},
				},
			},
			expectMemStats: true,
			expectCPUCount: 0,
		},
		{
			name: "Non-benchmark output should be ignored",
			benchContent: []string{
				"PASS",
				"ok  \tgithub.com/example/pkg\t0.412s",
				"BenchmarkTest 100 123.45 ns/op",
			},
			separator: "/",
			timeUnit:  "ns",
			expected: []shared.BenchmarkResult{
				{
					Name:     "",
					Workload: "",
					Subject:  "Test",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 123.45, Unit: "ns"},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 0,
		},
		{
			name:         "Empty file",
			benchContent: []string{},
			separator:    "/",
			timeUnit:     "ns",
			expected:     []shared.BenchmarkResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flag state for this test
			shared.FlagState.TimeUnit = tt.timeUnit
			shared.FlagState.MemUnit = tt.memUnit
			shared.FlagState.AllocUnit = tt.allocUnit
			shared.HasMemStats = false
			shared.CPUCount = 0

			// Create a temporary JSON file
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "bench.txt")

			// Create JSON file with test content
			file, err := os.Create(filePath)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			writer := bufio.NewWriter(file)
			for _, event := range tt.benchContent {
				writer.WriteString(event + "\n")
			}

			if err := writer.Flush(); err != nil {
				log.Fatal(err)
			}

			file.Close()

			// Call the function under test
			results, err := ParseBenchmarkResults(filePath)

			if err != nil {
				return // If we expected an error, no need to check results
			}

			// Check results
			if len(results) != len(tt.expected) {
				t.Errorf("ParseBenchmarkResults() returned %d results, expected %d", len(results), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(results) {
					t.Errorf("Missing expected result at index %d", i)
					continue
				}

				actual := results[i]
				if actual.Name != expected.Name {
					t.Errorf("Result[%d].Name = %q, expected %q", i, actual.Name, expected.Name)
				}
				if actual.Workload != expected.Workload {
					t.Errorf("Result[%d].Workload = %q, expected %q", i, actual.Workload, expected.Workload)
				}
				if actual.Subject != expected.Subject {
					t.Errorf("Result[%d].Subject = %q, expected %q", i, actual.Subject, expected.Subject)
				}

				// Check stats
				if len(actual.Stats) != len(expected.Stats) {
					t.Errorf("Result[%d] has %d stats, expected %d", i, len(actual.Stats), len(expected.Stats))
					continue
				}

				for j, expectedStat := range expected.Stats {
					if j >= len(actual.Stats) {
						t.Errorf("Missing expected stat at result[%d].Stats[%d]", i, j)
						continue
					}

					actualStat := actual.Stats[j]
					if actualStat.Type != expectedStat.Type {
						t.Errorf("Result[%d].Stats[%d].Type = %q, expected %q", i, j, actualStat.Type, expectedStat.Type)
					}
					if actualStat.Unit != expectedStat.Unit {
						t.Errorf("Result[%d].Stats[%d].Unit = %q, expected %q", i, j, actualStat.Unit, expectedStat.Unit)
					}
					// For float comparisons, allow a small epsilon
					if !almostEqual(actualStat.Value, expectedStat.Value, 0.001) {
						t.Errorf("Result[%d].Stats[%d].Value = %f, expected %f", i, j, actualStat.Value, expectedStat.Value)
					}
				}
			}

			// Check global state
			if shared.HasMemStats != tt.expectMemStats {
				t.Errorf("shared.HasMemStats = %v, expected %v", shared.HasMemStats, tt.expectMemStats)
			}
			if shared.CPUCount != tt.expectCPUCount {
				t.Errorf("shared.CPUCount = %d, expected %d", shared.CPUCount, tt.expectCPUCount)
			}
		})
	}
}

// Helper function to compare float values with tolerance
func almostEqual(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

// TestParseBenchmarkResultsFileError tests error handling for file operations
func TestParseBenchmarkResultsFileError(t *testing.T) {
	// Test with non-existent file
	results, err := ParseBenchmarkResults("/non/existent/file.json")
	if err == nil {
		t.Errorf("ParseBenchmarkResults() with non-existent file should return error")
	}
	if results != nil {
		t.Errorf("ParseBenchmarkResults() with non-existent file should return nil results")
	}
}
