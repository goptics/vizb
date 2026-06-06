package golang

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
)

type testBlock struct {
	name           string
	benchContent   []string
	pattern        string
	timeUnit       string
	memUnit        string
	allocUnit      string
	expected       []shared.DataPoint
	expectMemStats bool
	expectCPUCount int
}

func TestParseGoBenchmark(t *testing.T) {
	origTimeUnit := shared.FlagState.TimeUnit
	origMemUnit := shared.FlagState.MemUnit
	origAllocUnit := shared.FlagState.NumberUnit

	defer func() {
		shared.FlagState.TimeUnit = origTimeUnit
		shared.FlagState.MemUnit = origMemUnit
		shared.FlagState.NumberUnit = origAllocUnit
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
			pattern:  "y",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Simple",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
					},
				},
				{
					Name:  "",
					XAxis: "",
					YAxis: "SimpleBench",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 100.45},
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
			pattern:   "y",
			timeUnit:  "ms",
			memUnit:   "KB",
			allocUnit: "K",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "WithMem",
					Stats: []shared.Stat{
						{Type: "Execution Time (ms/op)", Value: 0.00},
						{Type: "Memory Usage (KB/op)", Value: 0.06},
						{Type: "Allocations (K/op)", Value: 0.00},
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
			pattern:   "n/x/y",
			timeUnit:  "ns",
			memUnit:   "b",
			allocUnit: "",
			expected: []shared.DataPoint{
				{
					Name:  "Group",
					XAxis: "Task",
					YAxis: "SubjectA",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
						{Type: "Memory Usage (b/op)", Value: 512.0},
						{Type: "Allocations/op", Value: 2.0},
					},
				},
				{
					Name:  "Group",
					XAxis: "Task",
					YAxis: "SubjectB",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 234.56},
						{Type: "Memory Usage (b/op)", Value: 1024.0},
						{Type: "Allocations/op", Value: 4.0},
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
			pattern:  "n/y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{
					Name:  "Parallel",
					XAxis: "",
					YAxis: "SubjectA",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
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
			pattern:  "y",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Test",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
					},
				},
			},

			expectMemStats: false,
			expectCPUCount: 0,
		},
		{
			name: "Benchmarks with varying iterations",
			benchContent: []string{
				"BenchmarkA 100 100.0 ns/op",
				"BenchmarkB 200 100.0 ns/op",
			},
			timeUnit: "ns",
			pattern:  "y",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "A",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 100.0},
						{Type: "Iterations", Value: 100},
					},
				},
				{
					Name:  "",
					XAxis: "",
					YAxis: "B",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 100.0},
						{Type: "Iterations", Value: 200},
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
			expected:     []shared.DataPoint{},
		},
		{
			name: "Benchmark with B/s (Throughput)",
			benchContent: []string{
				"BenchmarkThroughput 100 123.45 ns/op 512.0 B/s",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Throughput",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
						{Type: "Throughput (B/s)", Value: 512.0},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 0,
		},
		{
			name: "Benchmark with MB/s (Throughput)",
			benchContent: []string{
				"BenchmarkThroughput 100 123.45 ns/op 1024.0 MB/s",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Throughput",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
						{Type: "Throughput (MB/s)", Value: 1024.0},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 0,
		},
		{
			name: "Benchmark with GB/s (Throughput)",
			benchContent: []string{
				"BenchmarkThroughput 100 123.45 ns/op 2.5 GB/s",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Throughput",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
						{Type: "Throughput (GB/s)", Value: 2.5},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 0,
		},
		{
			name: "Benchmark with custom throughput metric (res/s)",
			benchContent: []string{
				"BenchmarkCustom 100 123.45 ns/op 5000.0 res/s",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Custom",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
						{Type: "Throughput (res/s)", Value: 5000.0},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 0,
		},
		{
			name: "Benchmark with custom metric (non-throughput)",
			benchContent: []string{
				"BenchmarkCustom 100 123.45 ns/op 42.5 customUnit",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{
					Name:  "",
					XAxis: "",
					YAxis: "Custom",
					Stats: []shared.Stat{
						{Type: "Execution Time (ns/op)", Value: 123.45},
						{Type: "Metric (customUnit)", Value: 42.5},
					},
				},
			},
			expectMemStats: false,
			expectCPUCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.FlagState.TimeUnit = tt.timeUnit
			shared.FlagState.MemUnit = tt.memUnit
			shared.FlagState.NumberUnit = tt.allocUnit
			shared.FlagState.GroupPattern = tt.pattern
			shared.CPUCount = 0

			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "bench.txt")

			file, err := os.Create(filePath)
			require.NoError(t, err, "Failed to create test file")

			writer := bufio.NewWriter(file)
			for _, event := range tt.benchContent {
				writer.WriteString(event)
				writer.WriteString("\n")
			}

			err = writer.Flush()
			if err != nil {
				log.Fatal(err)
			}

			file.Close()

			results := ParseGoBenchmark(filePath)

			require.Len(t, results, len(tt.expected), "ParseGoBenchmark() returned %d results, expected %d", len(results), len(tt.expected))

			for i, expected := range tt.expected {
				actual := results[i]
				assert.Equal(t, expected.Name, actual.Name, "Result[%d].Name mismatch", i)
				assert.Equal(t, expected.XAxis, actual.XAxis, "Result[%d].XAxis mismatch", i)
				assert.Equal(t, expected.YAxis, actual.YAxis, "Result[%d].YAxis mismatch", i)

				require.Len(t, actual.Stats, len(expected.Stats), "Result[%d] stats count mismatch", i)

				for j, expectedStat := range expected.Stats {
					actualStat := actual.Stats[j]
					assert.Equal(t, expectedStat.Type, actualStat.Type, "Result[%d].Stats[%d].Type mismatch", i, j)
					assert.InDelta(t, expectedStat.Value, actualStat.Value, 0.001, "Result[%d].Stats[%d].Value mismatch", i, j)
				}
			}

			assert.Equal(t, tt.expectCPUCount, shared.CPUCount, "CPUCount mismatch")
		})
	}
}

