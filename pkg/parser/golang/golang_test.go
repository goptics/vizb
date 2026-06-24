package golang

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

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

// GoBenchmarkSuite exercises ParseGoBenchmark with a per-case parser.Config,
// resetting the shared.CPUCount global between cases.
type GoBenchmarkSuite struct {
	suite.Suite
}

func (s *GoBenchmarkSuite) SetupTest() {
	shared.CPUCount = 0
}

func (s *GoBenchmarkSuite) TearDownTest() {
	shared.CPUCount = 0
}

func (s *GoBenchmarkSuite) writeFile(lines []string) string {
	filePath := filepath.Join(s.T().TempDir(), "bench.txt")
	file, err := os.Create(filePath)
	s.Require().NoError(err, "Failed to create test file")

	writer := bufio.NewWriter(file)
	for _, event := range lines {
		_, err := writer.WriteString(event)
		s.Require().NoError(err)
		_, err = writer.WriteString("\n")
		s.Require().NoError(err)
	}
	s.Require().NoError(writer.Flush())
	file.Close()
	return filePath
}

func (s *GoBenchmarkSuite) TestParseGoBenchmark() {
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
				{YAxis: "Simple", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)}}},
				{YAxis: "SimpleBench", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: shared.F64(100.45)}}},
			},
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
				{YAxis: "WithMem", Stats: []shared.Stat{
					{Type: "Execution Time (ms/op)", Value: shared.F64(0.00)},
					{Type: "Memory Usage (KB/op)", Value: shared.F64(0.06)},
					{Type: "Allocations (K/op)", Value: shared.F64(0.00)},
				}},
			},
			expectMemStats: true,
		},
		{
			name: "Multiple benchmarks with workloads",
			benchContent: []string{
				"BenchmarkGroup/Task/SubjectA 100 123.45 ns/op 64.0 B/op 2 allocs/op",
				"BenchmarkGroup/Task/SubjectB 100 234.56 ns/op 128.0 B/op 4 allocs/op",
			},
			pattern:  "n/x/y",
			timeUnit: "ns",
			memUnit:  "b",
			expected: []shared.DataPoint{
				{Name: "Group", XAxis: "Task", YAxis: "SubjectA", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)},
					{Type: "Memory Usage (b/op)", Value: shared.F64(512.0)},
					{Type: "Allocations/op", Value: shared.F64(2.0)},
				}},
				{Name: "Group", XAxis: "Task", YAxis: "SubjectB", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(234.56)},
					{Type: "Memory Usage (b/op)", Value: shared.F64(1024.0)},
					{Type: "Allocations/op", Value: shared.F64(4.0)},
				}},
			},
			expectMemStats: true,
		},
		{
			name: "Benchmark with CPU count in subject",
			benchContent: []string{
				"BenchmarkParallel/SubjectA-8 100 123.45 ns/op",
			},
			pattern:  "n/y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{Name: "Parallel", YAxis: "SubjectA", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)},
				}},
			},
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
				{YAxis: "Test", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)}}},
			},
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
				{YAxis: "A", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(100.0)},
					{Type: "Iterations", Value: shared.F64(100)},
				}},
				{YAxis: "B", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(100.0)},
					{Type: "Iterations", Value: shared.F64(200)},
				}},
			},
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
				{YAxis: "Throughput", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)},
					{Type: "Throughput (B/s)", Value: shared.F64(512.0)},
				}},
			},
		},
		{
			name: "Benchmark with MB/s (Throughput)",
			benchContent: []string{
				"BenchmarkThroughput 100 123.45 ns/op 1024.0 MB/s",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{YAxis: "Throughput", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)},
					{Type: "Throughput (MB/s)", Value: shared.F64(1024.0)},
				}},
			},
		},
		{
			name: "Benchmark with GB/s (Throughput)",
			benchContent: []string{
				"BenchmarkThroughput 100 123.45 ns/op 2.5 GB/s",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{YAxis: "Throughput", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)},
					{Type: "Throughput (GB/s)", Value: shared.F64(2.5)},
				}},
			},
		},
		{
			name: "Benchmark with custom throughput metric (res/s)",
			benchContent: []string{
				"BenchmarkCustom 100 123.45 ns/op 5000.0 res/s",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{YAxis: "Custom", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)},
					{Type: "Throughput (res/s)", Value: shared.F64(5000.0)},
				}},
			},
		},
		{
			name: "Benchmark with custom metric (non-throughput)",
			benchContent: []string{
				"BenchmarkCustom 100 123.45 ns/op 42.5 customUnit",
			},
			pattern:  "y",
			timeUnit: "ns",
			expected: []shared.DataPoint{
				{YAxis: "Custom", Stats: []shared.Stat{
					{Type: "Execution Time (ns/op)", Value: shared.F64(123.45)},
					{Type: "Metric (customUnit)", Value: shared.F64(42.5)},
				}},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			shared.CPUCount = 0
			cfg := parser.Config{
				GroupPattern: tt.pattern,
				TimeUnit:     tt.timeUnit,
				MemUnit:      tt.memUnit,
				NumberUnit:   tt.allocUnit,
			}

			results := ParseGoBenchmark(s.writeFile(tt.benchContent), cfg)

			s.Require().Len(results, len(tt.expected))

			for i, expected := range tt.expected {
				actual := results[i]
				s.Equal(expected.Name, actual.Name, "Result[%d].Name", i)
				s.Equal(expected.XAxis, actual.XAxis, "Result[%d].XAxis", i)
				s.Equal(expected.YAxis, actual.YAxis, "Result[%d].YAxis", i)

				s.Require().Len(actual.Stats, len(expected.Stats), "Result[%d] stats count", i)

				for j, expectedStat := range expected.Stats {
					actualStat := actual.Stats[j]
					s.Equal(expectedStat.Type, actualStat.Type, "Result[%d].Stats[%d].Type", i, j)
					s.InDelta(*expectedStat.Value, *actualStat.Value, 0.001, "Result[%d].Stats[%d].Value", i, j)
				}
			}

			s.Equal(tt.expectCPUCount, shared.CPUCount, "CPUCount")
		})
	}
}

