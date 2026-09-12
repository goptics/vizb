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

func (s *FontSizeSuite) TestParseStrictNumber() {
	fs, err := shared.ParseFontSizeJSONStrict([]byte("14"))
	s.Require().NoError(err)
	s.Require().NotNil(fs)
	s.Equal(14.0, *fs.Series)
	s.Equal(14.0, *fs.Legend)
	s.Equal(14.0, *fs.Label)
}

func (s *FontSizeSuite) TestParseStrictObject() {
	fs, err := shared.ParseFontSizeJSONStrict([]byte(`{"series":16,"legend":10}`))
	s.Require().NoError(err)
	s.Require().NotNil(fs)
	s.Equal(16.0, *fs.Series)
	s.Equal(10.0, *fs.Legend)
	s.Nil(fs.Label)
}

func (s *FontSizeSuite) TestParseStrictEmptyOrNull() {
	for _, raw := range []string{"", "  ", "null"} {
		fs, err := shared.ParseFontSizeJSONStrict([]byte(raw))
		s.Require().NoError(err, raw)
		s.Nil(fs, raw)
	}
}

func (s *FontSizeSuite) TestParseStrictErrors() {
	for _, test := range []struct {
		raw string
		err string
	}{
		{raw: "abc", err: "must be a number or object"},
		{raw: `"14px"`, err: "must be a number or object"},
		{raw: "{", err: "must be a number or object"},
		{raw: "[]", err: "must be a number or object"},
		{raw: "0", err: "must be a finite number greater than 0"},
		{raw: "-1", err: "must be a finite number greater than 0"},
		{raw: `{"series":"x"}`, err: "series must be a finite number greater than 0"},
		{raw: `{"series":0}`, err: "series must be a finite number greater than 0"},
		{raw: `{"bogus":1}`, err: `unknown key "bogus"`},
	} {
		fs, err := shared.ParseFontSizeJSONStrict([]byte(test.raw))
		s.Require().Error(err, test.raw)
		s.Nil(fs, test.raw)
		s.EqualError(err, test.err, test.raw)
	}
}

func (s *FontSizeSuite) TestParseStrictDuplicateNormalizedKey() {
	fs, err := shared.ParseFontSizeJSONStrict([]byte(`{"series":16," Series ":10}`))
	s.Require().Error(err)
	s.Nil(fs)
	s.EqualError(err, `duplicate key "series"`)
}

func (s *FontSizeSuite) TestUnmarshalDuplicateNormalizedKeySkipped() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte(`{"series":16," Series ":10}`), &fs))
	s.True(fs.Empty())
}

func (s *FontSizeSuite) TestUnmarshalNull() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte("null"), &fs))
	s.True(fs.Empty())
}

func (s *FontSizeSuite) TestUnmarshalZeroDropped() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte("0"), &fs))
	s.True(fs.Empty())
}

func (s *FontSizeSuite) TestUnmarshalUnknownOnlyObject() {
	var fs shared.FontSize
	s.Require().NoError(json.Unmarshal([]byte(`{"bogus":1}`), &fs))
	s.True(fs.Empty())
}

func (s *FontSizeSuite) TestMarshalUnequalObject() {
	raw, err := json.Marshal(shared.FontSize{
		Series: shared.F64(16),
		Legend: shared.F64(10),
		Label:  shared.F64(12),
	})
	s.Require().NoError(err)
	s.JSONEq(`{"series":16,"legend":10,"label":12}`, string(raw))
}

func (s *FontSizeSuite) TestEmptyNilReceiver() {
	var fs *shared.FontSize
	s.True(fs.Empty())
}

func (s *FontSizeSuite) TestParseBagEmptySegment() {
	fs, warns := shared.ParseFontSizeFlag("series=14;;legend=10")
	s.Empty(warns)
	s.Require().NotNil(fs)
	s.Equal(14.0, *fs.Series)
	s.Equal(10.0, *fs.Legend)
}

func (s *FontSizeSuite) TestParseBagMissingEquals() {
	fs, warns := shared.ParseFontSizeFlag("series=14;oops")
	s.Len(warns, 1)
	s.Require().NotNil(fs)
	s.Equal(14.0, *fs.Series)
}

func (s *FontSizeSuite) TestParseBagAllInvalid() {
	fs, warns := shared.ParseFontSizeFlag("series=abc")
	s.Nil(fs)
	s.Len(warns, 1)
}

func (s *FontSizeSuite) TestParseBareNegativeInf() {
	fs, warns := shared.ParseFontSizeFlag("-Inf")
	s.Nil(fs)
	s.NotEmpty(warns)
}