func TestConvertGoJsonBenchToText(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Valid JSON events", func(t *testing.T) {
		jsonFile := filepath.Join(tempDir, "events.json")
		content := `{"Action":"output","Output":"BenchmarkA 100 100 ns/op\n"}
	{"Action":"output","Output":"BenchmarkB 200 200 ns/op\n"}`
		err := os.WriteFile(jsonFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to write json file")

		txtFile := ConvertGoJsonBenchToText(jsonFile)
		defer os.Remove(txtFile)

		txtContent, err := os.ReadFile(txtFile)
		require.NoError(t, err, "Failed to read result file")

		expected := "BenchmarkA 100 100 ns/op\nBenchmarkB 200 200 ns/op\n"
		assert.Equal(t, expected, string(txtContent))
	})

	t.Run("Mixed actions", func(t *testing.T) {
		jsonFile := filepath.Join(tempDir, "mixed.json")
		content := `{"Action":"run","Test":"BenchmarkA"}
	{"Action":"output","Output":"BenchmarkA 100 100 ns/op\n"}
	{"Action":"pass","Test":"BenchmarkA"}`
		err := os.WriteFile(jsonFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to write json file")

		txtFile := ConvertGoJsonBenchToText(jsonFile)
		defer os.Remove(txtFile)

		txtContent, err := os.ReadFile(txtFile)
		require.NoError(t, err, "Failed to read result file")

		expected := "BenchmarkA 100 100 ns/op\n"
		assert.Equal(t, expected, string(txtContent))
	})
}

func TestShouldIncludeBenchmark(t *testing.T) {
	origFilter := shared.FlagState.FilterRegex
	defer func() {
		shared.FlagState.FilterRegex = origFilter
	}()

	tests := []struct {
		name      string
		filter    string
		benchName string
		expected  bool
	}{
		{
			name:      "Empty filter includes all",
			filter:    "",
			benchName: "AnyBenchmark",
			expected:  true,
		},
		{
			name:      "Exact match",
			filter:    "^TestBenchmark$",
			benchName: "TestBenchmark",
			expected:  true,
		},
		{
			name:      "Exact match fails on partial",
			filter:    "^TestBenchmark$",
			benchName: "TestBenchmarkExtra",
			expected:  false,
		},
		{
			name:      "Partial match with substring",
			filter:    "Success",
			benchName: "BenchmarkValidateSuccess",
			expected:  true,
		},
		{
			name:      "Partial match fails",
			filter:    "Success",
			benchName: "BenchmarkValidateFail",
			expected:  false,
		},
		{
			name:      "Regex with alternation",
			filter:    "Encode|Decode",
			benchName: "BenchmarkEncode",
			expected:  true,
		},
		{
			name:      "Regex with alternation second option",
			filter:    "Encode|Decode",
			benchName: "BenchmarkDecode",
			expected:  true,
		},
		{
			name:      "Regex with alternation no match",
			filter:    "Encode|Decode",
			benchName: "BenchmarkTransform",
			expected:  false,
		},
		{
			name:      "Regex with prefix anchor",
			filter:    "^Parse",
			benchName: "ParseJSON",
			expected:  true,
		},
		{
			name:      "Regex with prefix anchor fails",
			filter:    "^Parse",
			benchName: "JSONParse",
			expected:  false,
		},
		{
			name:      "Regex with suffix anchor",
			filter:    "Fast$",
			benchName: "BenchmarkFast",
			expected:  true,
		},
		{
			name:      "Regex with suffix anchor fails",
			filter:    "Fast$",
			benchName: "FastBenchmark",
			expected:  false,
		},
		{
			name:      "Complex regex pattern",
			filter:    "Benchmark(Parse|Validate).*JSON",
			benchName: "BenchmarkParseComplexJSON",
			expected:  true,
		},
		{
			name:      "Complex regex pattern no match",
			filter:    "Benchmark(Parse|Validate).*JSON",
			benchName: "BenchmarkParseXML",
			expected:  false,
		},
		{
			name:      "Case sensitive match",
			filter:    "test",
			benchName: "Test",
			expected:  false,
		},
		{
			name:      "Case insensitive regex",
			filter:    "(?i)test",
			benchName: "TEST",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.FlagState.FilterRegex = tt.filter

			result := parser.ShouldIncludeBenchmark(tt.benchName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
