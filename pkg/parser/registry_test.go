package parser_test

import (
	"strings"
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	_ "github.com/goptics/vizb/pkg/parser/csv"
	_ "github.com/goptics/vizb/pkg/parser/golang"
	"github.com/stretchr/testify/suite"
)

// RegistrySuite covers parser registry helpers.
type RegistrySuite struct {
	suite.Suite
}

func (s *RegistrySuite) TestGetParserUnknown() {
	_, err := parser.GetParser("nope")
	s.Error(err)
	s.Contains(err.Error(), "unknown parser")
	s.Contains(err.Error(), "available parsers")
}

func (s *RegistrySuite) TestGetParserKnown() {
	fn, err := parser.GetParser("go")
	s.NoError(err)
	s.NotNil(fn)

	points, _, _, err := fn(strings.NewReader(""), parser.Config{})
	s.NoError(err)
	s.Empty(points)
}

func (s *RegistrySuite) TestAvailableParsersSorted() {
	keys := parser.AvailableParsers()
	s.NotEmpty(keys)
	for i := 1; i < len(keys); i++ {
		s.Less(keys[i-1], keys[i])
	}
}

func (s *RegistrySuite) TestShouldIncludeBenchmarkMatch() {
	cfg := parser.Config{Filter: "Foo"}
	include, err := parser.ShouldIncludeBenchmark("BenchmarkFoo-8", cfg)
	s.NoError(err)
	s.True(include)
	include, err = parser.ShouldIncludeBenchmark("BenchmarkBar-8", cfg)
	s.NoError(err)
	s.False(include)
}

func (s *RegistrySuite) TestShouldIncludeBenchmarkEmptyFilter() {
	include, err := parser.ShouldIncludeBenchmark("anything", parser.Config{})
	s.NoError(err)
	s.True(include)
}

func (s *RegistrySuite) TestShouldIncludeBenchmarkInvalidRegexReturnsError() {
	_, err := parser.ShouldIncludeBenchmark("bench", parser.Config{Filter: "["})
	s.ErrorContains(err, "invalid filter regex")
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}
