package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goptics/vizb/shared"
)

type BenchmarkLine interface {
	ExtractName(line string) string
	HasBenchmark(line string) bool
}

// Base implementation for raw benchmarks
type RawBenchmark struct{}

func (r *RawBenchmark) ExtractName(line string) string {
	if !r.HasBenchmark(line) {
		return ""
	}

	fields := strings.Fields(line)

	if len(fields) == 0 {
		return ""
	}

	name := fields[0]

	// Strip trailing `-<digits>` if present (e.g., "-8")
	if i := strings.LastIndex(name, "-"); i != -1 {
		return name[:i]
	}

	return name
}

func (r *RawBenchmark) HasBenchmark(line string) bool {
	return strings.Contains(line, "ns/op")
}

// JSONBenchmark extends RawBenchmark but overrides ExtractName
type JSONBenchmark struct {
	RawBenchmark
	Event *shared.BenchEvent
}

func (j *JSONBenchmark) ExtractName(line string) string {
	if j.Event != nil && j.Event.Test != "" && strings.HasPrefix(j.Event.Test, "Benchmark") {
		return j.Event.Test
	}

	return ""
}

// ProgressBar is a small interface for dependency injection
type ProgressBar interface {
	Describe(string)
}

// BenchmarkProgressManager holds state + orchestrates benchmark processing
type BenchmarkProgressManager struct {
	bar              ProgressBar
	benchmarkCount   int
	currentBenchName string
}

func NewBenchmarkProgressManager(bar ProgressBar) *BenchmarkProgressManager {
	return &BenchmarkProgressManager{bar: bar}
}

func (m *BenchmarkProgressManager) updateProgress() {
	desc := fmt.Sprintf(
		"Running Benchmarks [%s] (%d completed)",
		m.currentBenchName,
		m.benchmarkCount,
	)

	m.bar.Describe(desc)
}

func (m *BenchmarkProgressManager) ProcessLine(line string) {
	var ev shared.BenchEvent
	var parser BenchmarkLine

	if err := json.Unmarshal([]byte(line), &ev); err == nil {
		parser = &JSONBenchmark{Event: &ev}
	} else {
		parser = &RawBenchmark{}
	}

	if parser.HasBenchmark(line) {
		m.benchmarkCount++
	}

	if name := parser.ExtractName(line); name != "" {
		m.currentBenchName = name
		m.updateProgress()
	}
}
