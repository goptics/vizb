package rust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

var testDivanTable = `sort_divan              fastest       │ slowest       │ median        │ mean          │ samples │ iters
├─ bubble_sort_100      4.36 µs       │ 9.68 µs       │ 4.646 µs      │ 4.733 µs      │ 100     │ 100
├─ bubble_sort_500      103.7 µs      │ 174.6 µs      │ 108 µs        │ 119.1 µs      │ 100     │ 100
├─ bubble_sort_2000     1.36 ms       │ 2.005 ms      │ 1.391 ms      │ 1.458 ms      │ 100     │ 100
├─ insertion_sort_100   1.131 µs      │ 1.423 µs      │ 1.139 µs      │ 1.149 µs      │ 100     │ 200
├─ insertion_sort_500   20.86 µs      │ 23.8 µs       │ 21.42 µs      │ 21.39 µs      │ 100     │ 100
╰─ insertion_sort_2000  289 µs        │ 299.9 µs      │ 292.8 µs      │ 292.4 µs      │ 100     │ 100
`

func writeDivanTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseDivanBenchmark(t *testing.T) {
	origPattern := shared.FlagState.GroupPattern
	origFilter := shared.FlagState.FilterRegex
	origTimeUnit := shared.FlagState.TimeUnit
	defer func() {
		shared.FlagState.GroupPattern = origPattern
		shared.FlagState.FilterRegex = origFilter
		shared.FlagState.TimeUnit = origTimeUnit
	}()

	t.Run("Real divan output", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "ns"

		results := ParseDivanBenchmark(writeDivanTestFile(t, testDivanTable))
		assert.Len(t, results, 6)

		assertStat(t, results[0].Stats[0], "Latency fastest (ns)", 4360, "")
		assertStat(t, results[0].Stats[1], "Latency slowest (ns)", 9680, "±")
		assertStat(t, results[0].Stats[2], "Latency median (ns)", 4646, "")
		assertStat(t, results[0].Stats[3], "Latency mean (ns)", 4733, "")
		assertStat(t, results[0].Stats[4], "Samples", 100, "")

		assert.Equal(t, "bubble_sort_100", results[0].YAxis)

		assertStat(t, results[2].Stats[0], "Latency fastest (ns)", 1360000, "")
		assertStat(t, results[2].Stats[1], "Latency slowest (ns)", 2005000, "±")

		assertStat(t, results[5].Stats[3], "Latency mean (ns)", 292400, "")
		assert.Equal(t, "insertion_sort_2000", results[5].YAxis)
	})

	t.Run("Unit conversion to us", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "us"

		results := ParseDivanBenchmark(writeDivanTestFile(t, testDivanTable))
		assert.Len(t, results, 6)

		assertStat(t, results[0].Stats[0], "Latency fastest (us)", 4.36, "")
		assertStat(t, results[0].Stats[3], "Latency mean (us)", 4.73, "")
		assertStat(t, results[2].Stats[0], "Latency fastest (us)", 1360, "")
	})

	t.Run("Filter regex", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y"
		shared.FlagState.FilterRegex = "insertion"
		shared.FlagState.TimeUnit = "ns"

		results := ParseDivanBenchmark(writeDivanTestFile(t, testDivanTable))
		assert.Len(t, results, 3)
		for _, r := range results {
			assert.Contains(t, r.YAxis, "insertion")
		}
	})

	t.Run("Empty file", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "ns"

		results := ParseDivanBenchmark(writeDivanTestFile(t, ""))
		assert.Empty(t, results)
	})
}
