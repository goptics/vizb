package rust

import (
	"os"
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

func (s *BetterSuite) TestLatencyIsLowerBetter() {
	cases := map[string]func(f *os.File) ([]shared.DataPoint, error){
		"../../../examples/rust/criterion.txt": func(f *os.File) ([]shared.DataPoint, error) {
			p, _, _, err := ParseCriterionBenchmark(f, parser.Config{TimeUnit: "ns", GroupPattern: "y"})
			return p, err
		},
		"../../../examples/rust/divan.txt": func(f *os.File) ([]shared.DataPoint, error) {
			p, _, _, err := ParseDivanBenchmark(f, parser.Config{TimeUnit: "ns", GroupPattern: "y"})
			return p, err
		},
	}
	for path, parse := range cases {
		s.Run(path, func() {
			for typ, better := range s.directions(parse, path) {
				want := shared.BetterLow
				if typ == "Samples" {
					want = ""
				}
				s.Equal(want, better, typ)
			}
		})
	}
}
