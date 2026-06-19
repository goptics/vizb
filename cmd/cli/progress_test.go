package cli

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// MockProgressBar records the descriptions pushed to it.
type MockProgressBar struct {
	descriptions []string
}

func (m *MockProgressBar) Describe(desc string) {
	m.descriptions = append(m.descriptions, desc)
}

func (m *MockProgressBar) Finish() error {
	return nil
}

// ProgressSuite covers the benchmark progress manager and its line parsers.
type ProgressSuite struct {
	suite.Suite
}

func (s *ProgressSuite) TestHasBenchmark() {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"Line with ns/op", "BenchmarkExample-8    1000000    1234 ns/op", true},
		{"Line without ns/op", "=== RUN   BenchmarkExample", false},
		{"Empty line", "", false},
		{"JSON line with ns/op", `{"Action":"pass","Test":"BenchmarkExample-8","Output":"1000 ns/op"}`, true},
		{"Regular text with ns/op", "Some random text with ns/op in it", true},
		{"Line with just numbers", "123456789", false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, hasBenchmark(tt.line))
		})
	}
}

func (s *ProgressSuite) TestRawBenchmarkExtractName() {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"Standard benchmark line", "BenchmarkExample-8    1000000    1234 ns/op", "BenchmarkExample"},
		{"RUN line with benchmark name", "=== RUN   BenchmarkMemoryAllocation", ""},
		{"RUN line with benchmark name and suffix", "=== RUN   BenchmarkStringConcat-8", ""},
		{"Empty line", "", ""},
		{"Line with only spaces", "   ", ""},
		{"CONT line", "=== CONT  BenchmarkExample", ""},
		{"PAUSE line", "=== PAUSE BenchmarkLongRunning", ""},
		{"Line with multiple dashes", "=== RUN   Benchmark-Complex-Name-8", ""},
		{"Line without dash suffix", "=== RUN   BenchmarkSimple", ""},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			raw := &RawDataLine{}
			s.Equal(tt.expected, raw.ExtractName(tt.line))
		})
	}
}

func (s *ProgressSuite) TestJSONBenchmarkExtractName() {
	tests := []struct {
		name     string
		event    *shared.BenchEvent
		expected string
	}{
		{"Valid benchmark event", &shared.BenchEvent{Action: "run", Test: "BenchmarkExample"}, "BenchmarkExample"},
		{"Valid benchmark event with suffix", &shared.BenchEvent{Action: "run", Test: "BenchmarkMemoryAllocation-8"}, "BenchmarkMemoryAllocation-8"},
		{"Non-benchmark test", &shared.BenchEvent{Action: "run", Test: "TestExample"}, ""},
		{"Empty test name", &shared.BenchEvent{Action: "run", Test: ""}, ""},
		{"Nil event", nil, ""},
		{"Event with empty action", &shared.BenchEvent{Action: "", Test: "BenchmarkExample"}, "BenchmarkExample"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			jsonBench := &JSONDataLine{Event: tt.event}
			s.Equal(tt.expected, jsonBench.ExtractName(""))
		})
	}
}

func (s *ProgressSuite) TestNewDataProgressManager() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	s.NotNil(manager)
	s.Equal(mockBar, manager.bar)
	s.Equal(0, manager.dataCount)
	s.Equal("", manager.currentDataName)
}

func (s *ProgressSuite) TestUpdateProgress() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	manager.currentDataName = "BenchmarkExample"
	manager.dataCount = 5
	manager.updateProgress()

	s.Len(mockBar.descriptions, 1)
	s.Contains(mockBar.descriptions[0], "BenchmarkExample")
	s.Contains(mockBar.descriptions[0], "5 records")
	s.Contains(mockBar.descriptions[0], "Processing Data")
}

func (s *ProgressSuite) TestProcessLineWithJSON() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	manager.ProcessLine(`{"Action":"run","Test":"BenchmarkMemoryAlloc"}`)
	s.Equal("BenchmarkMemoryAlloc", manager.currentDataName)
	s.Equal(0, manager.dataCount)
	s.GreaterOrEqual(len(mockBar.descriptions), 1)

	manager.ProcessLine(`{"Action":"pass","Test":"BenchmarkMemoryAlloc","Output":"1000 ns/op"}`)
	s.Equal(1, manager.dataCount)
}

