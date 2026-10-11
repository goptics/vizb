package shared_test

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// AffixSuite covers the label affix parser and its wire type.
type AffixSuite struct {
	suite.Suite
}

func (s *AffixSuite) TestPlainTextIsGlobal() {
	s.Run("percent", func() {
		got, err := shared.ParseAffix("%")
		s.Require().NoError(err)
		s.Equal(&shared.Affix{Global: "%"}, got)
	})
	s.Run("multi-byte symbol", func() {
		got, err := shared.ParseAffix("$")
		s.Require().NoError(err)
		s.Equal(&shared.Affix{Global: "$"}, got)
	})
}

func (s *AffixSuite) TestPerColumnCommaForm() {
	got, err := shared.ParseAffix("col1=%,col2=ms")
	s.Require().NoError(err)
	s.Equal("", got.Global)
	s.Equal(map[string]string{"col1": "%", "col2": "ms"}, got.Columns)
}

func (s *AffixSuite) TestPerColumnSemicolonForm() {
	got, err := shared.ParseAffix("col1=%;col2=ms")
	s.Require().NoError(err)
	s.Equal(&shared.Affix{Columns: map[string]string{"col1": "%", "col2": "ms"}}, got)
}

func (s *AffixSuite) TestMixedGlobalAndColumns() {
	got, err := shared.ParseAffix("%,col1=ms")
	s.Require().NoError(err)
	s.Equal("%", got.Global)
	s.Equal(map[string]string{"col1": "ms"}, got.Columns)
}

func (s *AffixSuite) TestWhitespaceAroundTokensIsTrimmed() {
	got, err := shared.ParseAffix(" % , col1 = ms ")
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
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			got, err := shared.ParseAffix(c.input)
			s.Require().Error(err, c.input)
			s.Nil(got, c.input)
			s.Contains(err.Error(), c.want, c.input)
		})
	}
}

func (s *AffixSuite) TestJSONRoundTrip() {
	raw, err := json.Marshal(&shared.Affix{Global: "$"})
	s.Require().NoError(err)
	s.JSONEq(`{"global":"$"}`, string(raw))

	raw, err = json.Marshal(&shared.Affix{Columns: map[string]string{"col1": "ms"}})
	s.Require().NoError(err)
	s.JSONEq(`{"columns":{"col1":"ms"}}`, string(raw))

	var got shared.Affix
	s.Require().NoError(json.Unmarshal([]byte(`{"global":"$","columns":{"col1":"ms"}}`), &got))
	s.Equal("$", got.Global)
	s.Equal(map[string]string{"col1": "ms"}, got.Columns)
}

func TestAffixSuite(t *testing.T) {
	suite.Run(t, new(AffixSuite))
}
