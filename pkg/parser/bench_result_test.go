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
	pattern        string
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
	origAllocUnit := shared.FlagState.NumberUnit

	// Restore flag state after tests
	defer func() {
		shared.FlagState.TimeUnit = origTimeUnit
		shared.FlagState.MemUnit = origMemUnit
		shared.FlagState.NumberUnit = origAllocUnit
		shared.HasMemStats = false
		shared.CPUCount = 0
	}()

	tests := []testBlock{
		{
			name: "Basic benchmark without memory stats",
			benchContent: []string{
				"BenchmarkSimple 100 123.45 ns/op",
				"BenchmarkSimpleBench 100 100.45 ns/op",
			},
			timeUnit: "ns",
			pattern:  "s",
			expected: []shared.BenchmarkResult{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Simple",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 123.45, Unit: "ns"},
					},
				},
				{
					Name:  "",
					XAxis: "",
					YAxis: "SimpleBench",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 100.45, Unit: "ns"},
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
			pattern:   "s",
			timeUnit:  "ms",
			memUnit:   "kb",
			allocUnit: "K",
			expected: []shared.BenchmarkResult{
				{
					Name:  "",
					XAxis: "",
					YAxis: "WithMem",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 0.00012345, Unit: "ms"},
						{Type: "Memory Usage", Value: 0.0625, Unit: "kb"},
						{Type: "Allocations", Value: 0.002000, Unit: "K"},
					},
				},
			},
			expectMemStats: true,
			expectCPUCount: 0,
		},
		{
			name: "Multiple benchmarks with workloads",
			benchContent: []string{
				"BenchmarkGroup/Task/SubjectA 100 123.45 ns/op 64.0 B/op 2 allocs/op",
				"BenchmarkGroup/Task/SubjectB 100 234.56 ns/op 128.0 B/op 4 allocs/op",
			},
			pattern:   "n/w/s",
			timeUnit:  "ns",
			memUnit:   "b",
			allocUnit: "",
			expected: []shared.BenchmarkResult{
				{
					Name:  "Group",
					XAxis: "Task",
					YAxis: "SubjectA",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 123.45, Unit: "ns"},
						{Type: "Memory Usage", Value: 512.0, Unit: "b"}, // 64*8=512 bits
						{Type: "Allocations", Value: 2.0, Unit: ""},
					},
				},
				{
					Name:  "Group",
					XAxis: "Task",
					YAxis: "SubjectB",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 234.56, Unit: "ns"},
						{Type: "Memory Usage", Value: 1024.0, Unit: "b"}, // 128*8=1024 bits
						{Type: "Allocations", Value: 4.0, Unit: ""},
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
			pattern:  "n/s",
			timeUnit: "ns",
			expected: []shared.BenchmarkResult{
				{
					Name:  "Parallel",
					XAxis: "",
					YAxis: "SubjectA",
					Stats: []shared.Stat{
						{Type: "Execution Time", Value: 123.45, Unit: "ns"},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 8,
		},
		{
			name: "Non-benchmark output should be ignored",
			benchContent: []string{
				"PASS",
				"ok  \tgithub.com/example/pkg\t0.412s",
				"BenchmarkTest 100 123.45 ns/op",
			},
			timeUnit: "ns",
			pattern:  "s",
			expected: []shared.BenchmarkResult{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Test",
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
			timeUnit:     "ns",
			expected:     []shared.BenchmarkResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flag state for this test
			shared.FlagState.TimeUnit = tt.timeUnit
			shared.FlagState.MemUnit = tt.memUnit
			shared.FlagState.NumberUnit = tt.allocUnit
			shared.FlagState.GroupPattern = tt.pattern
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
			results := ParseBenchmarkResults(filePath)

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
				if actual.XAxis != expected.XAxis {
					t.Errorf("Result[%d].XAxis = %q, expected %q", i, actual.XAxis, expected.XAxis)
				}
				if actual.YAxis != expected.YAxis {
					t.Errorf("Result[%d].YAxis = %q, expected %q", i, actual.YAxis, expected.YAxis)
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