func (s *ProgressSuite) TestProcessLineWithRawText() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	manager.ProcessLine("=== RUN   BenchmarkStringConcat-8")
	s.Equal("", manager.currentDataName)
	s.Equal(0, manager.dataCount)
	s.Len(mockBar.descriptions, 0)

	manager.ProcessLine("BenchmarkStringConcat-8    1000000    1234 ns/op")
	s.Equal(1, manager.dataCount)
	s.Equal("BenchmarkStringConcat", manager.currentDataName)
	s.Len(mockBar.descriptions, 1)
}

func (s *ProgressSuite) TestProcessLineWithMixedContent() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	lines := []string{
		"=== RUN   BenchmarkFirst",
		"BenchmarkFirst-8    1000    1000 ns/op",
		`{"Action":"run","Test":"BenchmarkSecond"}`,
		`{"Action":"pass","Test":"BenchmarkSecond","Output":"2000 ns/op"}`,
		"=== RUN   BenchmarkThird-4",
		"BenchmarkThird-4    500    3000 ns/op",
	}
	for _, line := range lines {
		manager.ProcessLine(line)
	}

	s.Equal("BenchmarkThird", manager.currentDataName)
	s.Equal(3, manager.dataCount)
	s.GreaterOrEqual(len(mockBar.descriptions), 2)
}

func (s *ProgressSuite) TestProcessLineWithInvalidJSON() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	manager.ProcessLine(`{"Action":"run","Test":invalid}`)
	s.Equal(0, manager.dataCount)
	s.Equal("", manager.currentDataName)
}

func (s *ProgressSuite) TestProcessLineWithEmptyLines() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	for _, line := range []string{"", "   ", "\n", "\t"} {
		manager.ProcessLine(line)
	}

	s.Equal(0, manager.dataCount)
	s.Equal("", manager.currentDataName)
	s.Len(mockBar.descriptions, 0)
}

func (s *ProgressSuite) TestProcessLineIncrementsBenchmarkCount() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	for _, line := range []string{
		"BenchmarkA-8    1000    1000 ns/op",
		"BenchmarkB-8    2000    2000 ns/op",
		"Some line with ns/op but not benchmark result",
		"BenchmarkC-4    3000    3000 ns/op",
	} {
		manager.ProcessLine(line)
	}

	s.Equal(4, manager.dataCount)
}

func (s *ProgressSuite) TestRealWorldBenchmarkOutput() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	realOutput := []string{
		"goos: linux",
		"goarch: amd64",
		"pkg: github.com/example/benchmarks",
		"cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz",
		"=== RUN   BenchmarkStringBuilder",
		"=== RUN   BenchmarkStringBuilder/small",
		"=== RUN   BenchmarkStringBuilder/medium",
		"=== RUN   BenchmarkStringBuilder/large",
		"BenchmarkStringBuilder/small-12         3000000      456 ns/op",
		"BenchmarkStringBuilder/medium-12        1000000     1234 ns/op",
		"BenchmarkStringBuilder/large-12          500000     2345 ns/op",
		"=== RUN   BenchmarkSliceAppend",
		"BenchmarkSliceAppend-12                  2000000      789 ns/op",
		"PASS",
		"ok      github.com/example/benchmarks    4.567s",
	}
	for _, line := range realOutput {
		manager.ProcessLine(line)
	}

	s.GreaterOrEqual(manager.dataCount, 4)
	s.Equal("BenchmarkSliceAppend", manager.currentDataName)
	s.Greater(len(mockBar.descriptions), 0)
}

func (s *ProgressSuite) TestJSONBenchmarkOutput() {
	mockBar := &MockProgressBar{}
	manager := NewDataProgressManager(mockBar)

	jsonEvents := []shared.BenchEvent{
		{Action: "start", Test: ""},
		{Action: "run", Test: "BenchmarkExample"},
		{Action: "output", Test: "BenchmarkExample", Output: "BenchmarkExample-8   "},
		{Action: "output", Test: "BenchmarkExample", Output: "1000000"},
		{Action: "output", Test: "BenchmarkExample", Output: "1234 ns/op"},
		{Action: "pass", Test: "BenchmarkExample"},
	}
	for _, event := range jsonEvents {
		jsonLine, _ := json.Marshal(event)
		manager.ProcessLine(string(jsonLine))
	}

	s.Equal("BenchmarkExample", manager.currentDataName)
	s.GreaterOrEqual(manager.dataCount, 1)
}

func TestProgressSuite(t *testing.T) {
	suite.Run(t, new(ProgressSuite))
}
