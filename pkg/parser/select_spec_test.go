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

func TestSelectSpecSuite(t *testing.T) {
	suite.Run(t, new(SelectSpecSuite))
}