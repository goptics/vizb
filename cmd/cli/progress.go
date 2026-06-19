package cli

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

type DataLine interface {
	ExtractName(string) string
}

// RawDataLine is the base implementation for raw benchmark text.
type RawDataLine struct{}

func (r *RawDataLine) ExtractName(line string) string {
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

// JSONDataLine extends RawDataLine but overrides ExtractName.
type JSONDataLine struct {
	Event *shared.BenchEvent
}

func (j *JSONDataLine) ExtractName(_ string) string {
	if j.Event != nil && j.Event.Test != "" && strings.HasPrefix(j.Event.Test, "Benchmark") {
		return j.Event.Test
	}

	return ""
}

// ProgressBar is a small interface for dependency injection.
type ProgressBar interface {
	Describe(string)
	Finish() error
}

// DataProgressManager holds state + orchestrates data processing progress display.
type DataProgressManager struct {
	bar             ProgressBar
	dataCount       int
	currentDataName string
}

// NewDataProgressManager creates a new instance of DataProgressManager
// with the provided progress bar interface for displaying data processing progress.
func NewDataProgressManager(bar ProgressBar) *DataProgressManager {
	return &DataProgressManager{bar: bar}
}

func (m *DataProgressManager) updateProgress() {
	desc := fmt.Sprintf(
		"Processing Data [%s] (%d records)",
		m.currentDataName,
		m.dataCount,
	)

	m.bar.Describe(style.Info.Render(desc))
}

func (m *DataProgressManager) Finish() error {
	return m.bar.Finish()
}

func (m *DataProgressManager) ProcessLine(line string) {
	var ev shared.BenchEvent
	var p DataLine

	if err := json.Unmarshal([]byte(line), &ev); err == nil {
		p = &JSONDataLine{Event: &ev}
	} else {
		p = &RawDataLine{}
	}

	if hasBenchmark(line) {
		m.dataCount++
	}

	if name := p.ExtractName(line); name != "" {
		m.currentDataName = name
		m.updateProgress()
	}
}
