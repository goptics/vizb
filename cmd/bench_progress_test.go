package cmd

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

// Mock progress bar for testing
type MockProgressBar struct {
	descriptions []string
}

func (m *MockProgressBar) Describe(desc string) {
	m.descriptions = append(m.descriptions, desc)
}

func (m *MockProgressBar) Finish() error {
	return nil
}

func TestHasBenchmark(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := hasBenchmark(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRawBenchmarkExtractName(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Standard benchmark line",
			line:     "BenchmarkExample-8    1000000    1234 ns/op",
			expected: "BenchmarkExample", // Extract benchmark name from result line
		},
		{
			name:     "RUN line with benchmark name",
			line:     "=== RUN   BenchmarkMemoryAllocation",
			expected: "", // RUN lines don't contain ns/op, so no extraction
		},
		{
			name:     "RUN line with benchmark name and suffix",
			line:     "=== RUN   BenchmarkStringConcat-8",
			expected: "", // RUN lines don't contain ns/op, so no extraction
		},
		{
			name:     "Empty line",
			line:     "",
			expected: "",
		},
		{
			name:     "Line with only spaces",
			line:     "   ",
			expected: "",
		},
		{
			name:     "CONT line",
			line:     "=== CONT  BenchmarkExample",
			expected: "", // CONT lines don't contain ns/op, so no extraction
		},
		{
			name:     "PAUSE line",
			line:     "=== PAUSE BenchmarkLongRunning",
			expected: "", // PAUSE lines don't contain ns/op, so no extraction
		},
		{
			name:     "Line with multiple dashes",
			line:     "=== RUN   Benchmark-Complex-Name-8",
			expected: "", // RUN lines don't contain ns/op, so no extraction
		},
		{
			name:     "Line without dash suffix",
			line:     "=== RUN   BenchmarkSimple",
			expected: "", // RUN lines don't contain ns/op, so no extraction
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := &RawBenchmark{}
			result := raw.ExtractName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONBenchmarkExtractName(t *testing.T) {
	tests := []struct {
		name     string
		event    *shared.BenchEvent
		expected string
	}{
		{
			name: "Valid benchmark event",
			event: &shared.BenchEvent{
				Action: "run",
				Test:   "BenchmarkExample",
			},
			expected: "BenchmarkExample",
		},
		{
			name: "Valid benchmark event with suffix",
			event: &shared.BenchEvent{
				Action: "run",
				Test:   "BenchmarkMemoryAllocation-8",
			},
			expected: "BenchmarkMemoryAllocation-8",
		},
		{
			name: "Non-benchmark test",
			event: &shared.BenchEvent{
				Action: "run",
				Test:   "TestExample",
			},
			expected: "",
		},
		{
			name: "Empty test name",
			event: &shared.BenchEvent{
				Action: "run",
				Test:   "",
			},
			expected: "",
		},
		{
			name:     "Nil event",
			event:    nil,
			expected: "",
		},
		{
			name: "Event with empty action",
			event: &shared.BenchEvent{
				Action: "",
				Test:   "BenchmarkExample",
			},
			expected: "BenchmarkExample",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBench := &JSONBenchmark{Event: tt.event}
			result := jsonBench.ExtractName("")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBenchmarkProgressManager(t *testing.T) {
	t.Run("NewBenchmarkProgressManager", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		assert.NotNil(t, manager)
		assert.Equal(t, mockBar, manager.bar)
		assert.Equal(t, 0, manager.benchmarkCount)
		assert.Equal(t, "", manager.currentBenchName)
	})

	t.Run("updateProgress", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		manager.currentBenchName = "BenchmarkExample"
		manager.benchmarkCount = 5

		manager.updateProgress()

		assert.Len(t, mockBar.descriptions, 1)
		assert.Contains(t, mockBar.descriptions[0], "BenchmarkExample")
		assert.Contains(t, mockBar.descriptions[0], "5 completed")
		assert.Contains(t, mockBar.descriptions[0], "Running Benchmarks")
	})

	t.Run("ProcessLine with JSON", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		// Test JSON line with benchmark start
		jsonLine := `{"Action":"run","Test":"BenchmarkMemoryAlloc"}`
		manager.ProcessLine(jsonLine)

		assert.Equal(t, "BenchmarkMemoryAlloc", manager.currentBenchName)
		assert.Equal(t, 0, manager.benchmarkCount)     // No ns/op yet
		assert.True(t, len(mockBar.descriptions) >= 1) // Should have at least one update

		// Test JSON line with benchmark result
		resultLine := `{"Action":"pass","Test":"BenchmarkMemoryAlloc","Output":"1000 ns/op"}`
		manager.ProcessLine(resultLine)

		assert.Equal(t, 1, manager.benchmarkCount)     // Should increment
		assert.True(t, len(mockBar.descriptions) >= 1) // Should have progress updates
	})

	t.Run("ProcessLine with Raw text", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		// Test RUN line - doesn't extract name (no ns/op)
		runLine := "=== RUN   BenchmarkStringConcat-8"
		manager.ProcessLine(runLine)

		assert.Equal(t, "", manager.currentBenchName) // RUN lines don't extract names
		assert.Equal(t, 0, manager.benchmarkCount)
		assert.Len(t, mockBar.descriptions, 0) // No progress update

		// Test benchmark result line
		resultLine := "BenchmarkStringConcat-8    1000000    1234 ns/op"
		manager.ProcessLine(resultLine)

		assert.Equal(t, 1, manager.benchmarkCount)                         // Should increment
		assert.Equal(t, "BenchmarkStringConcat", manager.currentBenchName) // Extract name from result line
		assert.Len(t, mockBar.descriptions, 1)                             // Progress update
	})

	t.Run("ProcessLine with mixed content", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

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

		// Last name extracted would be from the last benchmark result line
		assert.Equal(t, "BenchmarkThird", manager.currentBenchName) // From the last benchmark result line
		assert.Equal(t, 3, manager.benchmarkCount)
		assert.True(t, len(mockBar.descriptions) >= 2) // Should have multiple updates
	})

	t.Run("ProcessLine with invalid JSON", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		// Invalid JSON should be treated as raw text
		invalidJSON := `{"Action":"run","Test":invalid}`
		manager.ProcessLine(invalidJSON)

		// Should not panic and should process as raw text
		assert.Equal(t, 0, manager.benchmarkCount)
		// Invalid JSON treated as raw text doesn't contain ns/op, so no name extraction
		assert.Equal(t, "", manager.currentBenchName) // No extraction since no ns/op
	})

	t.Run("ProcessLine with empty lines", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		lines := []string{
			"",
			"   ",
			"\n",
			"\t",
		}

		for _, line := range lines {
			manager.ProcessLine(line)
		}

		assert.Equal(t, 0, manager.benchmarkCount)
		assert.Equal(t, "", manager.currentBenchName)
		assert.Len(t, mockBar.descriptions, 0) // No progress updates
	})

	t.Run("ProcessLine increments benchmark count correctly", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		benchmarkLines := []string{
			"BenchmarkA-8    1000    1000 ns/op",
			"BenchmarkB-8    2000    2000 ns/op",
			"Some line with ns/op but not benchmark result",
			"BenchmarkC-4    3000    3000 ns/op",
		}

		for _, line := range benchmarkLines {
			manager.ProcessLine(line)
		}

		// Should count all lines containing "ns/op"
		assert.Equal(t, 4, manager.benchmarkCount)
	})
}

// Test edge cases and integration scenarios
func TestBenchmarkProgressIntegration(t *testing.T) {
	t.Run("Real world benchmark output", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

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

		// Should have processed multiple benchmarks
		assert.True(t, manager.benchmarkCount >= 4, "Should have found at least 4 benchmark results")
		assert.Equal(t, "BenchmarkSliceAppend", manager.currentBenchName) // Last benchmark result line
		assert.True(t, len(mockBar.descriptions) > 0, "Should have progress updates")
	})

	t.Run("JSON benchmark output", func(t *testing.T) {
		mockBar := &MockProgressBar{}
		manager := NewBenchmarkProgressManager(mockBar)

		// Simulate go test -bench=. -json output
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

		assert.Equal(t, "BenchmarkExample", manager.currentBenchName)
		assert.True(t, manager.benchmarkCount >= 1, "Should have found benchmark results")
	})
}
