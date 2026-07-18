package parser

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goptics/vizb/shared"
)

var (
	detectAnsiRe   = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	divanRe        = regexp.MustCompile(`^[├╰]─\s+\S+`)
	criterionRe    = regexp.MustCompile(`\S+\s+time:\s+\[`)
	tinybenchRowRe = regexp.MustCompile(`│\s*\d+\s*│`)
	goBenchRe      = regexp.MustCompile(`^Benchmark\S*\s+\d+`)
)

// DetectParser inspects the file's content (not just its extension) and returns
// the best-matching parser key. It never fails: when nothing matches it falls
// back to "go" (the historical default). Signatures are tested strong→weak.
func DetectParser(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return "go"
	}
	defer f.Close()

	csvHint := strings.EqualFold(filepath.Ext(filename), ".csv") || looksLikeCSV(filename)
	return detectParser(f, csvHint)
}

// DetectParserBytes detects a parser from inline content without relying on a
// filename or extension. It shares the same signature ordering as DetectParser.
func DetectParserBytes(data []byte) string {
	return detectParser(bytes.NewReader(data), looksLikeCSVReader(bytes.NewReader(data)))
}

func detectParser(input io.Reader, csvHint bool) string {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		firstNonEmpty string
		sawGoJSON     bool
		sawDivan      bool
		sawCriterion  bool
		sawTinybench  bool
		sawVitest     bool
		sawGoText     bool
	)

	for lines := 0; scanner.Scan() && lines < 200; lines++ {
		line := detectAnsiRe.ReplaceAllString(scanner.Text(), "")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if firstNonEmpty == "" {
			firstNonEmpty = trimmed
			var event shared.BenchEvent
			sawGoJSON = json.Unmarshal([]byte(trimmed), &event) == nil && event.Action != ""
		}

		switch {
		case divanRe.MatchString(trimmed),
			strings.Contains(line, "fastest") && strings.Contains(line, "slowest") && strings.Contains(line, "median"):
			sawDivan = true
		}

		if criterionRe.MatchString(trimmed) {
			sawCriterion = true
		}
		if strings.Contains(line, "Task name") || tinybenchRowRe.MatchString(line) {
			sawTinybench = true
		}
		if strings.HasPrefix(trimmed, "·") && len(strings.Fields(trimmed)) >= 11 {
			sawVitest = true
		}
		if strings.Contains(line, "ns/op") ||
			strings.HasPrefix(trimmed, "goos:") || strings.HasPrefix(trimmed, "goarch:") ||
			strings.HasPrefix(trimmed, "pkg:") || strings.HasPrefix(trimmed, "cpu:") ||
			goBenchRe.MatchString(trimmed) {
			sawGoText = true
		}
	}

	// Scan errors are non-fatal: detection is best-effort on the sampled prefix
	// and falls back to "go" below regardless.
	_ = scanner.Err()

	// 1. Go benchmark -json event stream (JSONL with an "Action" field).
	if sawGoJSON {
		return "go"
	}

	// 2. Generic JSON array.
	if strings.HasPrefix(firstNonEmpty, "[") {
		return "json"
	}

	// 3-6. Tool-specific text formats (strong markers).
	switch {
	case sawDivan:
		return "rs:divan"
	case sawCriterion:
		return "rs:criterion"
	case sawVitest:
		return "js:vitest"
	case sawTinybench:
		return "js:tinybench"
	}

	// 7. CSV — extension hint or structural sniff.
	if csvHint {
		return "csv"
	}

	// 8-9. Go benchmark text, else fallback.
	if sawGoText {
		return "go"
	}

	return "go"
}

// looksLikeCSV reports whether the file parses as comma-separated rows with at
// least two columns in each of the first two records.
func looksLikeCSV(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()

	return looksLikeCSVReader(f)
}

func looksLikeCSVReader(input io.Reader) bool {
	r := csv.NewReader(input)
	r.FieldsPerRecord = -1

	first, err := r.Read()
	if err != nil || len(first) < 2 {
		return false
	}

	second, err := r.Read()
	if err != nil || len(second) < 2 {
		return false
	}

	return true
}
