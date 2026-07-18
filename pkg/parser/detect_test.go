package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DetectSuite struct {
	suite.Suite
}

func writeDetectFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

const (
	goTextSample = "goos: linux\ngoarch: amd64\npkg: x/y\ncpu: Test CPU\nBenchmarkFoo-8   \t1000000\t       123 ns/op\nPASS\n"

	goJSONSample = `{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"x/y","Output":"BenchmarkFoo-8 100 12 ns/op\n"}` + "\n"

	divanSample = "sort_divan              fastest       │ slowest       │ median        │ mean          │ samples │ iters\n" +
		"├─ bubble_sort_100      4.36 µs       │ 9.68 µs       │ 4.646 µs      │ 4.733 µs      │ 100     │ 100\n" +
		"╰─ insertion_sort_100   1.131 µs      │ 1.423 µs      │ 1.139 µs      │ 1.149 µs      │ 100     │ 200\n"

	criterionSample = "fib_10                  time:   [21.234 ns 21.456 ns 21.678 ns]\n" +
		"                        change: [-1.0% +0.5% +2.0%]\n"

	tinybenchSample = "┌─────────┬───────────┬──────────────────┐\n" +
		"│ (index) │ Task name │ Latency avg (ns) │\n" +
		"├─────────┼───────────┼──────────────────┤\n" +
		"│ 0       │ 'foo'     │ '1234'           │\n" +
		"└─────────┴───────────┴──────────────────┘\n"

	vitestSample = " · foo 1234 0.1 0.2 0.3 0.4 0.5 0.6 0.7 ±1.5% 100\n"

	csvSample = "name,sells,stocks\na,10,5\nb,20,7\n"

	jsonArraySample = `[{"name":"a","sells":10},{"name":"b","sells":20}]`

	vizbBenchmarkSample = `{"name":"Benchmarks","data":[{"name":"a","stats":[]}]}`
)

func (s *DetectSuite) TestDetectParser() {
	t := s.T()
	cases := []struct {
		name    string
		file    string
		content string
		want    string
	}{
		{"go text", "bench.txt", goTextSample, "go"},
		{"go json events", "bench.json", goJSONSample, "go"},
		{"rust divan", "divan.txt", divanSample, "rs:divan"},
		{"rust criterion", "criterion.txt", criterionSample, "rs:criterion"},
		{"js tinybench", "tiny.txt", tinybenchSample, "js:tinybench"},
		{"js vitest", "vitest.txt", vitestSample, "js:vitest"},
		{"csv by content", "data.txt", csvSample, "csv"},
		{"csv by extension", "data.csv", "name;only;semicolons\n", "csv"},
		{"json array by content", "data.txt", jsonArraySample, "json"},
		{"vizb benchmark json falls back to go", "out.json", vizbBenchmarkSample, "go"},
		{"empty falls back to go", "empty.txt", "", "go"},
		{"garbage falls back to go", "junk.txt", "just some random text\nwith no markers\n", "go"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeDetectFile(t, tc.file, tc.content)
			assert.Equal(t, tc.want, DetectParser(path))
		})
	}
}

func (s *DetectSuite) TestLooksLikeCSVEdgeCases() {
	t := s.T()
	t.Run("single line is not CSV", func(t *testing.T) {
		path := writeDetectFile(t, "one.txt", "just one line")
		assert.Equal(t, "go", DetectParser(path))
	})
	t.Run("header only without data rows is not CSV", func(t *testing.T) {
		path := writeDetectFile(t, "header.txt", "a,b,c\n")
		assert.Equal(t, "go", DetectParser(path))
	})
}

func (s *DetectSuite) TestDetectParserBytes() {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"go text", goTextSample, "go"},
		{"go json events", goJSONSample, "go"},
		{"rust divan", divanSample, "rs:divan"},
		{"rust criterion", criterionSample, "rs:criterion"},
		{"js tinybench", tinybenchSample, "js:tinybench"},
		{"js vitest", vitestSample, "js:vitest"},
		{"csv", csvSample, "csv"},
		{"json", jsonArraySample, "json"},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.Equal(tc.want, DetectParserBytes([]byte(tc.content)))
		})
	}
}

func TestDetectSuite(t *testing.T) {
	suite.Run(t, new(DetectSuite))
}
