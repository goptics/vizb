package rust

import (
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/stretchr/testify/suite"
)

var testDivanTable = `sort_divan              fastest       │ slowest       │ median        │ mean          │ samples │ iters
├─ bubble_sort_100      4.36 µs       │ 9.68 µs       │ 4.646 µs      │ 4.733 µs      │ 100     │ 100
├─ bubble_sort_500      103.7 µs      │ 174.6 µs      │ 108 µs        │ 119.1 µs      │ 100     │ 100
├─ bubble_sort_2000     1.36 ms       │ 2.005 ms      │ 1.391 ms      │ 1.458 ms      │ 100     │ 100
├─ insertion_sort_100   1.131 µs      │ 1.423 µs      │ 1.139 µs      │ 1.149 µs      │ 100     │ 200
├─ insertion_sort_500   20.86 µs      │ 23.8 µs       │ 21.42 µs      │ 21.39 µs      │ 100     │ 100
╰─ insertion_sort_2000  289 µs        │ 299.9 µs      │ 292.8 µs      │ 292.4 µs      │ 100     │ 100
`

// DivanSuite exercises ParseDivanBenchmark with a per-test parser.Config.
type DivanSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *DivanSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "y", TimeUnit: "ns"}
}

func (s *DivanSuite) TestRealDivanOutput() {
	results, _, _, err := ParseDivanBenchmark(rustTestInput(s.T(), testDivanTable), s.cfg)
	s.Require().NoError(err)
	s.Len(results, 6)

	assertStat(s.T(), results[0].Stats[0], "Latency fastest (ns)", 4360, "")
	assertStat(s.T(), results[0].Stats[1], "Latency slowest (ns)", 9680, "±")
	assertStat(s.T(), results[0].Stats[2], "Latency median (ns)", 4646, "")
	assertStat(s.T(), results[0].Stats[3], "Latency mean (ns)", 4733, "")
	assertStat(s.T(), results[0].Stats[4], "Samples", 100, "")

	s.Equal("bubble_sort_100", results[0].YAxis)

	assertStat(s.T(), results[2].Stats[0], "Latency fastest (ns)", 1360000, "")
	assertStat(s.T(), results[2].Stats[1], "Latency slowest (ns)", 2005000, "±")

	assertStat(s.T(), results[5].Stats[3], "Latency mean (ns)", 292400, "")
	s.Equal("insertion_sort_2000", results[5].YAxis)
}

func (s *DivanSuite) TestUnitConversionToUs() {
	s.cfg.TimeUnit = "us"

	results, _, _, err := ParseDivanBenchmark(rustTestInput(s.T(), testDivanTable), s.cfg)
	s.Require().NoError(err)
	s.Len(results, 6)

	assertStat(s.T(), results[0].Stats[0], "Latency fastest (us)", 4.36, "")
	assertStat(s.T(), results[0].Stats[3], "Latency mean (us)", 4.73, "")
	assertStat(s.T(), results[2].Stats[0], "Latency fastest (us)", 1360, "")
}

func (s *DivanSuite) TestFilterRegex() {
	s.cfg.Filter = "insertion"

	results, _, _, err := ParseDivanBenchmark(rustTestInput(s.T(), testDivanTable), s.cfg)
	s.Require().NoError(err)
	s.Len(results, 3)
	for _, r := range results {
		s.Contains(r.YAxis, "insertion")
	}
}

func (s *DivanSuite) TestEmptyFile() {
	results, _, _, err := ParseDivanBenchmark(rustTestInput(s.T(), ""), s.cfg)
	s.Require().NoError(err)
	s.Empty(results)
}

func (s *DivanSuite) TestReturnsErrors() {
	s.Run("invalid filter", func() {
		cfg := s.cfg
		cfg.Filter = "["
		_, _, _, err := ParseDivanBenchmark(rustTestInput(s.T(), testDivanTable), cfg)
		s.ErrorContains(err, "invalid filter regex")
	})

	s.Run("invalid benchmark group pattern", func() {
		cfg := s.cfg
		cfg.GroupPattern = "[n/y]"
		_, _, _, err := ParseDivanBenchmark(rustTestInput(s.T(), testDivanTable), cfg)
		s.ErrorContains(err, "bracket slots")
	})

	s.Run("reader failure", func() {
		_, _, _, err := ParseDivanBenchmark(rustErrorReader{}, s.cfg)
		s.ErrorContains(err, "read divan benchmark")
	})
}

func TestDivanSuite(t *testing.T) {
	suite.Run(t, new(DivanSuite))
}
