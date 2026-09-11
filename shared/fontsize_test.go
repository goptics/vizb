package shared_test

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type FontSizeSuite struct {
	suite.Suite
}

func TestFontSizeSuite(t *testing.T) {
	suite.Run(t, new(FontSizeSuite))
}

func (s *FontSizeSuite) TestMarshalAllEqualNumber() {
	fs := shared.FontSize{
		Series: shared.F64(14),
		Legend: shared.F64(14),
		Label:  shared.F64(14),
	}
	raw, err := json.Marshal(fs)
	s.Require().NoError(err)
	s.Equal("14", string(raw))
}

func (s *FontSizeSuite) TestMarshalPartialObject() {
	fs := shared.FontSize{Series: shared.F64(16), Legend: shared.F64(10)}
	raw, err := json.Marshal(fs)
	s.Require().NoError(err)
	s.JSONEq(`{"series":16,"legend":10}`, string(raw))
}

func (s *FontSizeSuite) TestMarshalFloat() {
	fs := shared.FontSize{
		Series: shared.F64(12.5),
		Legend: shared.F64(12.5),
		Label:  shared.F64(12.5),
	}
	raw, err := json.Marshal(fs)
	s.Require().NoError(err)
	s.Equal("12.5", string(raw))
}

func (s *FontSizeSuite) TestUnmarshalNumber() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte("14"), &fs))
	s.Equal(14.0, *fs.Series)
	s.Equal(14.0, *fs.Legend)
	s.Equal(14.0, *fs.Label)
}

func (s *FontSizeSuite) TestUnmarshalObject() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte(`{"series":16,"legend":10}`), &fs))
	s.Equal(16.0, *fs.Series)
	s.Equal(10.0, *fs.Legend)
	s.Nil(fs.Label)
}

func (s *FontSizeSuite) TestUnmarshalDropsInvalid() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte(`{"series":0,"legend":-1,"label":16,"title":14}`), &fs))
	s.Nil(fs.Series)
	s.Nil(fs.Legend)
	s.Equal(16.0, *fs.Label)
}

func (s *FontSizeSuite) TestUnmarshalInvalidNumberDropped() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte(`"14px"`), &fs))
	s.True(fs.Empty())
}

func (s *FontSizeSuite) TestParseBare() {
	fs, warns := shared.ParseFontSizeFlag("14")
	s.Empty(warns)
	s.Require().NotNil(fs)
	s.Equal(14.0, *fs.Series)
	s.Equal(14.0, *fs.Legend)
	s.Equal(14.0, *fs.Label)
}

func (s *FontSizeSuite) TestParseBareFloat() {
	fs, warns := shared.ParseFontSizeFlag("12.5")
	s.Empty(warns)
	s.Require().NotNil(fs)
	s.Equal(12.5, *fs.Series)
}

func (s *FontSizeSuite) TestParseBag() {
	fs, warns := shared.ParseFontSizeFlag("series=16;legend=10")
	s.Empty(warns)
	s.Require().NotNil(fs)
	s.Equal(16.0, *fs.Series)
	s.Equal(10.0, *fs.Legend)
	s.Nil(fs.Label)
}

func (s *FontSizeSuite) TestParseMixedValidInvalid() {
	fs, warns := shared.ParseFontSizeFlag("series=abc;legend=14;title=9")
	s.Len(warns, 2)
	s.Require().NotNil(fs)
	s.Nil(fs.Series)
	s.Equal(14.0, *fs.Legend)
}

func (s *FontSizeSuite) TestParseInvalidBare() {
	for _, raw := range []string{"0", "-1", "14px", "NaN", "Inf"} {
		fs, warns := shared.ParseFontSizeFlag(raw)
		s.Nil(fs, raw)
		s.NotEmpty(warns, raw)
	}
	fs, warns := shared.ParseFontSizeFlag("inf")
	s.Nil(fs)
	s.NotEmpty(warns)
}

func (s *FontSizeSuite) TestParseEmptyBag() {
	fs, warns := shared.ParseFontSizeFlag(";;")
	s.Nil(fs)
	s.NotEmpty(warns)
}

func (s *FontSizeSuite) TestParseCaseInsensitiveKeys() {
	fs, warns := shared.ParseFontSizeFlag("Series=16")
	s.Empty(warns)
	s.Require().NotNil(fs)
	s.Equal(16.0, *fs.Series)
}
