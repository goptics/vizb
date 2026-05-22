package javascript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
)

var testSortingTable = `┌─────────┬─────────────────────┬─────────────────────┬──────────────────┬────────────────────────┬────────────────────────┬─────────┐
│ (index) │ Task name           │ Latency avg (ns)    │ Latency med (ns) │ Throughput avg (ops/s) │ Throughput med (ops/s) │ Samples │
├─────────┼─────────────────────┼─────────────────────┼──────────────────┼────────────────────────┼────────────────────────┼─────────┤
│ 0       │ 'bubbleSort/100'    │ '3741.5 ± 0.18%'   │ '3703.0 ± 32.00' │ '268758 ± 0.02%'       │ '270051 ± 2354'        │ 267270  │
│ 1       │ 'insertionSort/100' │ '1197.1 ± 0.11%'   │ '1170.0 ± 4.00'  │ '842678 ± 0.01%'       │ '854701 ± 2912'        │ 835384  │
└─────────┴─────────────────────┴─────────────────────┴──────────────────┴────────────────────────┴────────────────────────┴─────────┘
┌─────────┬─────────────────────┬──────────────────┬───────────────────┬────────────────────────┬────────────────────────┬─────────┐
│ (index) │ Task name           │ Latency avg (ns) │ Latency med (ns)  │ Throughput avg (ops/s) │ Throughput med (ops/s) │ Samples │
├─────────┼─────────────────────┼──────────────────┼───────────────────┼────────────────────────┼────────────────────────┼─────────┤
│ 0       │ 'bubbleSort/500'    │ '128536 ± 0.04%' │ '128559 ± 1105.5' │ '7782 ± 0.03%'         │ '7779 ± 67'            │ 7780    │
│ 1       │ 'insertionSort/500' │ '32902 ± 0.09%'  │ '32676 ± 207.00'  │ '30427 ± 0.02%'        │ '30604 ± 194'          │ 30394   │
└─────────┴─────────────────────┴──────────────────┴───────────────────┴────────────────────────┴────────────────────────┴─────────┘
┌─────────┬──────────────────────┬───────────────────┬───────────────────┬────────────────────────┬────────────────────────┬─────────┐
│ (index) │ Task name            │ Latency avg (ns)  │ Latency med (ns)  │ Throughput avg (ops/s) │ Throughput med (ops/s) │ Samples │
├─────────┼──────────────────────┼───────────────────┼───────────────────┼────────────────────────┼────────────────────────┼─────────┤
│ 0       │ 'bubbleSort/2000'    │ '1731347 ± 0.11%' │ '1724439 ± 13108' │ '578 ± 0.10%'          │ '580 ± 4'              │ 578     │
│ 1       │ 'insertionSort/2000' │ '512841 ± 0.05%'  │ '510806 ± 1017.5' │ '1950 ± 0.04%'         │ '1958 ± 4'             │ 1950    │
└─────────┴──────────────────────┴───────────────────┴───────────────────┴────────────────────────┴────────────────────────┴─────────┘
`

func writeTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseTinyBenchBenchmark(t *testing.T) {
	origPattern := shared.FlagState.GroupPattern
	origFilter := shared.FlagState.FilterRegex
	origTimeUnit := shared.FlagState.TimeUnit
	defer func() {
		shared.FlagState.GroupPattern = origPattern
		shared.FlagState.FilterRegex = origFilter
		shared.FlagState.TimeUnit = origTimeUnit
	}()

	t.Run("Real tinybench output from sorting.txt", func(t *testing.T) {
		shared.FlagState.GroupPattern = "n/y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "ns"

		results := ParseTinyBenchBenchmark(writeTestFile(t, testSortingTable))

		if len(results) == 0 {
			t.Fatal("expected results, got none")
		}

		if len(results) != 6 {
			t.Fatalf("expected 6 results, got %d", len(results))
		}

		assertStat(t, results[0].Stats[0], "Latency avg (ns)", 3741.5, "")
		assertStat(t, results[0].Stats[1], "Latency RME (%)", 0.18, "±")
		assertStat(t, results[0].Stats[2], "Latency med (ns)", 3703.0, "")
		assertStat(t, results[0].Stats[3], "Latency MAD (ns)", 32.00, "±")
		assertStat(t, results[0].Stats[4], "Throughput avg (ops/s)", 268758, "")
		assertStat(t, results[0].Stats[5], "Throughput RME (%)", 0.02, "±")
		assertStat(t, results[0].Stats[6], "Throughput med (ops/s)", 270051, "")
		assertStat(t, results[0].Stats[7], "Throughput MAD (ops/s)", 2354, "±")
		assertStat(t, results[0].Stats[8], "Samples", 267270, "")

		if results[0].Name != "bubbleSort" {
			t.Errorf("Name = %q, want %q", results[0].Name, "bubbleSort")
		}
		if results[0].YAxis != "100" {
			t.Errorf("YAxis = %q, want %q", results[0].YAxis, "100")
		}

		last := results[5]
		if last.Name != "insertionSort" {
			t.Errorf("Name = %q, want %q", last.Name, "insertionSort")
		}
		if last.YAxis != "2000" {
			t.Errorf("YAxis = %q, want %q", last.YAxis, "2000")
		}
		assertStat(t, last.Stats[0], "Latency avg (ns)", 512841, "")
		assertStat(t, last.Stats[8], "Samples", 1950, "")
	})

	t.Run("Unit conversion to us", func(t *testing.T) {
		shared.FlagState.GroupPattern = "n/y"
		shared.FlagState.FilterRegex = ""
		shared.FlagState.TimeUnit = "us"

		results := ParseTinyBenchBenchmark(writeTestFile(t, testSortingTable))

		if len(results) != 6 {
			t.Fatalf("expected 6 results, got %d", len(results))
		}

		// 3741.5 ns → 3.74 us
		assertStat(t, results[0].Stats[0], "Latency avg (us)", 3.74, "")
		assertStat(t, results[0].Stats[2], "Latency med (us)", 3.7, "")
	})

	t.Run("Filter regex", func(t *testing.T) {
		shared.FlagState.GroupPattern = "n/y"
		shared.FlagState.FilterRegex = "bubbleSort"
		shared.FlagState.TimeUnit = "ns"

		results := ParseTinyBenchBenchmark(writeTestFile(t, testSortingTable))

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
		shared.FlagState.TimeUnit = "ns"

		results := ParseTinyBenchBenchmark(writeTestFile(t, ""))
		if len(results) != 0 {
			t.Errorf("expected 0 results for empty file, got %d", len(results))
		}
	})
}

func assertStat(t *testing.T, s shared.Stat, expectedType string, expectedValue float64, expectedSymbol string) {
	t.Helper()
	if s.Type != expectedType {
		t.Errorf("Stat type = %q, want %q", s.Type, expectedType)
	}
	if s.Value != expectedValue {
		t.Errorf("Stat value = %f, want %f", s.Value, expectedValue)
	}
	if s.Symbol != expectedSymbol {
		t.Errorf("Stat symbol = %q, want %q", s.Symbol, expectedSymbol)
	}
}
