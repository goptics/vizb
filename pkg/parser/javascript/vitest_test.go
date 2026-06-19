package javascript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
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

// writeJSTestFile writes content to a temp bench.txt and returns its path. Shared
// by the javascript suites.
func writeJSTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// assertStat asserts a stat's type, value, and symbol. Shared by the js suites.
func assertStat(t *testing.T, s shared.Stat, expectedType string, expectedValue float64, expectedSymbol string) {
	t.Helper()
	if s.Type != expectedType {
		t.Errorf("stat type mismatch: got %v want %v", s.Type, expectedType)
	}
	if s.Value == nil || *s.Value != expectedValue {
		t.Errorf("stat value mismatch: got %v want %v", s.Value, expectedValue)
	}
	if s.Symbol != expectedSymbol {
		t.Errorf("stat symbol mismatch: got %v want %v", s.Symbol, expectedSymbol)
	}
}

// VitestSuite exercises ParseVitestBenchmark with a per-test parser.Config.
type VitestSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *VitestSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "y/n", TimeUnit: "ms"}
}

func (s *VitestSuite) TestRealVitestOutput() {
	results := ParseVitestBenchmark(writeJSTestFile(s.T(), testVitestTable), s.cfg)
	s.Len(results, 6)

	assertStat(s.T(), results[0].Stats[0], "Throughput avg (ops/s)", 264413.96, "")
	assertStat(s.T(), results[0].Stats[1], "Latency min (ms)", 0.0037, "")
	assertStat(s.T(), results[0].Stats[2], "Latency max (ms)", 0.0974, "")
	assertStat(s.T(), results[0].Stats[3], "Latency avg (ms)", 0.0038, "")
	assertStat(s.T(), results[0].Stats[4], "Latency p75 (ms)", 0.0038, "")
	assertStat(s.T(), results[0].Stats[5], "Latency p99 (ms)", 0.0049, "")
	assertStat(s.T(), results[0].Stats[6], "Latency p995 (ms)", 0.006, "")
	assertStat(s.T(), results[0].Stats[7], "Latency p999 (ms)", 0.0076, "")
	assertStat(s.T(), results[0].Stats[8], "RME (%)", 0.09, "±")
	assertStat(s.T(), results[0].Stats[9], "Samples", 132207, "")

	s.Equal("bubbleSort", results[0].Name)
	s.Equal("n=100", results[0].YAxis)

	last := results[5]
	s.Equal("insertionSort", last.Name)
	s.Equal("n=2000", last.YAxis)
	assertStat(s.T(), last.Stats[0], "Throughput avg (ops/s)", 1933.38, "")
	assertStat(s.T(), last.Stats[9], "Samples", 967, "")
}

func (s *VitestSuite) TestUnitConversionToUs() {
	s.cfg.TimeUnit = "us"

	results := ParseVitestBenchmark(writeJSTestFile(s.T(), testVitestTable), s.cfg)
	s.Len(results, 6)

	assertStat(s.T(), results[0].Stats[3], "Latency avg (us)", 3.8, "")
	assertStat(s.T(), results[0].Stats[5], "Latency p99 (us)", 4.9, "")
}

func (s *VitestSuite) TestFilterRegex() {
	s.cfg.Filter = "bubbleSort"

	results := ParseVitestBenchmark(writeJSTestFile(s.T(), testVitestTable), s.cfg)
	s.Len(results, 3)
	for _, r := range results {
		s.Equal("bubbleSort", r.Name)
	}
}

func (s *VitestSuite) TestEmptyFile() {
	s.cfg.GroupPattern = "y"

	results := ParseVitestBenchmark(writeJSTestFile(s.T(), ""), s.cfg)
	s.Empty(results)
}

func TestVitestSuite(t *testing.T) {
	suite.Run(t, new(VitestSuite))
}
