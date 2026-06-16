package rust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
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

// writeRustTestFile writes content to a temp bench.txt and returns its path.
func writeRustTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// assertStat asserts a stat's type, value, and symbol. Shared by the rust suites.
func assertStat(t *testing.T, s shared.Stat, expectedType string, expectedValue float64, expectedSymbol string) {
	t.Helper()
	require := func(got, want any, msg string) {
		if got != want {
			t.Errorf("%s: got %v want %v", msg, got, want)
		}
	}
	require(s.Type, expectedType, "stat type mismatch")
	require(s.Value, expectedValue, "stat value mismatch")
	require(s.Symbol, expectedSymbol, "stat symbol mismatch")
}

// CriterionSuite exercises ParseCriterionBenchmark with a per-test parser.Config.
type CriterionSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *CriterionSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "n/y", TimeUnit: "ns"}
}

func (s *CriterionSuite) TestRealCargoCriterionOutput() {
	results := ParseCriterionBenchmark(writeRustTestFile(s.T(), testCargoTable), s.cfg)
	s.Len(results, 6)

	// First: bubbleSort/n=100 → 3.0524 µs = 3052.4 ns
	assertStat(s.T(), results[0].Stats[0], "Latency avg (ns)", 3052.4, "")
	assertStat(s.T(), results[0].Stats[1], "Latency lower (ns)", 3042.4, "")
	assertStat(s.T(), results[0].Stats[2], "Latency upper (ns)", 3063.7, "±")

	s.Equal("bubbleSort", results[0].Name)
	s.Equal("n=100", results[0].YAxis)

	// Second: insertionSort/n=100 → 821.61 ns (already in ns)
	assertStat(s.T(), results[1].Stats[0], "Latency avg (ns)", 821.61, "")
	assertStat(s.T(), results[1].Stats[1], "Latency lower (ns)", 819.49, "")
	assertStat(s.T(), results[1].Stats[2], "Latency upper (ns)", 824.5, "±")

	// Fifth: bubbleSort/n=2000 → 1.3827 ms = 1382700 ns
	assertStat(s.T(), results[4].Stats[0], "Latency avg (ns)", 1382700, "")

	// Last: insertionSort/n=2000 → 300.19 µs = 300190 ns
	assertStat(s.T(), results[5].Stats[0], "Latency avg (ns)", 300190, "")
	s.Equal("insertionSort", results[5].Name)
	s.Equal("n=2000", results[5].YAxis)
}

func (s *CriterionSuite) TestUnitConversionToUs() {
	s.cfg.TimeUnit = "us"

	results := ParseCriterionBenchmark(writeRustTestFile(s.T(), testCargoTable), s.cfg)
	s.Len(results, 6)

	assertStat(s.T(), results[0].Stats[0], "Latency avg (us)", 3.05, "")
	assertStat(s.T(), results[1].Stats[0], "Latency avg (us)", 0.82, "")
	assertStat(s.T(), results[4].Stats[0], "Latency avg (us)", 1382.7, "")
}

func (s *CriterionSuite) TestFilterRegex() {
	s.cfg.Filter = "bubbleSort"

	results := ParseCriterionBenchmark(writeRustTestFile(s.T(), testCargoTable), s.cfg)
	s.Len(results, 3)
	for _, r := range results {
		s.Equal("bubbleSort", r.Name)
	}
}

func (s *CriterionSuite) TestEmptyFile() {
	s.cfg.GroupPattern = "y"

	results := ParseCriterionBenchmark(writeRustTestFile(s.T(), ""), s.cfg)
	s.Empty(results)
}

func TestCriterionSuite(t *testing.T) {
	suite.Run(t, new(CriterionSuite))
}
