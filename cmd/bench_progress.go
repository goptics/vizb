package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/shared"
)

func hasBenchmark(line string) bool {
	return strings.Contains(line, "ns/op")
}

type BenchmarkLine interface {
	ExtractName(string) string
}

// Base implementation for raw benchmarks
type RawBenchmark struct{}

func (r *RawBenchmark) ExtractName(line string) string {
	if !hasBenchmark(line) {
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

// JSONBenchmark extends RawBenchmark but overrides ExtractName
type JSONBenchmark struct {
	Event *shared.BenchEvent
}

func (j *JSONBenchmark) ExtractName(_ string) string {
	if j.Event != nil && j.Event.Test != "" && strings.HasPrefix(j.Event.Test, "Benchmark") {
		return j.Event.Test
	}

	return ""
}

// ProgressBar is a small interface for dependency injection
type ProgressBar interface {
	Describe(string)
	Finish() error
}

// BenchmarkProgressManager holds state + orchestrates benchmark processing
type BenchmarkProgressManager struct {
	bar              ProgressBar
	benchmarkCount   int
	currentBenchName string
}

// NewBenchmarkProgressManager creates a new instance of BenchmarkProgressManager
// with the provided progress bar interface for displaying benchmark execution progress.
func NewBenchmarkProgressManager(bar ProgressBar) *BenchmarkProgressManager {
	return &BenchmarkProgressManager{bar: bar}
}

func (m *BenchmarkProgressManager) updateProgress() {
	desc := fmt.Sprintf(
		"Running Benchmarks [%s] (%d completed)",
		m.currentBenchName,
		m.benchmarkCount,
	)

	m.bar.Describe(style.Info.Render(desc))
}

func (m *BenchmarkProgressManager) Finish() error {
	return m.bar.Finish()
}

func (m *BenchmarkProgressManager) ProcessLine(line string) {
	var ev shared.BenchEvent
	var parser BenchmarkLine

	if err := json.Unmarshal([]byte(line), &ev); err == nil {
		parser = &JSONBenchmark{Event: &ev}
	} else {
		parser = &RawBenchmark{}
	}

	if hasBenchmark(line) {
		m.benchmarkCount++
	}

	if name := parser.ExtractName(line); name != "" {
		m.currentBenchName = name
		m.updateProgress()
	}
}
