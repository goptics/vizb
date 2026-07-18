package golang

import (
	"errors"
	"io"
	"strings"
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

type goErrorReader struct{}

func (goErrorReader) Read([]byte) (int, error) {
	return 0, errors.New("injected read failure")
}

type dataErrorReader struct {
	data []byte
	err  error
	done bool
}

func (r *dataErrorReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, r.err
	}
	r.done = true
	return copy(p, r.data), r.err
}

func (s *GoBenchmarkSuite) SetupTest() {
	shared.CPUCount = 0
}

func (s *GoBenchmarkSuite) TearDownTest() {
	shared.CPUCount = 0
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

			results, _, err := ParseGoBenchmark(strings.NewReader(strings.Join(tt.benchContent, "\n")), cfg)
			s.Require().NoError(err)

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

func (s *GoBenchmarkSuite) TestParseGoBenchmarkReturnsErrors() {
	benchmark := "BenchmarkExample 100 123 ns/op"

	s.Run("invalid filter", func() {
		_, _, err := ParseGoBenchmark(strings.NewReader(benchmark), parser.Config{
			GroupPattern: "y",
			Filter:       "[",
		})
		s.ErrorContains(err, "invalid filter regex")
	})

	s.Run("invalid benchmark group pattern", func() {
		_, _, err := ParseGoBenchmark(strings.NewReader(benchmark), parser.Config{
			GroupPattern: "[n/y]",
		})
		s.ErrorContains(err, "bracket slots")
	})

	s.Run("filter excludes benchmark", func() {
		results, _, err := ParseGoBenchmark(strings.NewReader(benchmark), parser.Config{
			GroupPattern: "y",
			Filter:       "^Other$",
		})
		s.Require().NoError(err)
		s.Empty(results)
	})

	s.Run("non-result record is ignored", func() {
		results, _, err := ParseGoBenchmark(strings.NewReader("BenchmarkBroken not-a-result\n"), parser.Config{
			GroupPattern: "y",
		})
		s.Require().NoError(err)
		s.Empty(results)
	})

	s.Run("reader failure", func() {
		_, _, err := ParseGoBenchmark(goErrorReader{}, parser.Config{GroupPattern: "y"})
		s.ErrorContains(err, "read Go benchmark")
	})

	s.Run("benchmark reader fails after text", func() {
		input := &dataErrorReader{
			data: []byte("BenchmarkExample 100 123 ns/op\n"),
			err:  errors.New("injected trailing failure"),
		}
		_, _, err := ParseGoBenchmark(input, parser.Config{GroupPattern: "y"})
		s.ErrorContains(err, "read Go benchmark")
	})
}

func (s *GoBenchmarkSuite) TestParseGoBenchmarkJSONEvents() {
	input := strings.Join([]string{
		`{"Action":"run","Test":"BenchmarkExample"}`,
		`{"Action":"output","Output":"BenchmarkExample-8 100 123 ns/op\n"}`,
		`{"Action":"pass","Test":"BenchmarkExample"}`,
	}, "\n")

	results, _, err := ParseGoBenchmark(strings.NewReader(input), parser.Config{
		GroupPattern: "y",
		TimeUnit:     "ns",
	})

	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Equal("Example", results[0].YAxis)
	s.Equal(8, shared.CPUCount)
	s.Require().Len(results[0].Stats, 1)
	s.Equal("Execution Time (ns/op)", results[0].Stats[0].Type)
	s.InDelta(123, *results[0].Stats[0].Value, 0.001)
}

func (s *GoBenchmarkSuite) TestParseGoBenchmarkJSONEventErrors() {
	s.Run("malformed later event", func() {
		input := "{\"Action\":\"run\"}\nnot-json\n"
		_, _, err := ParseGoBenchmark(strings.NewReader(input), parser.Config{GroupPattern: "y"})
		s.ErrorContains(err, "read Go benchmark JSON")
	})

	s.Run("first event reader failure", func() {
		input := &dataErrorReader{
			data: []byte(`{"Action":"run"}`),
			err:  errors.New("injected event failure"),
		}
		_, _, err := ParseGoBenchmark(input, parser.Config{GroupPattern: "y"})
		s.ErrorContains(err, "read Go benchmark JSON")
	})

	s.Run("later event reader failure", func() {
		input := &dataErrorReader{
			data: []byte("{\"Action\":\"run\"}\n"),
			err:  errors.New("injected event failure"),
		}
		_, _, err := ParseGoBenchmark(input, parser.Config{GroupPattern: "y"})
		s.ErrorContains(err, "read Go benchmark JSON")
	})

	s.Run("JSON object without action remains benchmark text", func() {
		input, err := prepareBenchmarkInput(strings.NewReader(`{"Output":"not an event"}`))
		s.Require().NoError(err)
		content, err := io.ReadAll(input)
		s.Require().NoError(err)
		s.Equal(`{"Output":"not an event"}`, string(content))
	})

	s.Run("single event without trailing newline", func() {
		input := `{"Action":"output","Output":"BenchmarkExample 100 123 ns/op\n"}`
		results, _, err := ParseGoBenchmark(strings.NewReader(input), parser.Config{
			GroupPattern: "y",
			TimeUnit:     "ns",
		})
		s.Require().NoError(err)
		s.Len(results, 1)
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
			result, err := parser.ShouldIncludeBenchmark(tt.benchName, parser.Config{Filter: tt.filter})
			s.Require().NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}

func TestShouldIncludeBenchmarkSuite(t *testing.T) {
	suite.Run(t, new(ShouldIncludeBenchmarkSuite))
}
