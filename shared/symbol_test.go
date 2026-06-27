package shared

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SymbolSuite struct {
	suite.Suite
}

func (s *SymbolSuite) TestValidateSymbolBuiltins() {
	for _, sym := range []string{"circle", "rect", "roundRect", "triangle", "diamond", "pin", "arrow", "none"} {
		s.NoError(ValidateSymbol(sym), sym)
	}
	s.NoError(ValidateSymbol("CIRCLE"))
}

func (s *SymbolSuite) TestValidateSymbolPathAndImage() {
	s.NoError(ValidateSymbol("path://M0,0 L10,10"))
	s.NoError(ValidateSymbol("image://https://example.com/icon.png"))
	s.NoError(ValidateSymbol("M0,0 L10,10"))
}

func (s *SymbolSuite) TestValidateSymbolRejectsUnknown() {
	s.Error(ValidateSymbol("hexagon"))
}

func (s *SymbolSuite) TestValidateSymbolSize() {
	s.NoError(ValidateSymbolSize(8))
	s.Error(ValidateSymbolSize(0))
	s.Error(ValidateSymbolSize(-1))
}

func TestSymbolSuite(t *testing.T) {
	suite.Run(t, new(SymbolSuite))
}