func (s *GoBenchmarkSuite) TestConvertGoJsonBenchToText() {
	tempDir := s.T().TempDir()

	s.Run("Valid JSON events", func() {
		jsonFile := filepath.Join(tempDir, "events.json")
		content := `{"Action":"output","Output":"BenchmarkA 100 100 ns/op\n"}
	{"Action":"output","Output":"BenchmarkB 200 200 ns/op\n"}`
		s.Require().NoError(os.WriteFile(jsonFile, []byte(content), 0644))

		txtFile := ConvertGoJsonBenchToText(jsonFile)
		defer os.Remove(txtFile)

		txtContent, err := os.ReadFile(txtFile)
		s.Require().NoError(err)

		s.Equal("BenchmarkA 100 100 ns/op\nBenchmarkB 200 200 ns/op\n", string(txtContent))
	})

	s.Run("Mixed actions", func() {
		jsonFile := filepath.Join(tempDir, "mixed.json")
		content := `{"Action":"run","Test":"BenchmarkA"}
	{"Action":"output","Output":"BenchmarkA 100 100 ns/op\n"}
	{"Action":"pass","Test":"BenchmarkA"}`
		s.Require().NoError(os.WriteFile(jsonFile, []byte(content), 0644))

		txtFile := ConvertGoJsonBenchToText(jsonFile)
		defer os.Remove(txtFile)

		txtContent, err := os.ReadFile(txtFile)
		s.Require().NoError(err)

		s.Equal("BenchmarkA 100 100 ns/op\n", string(txtContent))
	})
}

func TestGoBenchmarkSuite(t *testing.T) {
	suite.Run(t, new(GoBenchmarkSuite))
}

// ShouldIncludeBenchmarkSuite exercises parser.ShouldIncludeBenchmark via a
// per-case parser.Config filter.
type ShouldIncludeBenchmarkSuite struct {
	suite.Suite
}

func (s *ShouldIncludeBenchmarkSuite) TestShouldIncludeBenchmark() {
	tests := []struct {
		name      string
		filter    string
		benchName string
		expected  bool
	}{
		{"Empty filter includes all", "", "AnyBenchmark", true},
		{"Exact match", "^TestBenchmark$", "TestBenchmark", true},
		{"Exact match fails on partial", "^TestBenchmark$", "TestBenchmarkExtra", false},
		{"Partial match with substring", "Success", "BenchmarkValidateSuccess", true},
		{"Partial match fails", "Success", "BenchmarkValidateFail", false},
		{"Regex with alternation", "Encode|Decode", "BenchmarkEncode", true},
		{"Regex with alternation second option", "Encode|Decode", "BenchmarkDecode", true},
		{"Regex with alternation no match", "Encode|Decode", "BenchmarkTransform", false},
		{"Regex with prefix anchor", "^Parse", "ParseJSON", true},
		{"Regex with prefix anchor fails", "^Parse", "JSONParse", false},
		{"Regex with suffix anchor", "Fast$", "BenchmarkFast", true},
		{"Regex with suffix anchor fails", "Fast$", "FastBenchmark", false},
		{"Complex regex pattern", "Benchmark(Parse|Validate).*JSON", "BenchmarkParseComplexJSON", true},
		{"Complex regex pattern no match", "Benchmark(Parse|Validate).*JSON", "BenchmarkParseXML", false},
		{"Case sensitive match", "test", "Test", false},
		{"Case insensitive regex", "(?i)test", "TEST", true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := parser.ShouldIncludeBenchmark(tt.benchName, parser.Config{Filter: tt.filter})
			s.Equal(tt.expected, result)
		})
	}
}

func TestShouldIncludeBenchmarkSuite(t *testing.T) {
	suite.Run(t, new(ShouldIncludeBenchmarkSuite))
}
