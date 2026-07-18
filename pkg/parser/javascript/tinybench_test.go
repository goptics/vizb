package javascript

import (
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/stretchr/testify/suite"
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

// TinyBenchSuite exercises ParseTinyBenchBenchmark with a per-test parser.Config.
type TinyBenchSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *TinyBenchSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "n/y", TimeUnit: "ns"}
}

func (s *TinyBenchSuite) TestRealTinybenchOutput() {
	results, _, err := ParseTinyBenchBenchmark(javascriptTestInput(s.T(), testSortingTable), s.cfg)
	s.Require().NoError(err)
	s.Len(results, 6)

	assertStat(s.T(), results[0].Stats[0], "Latency avg (ns)", 3741.5, "")
	assertStat(s.T(), results[0].Stats[1], "Latency RME (%)", 0.18, "±")
	assertStat(s.T(), results[0].Stats[2], "Latency med (ns)", 3703.0, "")
	assertStat(s.T(), results[0].Stats[3], "Latency MAD (ns)", 32.00, "±")
	assertStat(s.T(), results[0].Stats[4], "Throughput avg (ops/s)", 268758, "")
	assertStat(s.T(), results[0].Stats[5], "Throughput RME (%)", 0.02, "±")
	assertStat(s.T(), results[0].Stats[6], "Throughput med (ops/s)", 270051, "")
	assertStat(s.T(), results[0].Stats[7], "Throughput MAD (ops/s)", 2354, "±")
	assertStat(s.T(), results[0].Stats[8], "Samples", 267270, "")

	s.Equal("bubbleSort", results[0].Name)
	s.Equal("100", results[0].YAxis)

	last := results[5]
	s.Equal("insertionSort", last.Name)
	s.Equal("2000", last.YAxis)
	assertStat(s.T(), last.Stats[0], "Latency avg (ns)", 512841, "")
	assertStat(s.T(), last.Stats[8], "Samples", 1950, "")
}

func (s *TinyBenchSuite) TestUnitConversionToUs() {
	s.cfg.TimeUnit = "us"

	results, _, err := ParseTinyBenchBenchmark(javascriptTestInput(s.T(), testSortingTable), s.cfg)
	s.Require().NoError(err)
	s.Len(results, 6)

	assertStat(s.T(), results[0].Stats[0], "Latency avg (us)", 3.74, "")
	assertStat(s.T(), results[0].Stats[2], "Latency med (us)", 3.7, "")
}

func (s *TinyBenchSuite) TestFilterRegex() {
	s.cfg.Filter = "bubbleSort"

	results, _, err := ParseTinyBenchBenchmark(javascriptTestInput(s.T(), testSortingTable), s.cfg)
	s.Require().NoError(err)
	s.Len(results, 3)
	for _, r := range results {
		s.Equal("bubbleSort", r.Name)
	}
}

func (s *TinyBenchSuite) TestEmptyFile() {
	s.cfg.GroupPattern = "y"

	results, _, err := ParseTinyBenchBenchmark(javascriptTestInput(s.T(), ""), s.cfg)
	s.Require().NoError(err)
	s.Empty(results)
}

func (s *TinyBenchSuite) TestReturnsErrors() {
	s.Run("invalid filter", func() {
		cfg := s.cfg
		cfg.Filter = "["
		_, _, err := ParseTinyBenchBenchmark(javascriptTestInput(s.T(), testSortingTable), cfg)
		s.ErrorContains(err, "invalid filter regex")
	})

	s.Run("invalid benchmark group pattern", func() {
		cfg := s.cfg
		cfg.GroupPattern = "[n/y]"
		_, _, err := ParseTinyBenchBenchmark(javascriptTestInput(s.T(), testSortingTable), cfg)
		s.ErrorContains(err, "bracket slots")
	})

	s.Run("reader failure", func() {
		_, _, err := ParseTinyBenchBenchmark(javascriptErrorReader{}, s.cfg)
		s.ErrorContains(err, "read tinybench benchmark")
	})
}

func TestTinyBenchSuite(t *testing.T) {
	suite.Run(t, new(TinyBenchSuite))
}
