package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ColsSpecSuite struct {
	suite.Suite
}

func (s *ColsSpecSuite) TestParseColsFlagBasic() {
	specs, err := ParseColsFlag("price,count")
	s.Require().NoError(err)
	s.Require().Len(specs, 2)
	s.Equal("price", specs[0].Source)
	s.Empty(specs[0].Label)
	s.Equal("count", specs[1].Source)
}

func (s *ColsSpecSuite) TestParseColsFlagRename() {
	specs, err := ParseColsFlag("price{Unit price},count{Total}")
	s.Require().NoError(err)
	s.Equal("price", specs[0].Source)
	s.Equal("Unit price", specs[0].Label)
	s.Equal("count", specs[1].Source)
	s.Equal("Total", specs[1].Label)
}

func (s *ColsSpecSuite) TestParseColsFlagQuotedSource() {
	specs, err := ParseColsFlag(`"price{USD}",count`)
	s.Require().NoError(err)
	s.Equal("price{USD}", specs[0].Source)
	s.Empty(specs[0].Label)
}

func (s *ColsSpecSuite) TestParseColsFlagDuplicateError() {
	_, err := ParseColsFlag("price,price")
	s.Error(err)
	s.Contains(err.Error(), "duplicate column 'price'")
}

func TestColsSpecSuite(t *testing.T) {
	suite.Run(t, new(ColsSpecSuite))
}