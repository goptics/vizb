package javascript

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
)

// BetterSuite checks each parser tags stats with the right better-direction.
type BetterSuite struct {
	suite.Suite
}

func TestBetterSuite(t *testing.T) {
	suite.Run(t, new(BetterSuite))
}

func (s *BetterSuite) directions(parse func(f *os.File) ([]shared.DataPoint, error), path string) map[string]string {
	f, err := os.Open(path)
	s.Require().NoError(err)
	defer f.Close()
	points, err := parse(f)
	s.Require().NoError(err)
	s.Require().NotEmpty(points)
	got := map[string]string{}
	for _, st := range points[0].Stats {
		got[st.Type] = st.Better
	}
	return got
}

func (s *BetterSuite) TestDirections() {
	cases := map[string]func(f *os.File) ([]shared.DataPoint, error){
		"../../../examples/javascript/tinybench.txt": func(f *os.File) ([]shared.DataPoint, error) {
			p, _, _, err := ParseTinyBenchBenchmark(f, parser.Config{TimeUnit: "ns", GroupPattern: "y"})
			return p, err
		},
		"../../../examples/javascript/vitest.txt": func(f *os.File) ([]shared.DataPoint, error) {
			p, _, _, err := ParseVitestBenchmark(f, parser.Config{TimeUnit: "ns", GroupPattern: "y"})
			return p, err
		},
	}
	for path, parse := range cases {
		s.Run(path, func() {
			got := s.directions(parse, path)
			s.Equal("", got["Samples"])
			s.Equal(shared.BetterHigh, got["Throughput avg (ops/s)"])
			for typ, better := range got {
				if strings.HasPrefix(typ, "Latency") || strings.Contains(typ, "RME") || strings.Contains(typ, "MAD") {
					s.Equal(shared.BetterLow, better, typ)
				}
			}
		})
	}
}
