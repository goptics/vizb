package javascript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
)

var testVitestTable = ` ✓ sort.bench.js > n=100 1356ms
     name                   hz     min     max    mean     p75     p99    p995    p999     rme  samples
   · bubbleSort     264,413.96  0.0037  0.0974  0.0038  0.0038  0.0049  0.0060  0.0076  ±0.09%   132207
   · insertionSort  639,846.68  0.0015  0.1212  0.0016  0.0016  0.0024  0.0025  0.0034  ±0.14%   319924

 ✓ sort.bench.js > n=500 1212ms
     name                  hz     min     max    mean     p75     p99    p995    p999     rme  samples
   · bubbleSort      7,715.01  0.1235  1.6623  0.1296  0.1298  0.1663  0.1713  0.1744  ±0.61%     3858
   · insertionSort  30,632.50  0.0320  0.1858  0.0326  0.0329  0.0345  0.0357  0.0441  ±0.09%    15317

 ✓ sort.bench.js > n=2000 1211ms
     name                 hz     min     max    mean     p75     p99    p995    p999     rme  samples
   · bubbleSort       581.49  1.6942  1.8435  1.7197  1.7342  1.8244  1.8252  1.8435  ±0.15%      291
   · insertionSort  1,933.38  0.5139  0.5336  0.5172  0.5181  0.5278  0.5288  0.5336  ±0.05%      967
`

func TestParseVitestBenchmark(t *testing.T) {
	origPattern := shared.FlagState.GroupPattern
	origFilter := shared.FlagState.FilterRegex
	defer func() {
		shared.FlagState.GroupPattern = origPattern
		shared.FlagState.FilterRegex = origFilter
	}()

	t.Run("Real vitest output from sort.bench.js", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y/n"
		shared.FlagState.FilterRegex = ""

		results := ParseVitestBenchmark(writeVitestTestFile(t, testVitestTable))

		if len(results) == 0 {
			t.Fatal("expected results, got none")
		}

		// Expect 6 rows: 2 sort algorithms x 3 sizes
		if len(results) != 6 {
			t.Fatalf("expected 6 results, got %d", len(results))
		}

		// First row: n=100/bubbleSort → name=bubbleSort, yAxis=n=100
		assertStat(t, results[0].Stats[0], "Throughput avg (ops/s)", 264413.96, "")
		assertStat(t, results[0].Stats[1], "Latency min (ms)", 0.0037, "")
		assertStat(t, results[0].Stats[2], "Latency max (ms)", 0.0974, "")
		assertStat(t, results[0].Stats[3], "Latency avg (ms)", 0.0038, "")
		assertStat(t, results[0].Stats[4], "Latency p75 (ms)", 0.0038, "")
		assertStat(t, results[0].Stats[5], "Latency p99 (ms)", 0.0049, "")
		assertStat(t, results[0].Stats[6], "Latency p995 (ms)", 0.006, "")
		assertStat(t, results[0].Stats[7], "Latency p999 (ms)", 0.0076, "")
		assertStat(t, results[0].Stats[8], "RME (%)", 0.09, "±")
		assertStat(t, results[0].Stats[9], "Samples", 132207, "")

		if results[0].Name != "bubbleSort" {
			t.Errorf("Name = %q, want %q", results[0].Name, "bubbleSort")
		}
		if results[0].YAxis != "n=100" {
			t.Errorf("YAxis = %q, want %q", results[0].YAxis, "n=100")
		}

		// Last row: n=2000/insertionSort
		last := results[5]
		if last.Name != "insertionSort" {
			t.Errorf("Name = %q, want %q", last.Name, "insertionSort")
		}
		if last.YAxis != "n=2000" {
			t.Errorf("YAxis = %q, want %q", last.YAxis, "n=2000")
		}
		assertStat(t, last.Stats[0], "Throughput avg (ops/s)", 1933.38, "")
		assertStat(t, last.Stats[9], "Samples", 967, "")
	})

	t.Run("Filter regex", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y/n"
		shared.FlagState.FilterRegex = "bubbleSort"

		results := ParseVitestBenchmark(writeVitestTestFile(t, testVitestTable))

		if len(results) != 3 {
			t.Fatalf("expected 3 bubbleSort results, got %d", len(results))
		}
		for _, r := range results {
			if r.Name != "bubbleSort" {
				t.Errorf("unexpected Name = %q", r.Name)
			}
		}
	})

	t.Run("Empty file", func(t *testing.T) {
		shared.FlagState.GroupPattern = "y"
		shared.FlagState.FilterRegex = ""

		results := ParseVitestBenchmark(writeVitestTestFile(t, ""))
		if len(results) != 0 {
			t.Errorf("expected 0 results for empty file, got %d", len(results))
		}
	})
}

func writeVitestTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
