package parser_test

import (
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	_ "github.com/goptics/vizb/pkg/parser/golang"
	"github.com/goptics/vizb/shared"
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
	s.True(parser.ShouldIncludeBenchmark("BenchmarkFoo-8", cfg))
	s.False(parser.ShouldIncludeBenchmark("BenchmarkBar-8", cfg))
}

func (s *RegistrySuite) TestShouldIncludeBenchmarkEmptyFilter() {
	s.True(parser.ShouldIncludeBenchmark("anything", parser.Config{}))
}

func (s *RegistrySuite) TestShouldIncludeBenchmarkInvalidRegexExits() {
	restore, exitCalled := shared.TrapOsExitPanic(s.T())
	defer restore()

	s.Panics(func() {
		parser.ShouldIncludeBenchmark("bench", parser.Config{Filter: "["})
	})
	s.True(*exitCalled)
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}
