package flags_test

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/internal/flags"
	"github.com/stretchr/testify/suite"
)

// AffixSuite covers the label affix parser and its wire type.
type AffixSuite struct {
	suite.Suite
}

func (s *AffixSuite) TestPlainTextIsGlobal() {
	s.Run("percent", func() {
		got, err := flags.ParseAffix("%")
		s.Require().NoError(err)
		s.Equal(&flags.Affix{Global: "%"}, got)
	})
	s.Run("multi-byte symbol", func() {
		got, err := flags.ParseAffix("$")
		s.Require().NoError(err)
		s.Equal(&flags.Affix{Global: "$"}, got)
	})
}

func (s *AffixSuite) TestPerColumnCommaForm() {
	got, err := flags.ParseAffix("col1=%,col2=ms")
	s.Require().NoError(err)
	s.Equal("", got.Global)
	s.Equal(map[string]string{"col1": "%", "col2": "ms"}, got.Columns)
}

func (s *AffixSuite) TestPerColumnSemicolonForm() {
	got, err := flags.ParseAffix("col1=%;col2=ms")
	s.Require().NoError(err)
	s.Equal(&flags.Affix{Columns: map[string]string{"col1": "%", "col2": "ms"}}, got)
}

func (s *AffixSuite) TestMixedGlobalAndColumns() {
	got, err := flags.ParseAffix("%,col1=ms")
	s.Require().NoError(err)
	s.Equal("%", got.Global)
	s.Equal(map[string]string{"col1": "ms"}, got.Columns)
}

func (s *AffixSuite) TestWhitespaceAroundTokensIsTrimmed() {
	got, err := flags.ParseAffix(" % , col1 = ms ")
	s.Require().NoError(err)
	s.Equal("%", got.Global)
	s.Equal(map[string]string{"col1": "ms"}, got.Columns)
}

func (s *AffixSuite) TestErrors() {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "empty"},
		{"dup global", "%,col=ms,%", "global"},
		{"dup column", "col=%,col=ms", "col"},
		{"empty key", "=x", "key"},
		{"empty symbol", "col=", "symbol"},
		{"empty token", ",", "token"},
		{"trailing separator", "col=ms,", "token"},
		{"mixed separators", "%,a=b;c=d", "mixed"},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			got, err := flags.ParseAffix(c.input)
			s.Require().Error(err, c.input)
			s.Nil(got, c.input)
			s.Contains(err.Error(), c.want, c.input)
		})
	}
}

func (s *AffixSuite) TestJSONRoundTrip() {
	raw, err := json.Marshal(&flags.Affix{Global: "$"})
	s.Require().NoError(err)
	s.JSONEq(`{"global":"$"}`, string(raw))

	raw, err = json.Marshal(&flags.Affix{Columns: map[string]string{"col1": "ms"}})
	s.Require().NoError(err)
	s.JSONEq(`{"columns":{"col1":"ms"}}`, string(raw))

	var got flags.Affix
	s.Require().NoError(json.Unmarshal([]byte(`{"global":"$","columns":{"col1":"ms"}}`), &got))
	s.Equal("$", got.Global)
	s.Equal(map[string]string{"col1": "ms"}, got.Columns)
}

func TestAffixSuite(t *testing.T) {
	suite.Run(t, new(AffixSuite))
}
