package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SelectSpecSuite struct {
	suite.Suite
}

func (s *SelectSpecSuite) TestParseSelectFlagBasic() {
	specs, err := ParseSelectFlag("price,count")
	s.Require().NoError(err)
	s.Require().Len(specs, 2)
	s.Equal("price", specs[0].Source)
	s.Empty(specs[0].Label)
	s.Equal("count", specs[1].Source)
}

func (s *SelectSpecSuite) TestParseSelectFlagRename() {
	specs, err := ParseSelectFlag("price{Unit price},count{Total}")
	s.Require().NoError(err)
	s.Equal("price", specs[0].Source)
	s.Equal("Unit price", specs[0].Label)
	s.Equal("count", specs[1].Source)
	s.Equal("Total", specs[1].Label)
}

func (s *SelectSpecSuite) TestParseSelectFlagQuotedSource() {
	specs, err := ParseSelectFlag(`"price{USD}",count`)
	s.Require().NoError(err)
	s.Equal("price{USD}", specs[0].Source)
	s.Empty(specs[0].Label)
}

func (s *SelectSpecSuite) TestParseSelectFlagDuplicateError() {
	_, err := ParseSelectFlag("price,price")
	s.Error(err)
	s.Contains(err.Error(), "duplicate column 'price'")
}

func (s *SelectSpecSuite) TestParseSelectFlagEmpty() {
	specs, err := ParseSelectFlag("")
	s.Require().NoError(err)
	s.Nil(specs)

	specs, err = ParseSelectFlag("   ")
	s.Require().NoError(err)
	s.Nil(specs)
}

func (s *SelectSpecSuite) TestParseSelectFlagUnclosedBrace() {
	_, err := ParseSelectFlag("price{unclosed")
	s.Error(err)
	s.Contains(err.Error(), "in --select")
}

func (s *SelectSpecSuite) TestParseSelectFlagEmptyLabel() {
	_, err := ParseSelectFlag("price{}")
	s.Error(err)
}

func (s *SelectSpecSuite) TestParseSelectFlagEmptyColumnName() {
	_, err := ParseSelectFlag("{OnlyLabel}")
	s.Error(err)
	s.Contains(err.Error(), "empty column name")
}

func (s *SelectSpecSuite) TestParseSelectFlagUnexpectedQuote() {
	_, err := ParseSelectFlag(`price"x,count`)
	s.Error(err)
	s.Contains(err.Error(), `unexpected '"'`)
}

func (s *SelectSpecSuite) TestParseSelectFlagUnclosedQuote() {
	_, err := ParseSelectFlag(`"unclosed`)
	s.Error(err)
	s.Contains(err.Error(), `unclosed '"'`)
}

func (s *SelectSpecSuite) TestParseSelectFlagInvalidQuotedColumn() {
	_, err := ParseSelectFlag(`"bad\q"`)
	s.Error(err)
	s.Contains(err.Error(), "invalid quoted column")
}

func TestSelectSpecSuite(t *testing.T) {
	suite.Run(t, new(SelectSpecSuite))
}
