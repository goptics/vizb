package rust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

var testCargoTable = `running 0 tests

test result: ok. 0 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s

bubbleSort/n=100        time:   [3.0424 µs 3.0524 µs 3.0637 µs]
                        change: [+3.5450% +4.5929% +5.7731%] (p = 0.00 < 0.05)
                        Performance has regressed.
Found 12 outliers among 100 measurements (12.00%)
  3 (3.00%) high mild
  9 (9.00%) high severe

insertionSort/n=100     time:   [819.49 ns 821.61 ns 824.50 ns]
                        change: [+31.070% +32.526% +33.878%] (p = 0.00 < 0.05)
                        Performance has regressed.
Found 22 outliers among 100 measurements (22.00%)
  1 (1.00%) low severe
  11 (11.00%) low mild
  2 (2.00%) high mild
  8 (8.00%) high severe

bubbleSort/n=500        time:   [104.13 µs 104.31 µs 104.48 µs]
                        change: [+0.5028% +0.6664% +0.8369%] (p = 0.00 < 0.05)
                        Change within noise threshold.
Found 8 outliers among 100 measurements (8.00%)
  4 (4.00%) low severe
  3 (3.00%) low mild
  1 (1.00%) high severe

insertionSort/n=500     time:   [21.731 µs 21.780 µs 21.836 µs]
                        change: [+1.2200% +1.5261% +1.8205%] (p = 0.00 < 0.05)
                        Performance has regressed.

bubbleSort/n=2000       time:   [1.3804 ms 1.3827 ms 1.3851 ms]
                        change: [+1.1242% +1.8427% +2.9253%] (p = 0.00 < 0.05)
                        Performance has regressed.
Found 2 outliers among 100 measurements (2.00%)
  2 (2.00%) high severe

insertionSort/n=2000    time:   [299.74 µs 300.19 µs 300.69 µs]
                        change: [+0.9356% +1.2778% +1.6400%] (p = 0.00 < 0.05)
                        Change within noise threshold.
Found 3 outliers among 100 measurements (3.00%)
  3 (3.00%) high mild
`

func writeTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertStat(t *testing.T, s shared.Stat, expectedType string, expectedValue float64, expectedSymbol string) {
	t.Helper()
	assert.Equal(t, expectedType, s.Type, "stat type mismatch")
	assert.Equal(t, expectedValue, s.Value, "stat value mismatch")
	assert.Equal(t, expectedSymbol, s.Symbol, "stat symbol mismatch")
}

func TestParseCriterionBenchmark(t *testing.T) {
	origPattern := shared.FlagState.GroupPattern
	origFilter := shared.FlagState.FilterRegex
	origTimeUnit := shared.FlagState.TimeUnit
	defer func() {
		shared.FlagState.GroupPattern = origPattern
		shared.FlagState.FilterRegex = origFilter
		shared.FlagState.TimeUnit = origTimeUnit
	}()

	t.Run("Real cargo criterion output", func(t *testing.T) {
		shared.FlagState.GroupPattern = "n/y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "ns"

		results := ParseCriterionBenchmark(writeTestFile(t, testCargoTable))
		assert.Len(t, results, 6)

		// First: bubbleSort/n=100 → 3.0524 µs = 3052.4 ns
		assertStat(t, results[0].Stats[0], "Latency avg (ns)", 3052.4, "")
		assertStat(t, results[0].Stats[1], "Latency lower (ns)", 3042.4, "")
		assertStat(t, results[0].Stats[2], "Latency upper (ns)", 3063.7, "±")

		assert.Equal(t, "bubbleSort", results[0].Name)
		assert.Equal(t, "n=100", results[0].YAxis)

		// Second: insertionSort/n=100 → 821.61 ns (already in ns)
		assertStat(t, results[1].Stats[0], "Latency avg (ns)", 821.61, "")
		assertStat(t, results[1].Stats[1], "Latency lower (ns)", 819.49, "")
		assertStat(t, results[1].Stats[2], "Latency upper (ns)", 824.5, "±")

		// Fifth: bubbleSort/n=2000 → 1.3827 ms = 1382700 ns
		assertStat(t, results[4].Stats[0], "Latency avg (ns)", 1382700, "")

		// Last: insertionSort/n=2000 → 300.19 µs = 300190 ns
		assertStat(t, results[5].Stats[0], "Latency avg (ns)", 300190, "")
		assert.Equal(t, "insertionSort", results[5].Name)
		assert.Equal(t, "n=2000", results[5].YAxis)
	})

	t.Run("Unit conversion to us", func(t *testing.T) {
		shared.FlagState.GroupPattern = "n/y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "us"

		results := ParseCriterionBenchmark(writeTestFile(t, testCargoTable))
		assert.Len(t, results, 6)

		assertStat(t, results[0].Stats[0], "Latency avg (us)", 3.05, "")
		assertStat(t, results[1].Stats[0], "Latency avg (us)", 0.82, "")
		assertStat(t, results[4].Stats[0], "Latency avg (us)", 1382.7, "")
	})

	t.Run("Filter regex", func(t *testing.T) {
		shared.FlagState.GroupPattern = "n/y"
		shared.FlagState.FilterRegex = "bubbleSort"
		shared.FlagState.TimeUnit = "ns"

		results := ParseCriterionBenchmark(writeTestFile(t, testCargoTable))
		assert.Len(t, results, 3)
		for _, r := range results {
			assert.Equal(t, "bubbleSort", r.Name)
		}
	})

	t.Run("Empty file", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "ns"

		results := ParseCriterionBenchmark(writeTestFile(t, ""))
		assert.Empty(t, results)
	})
}
