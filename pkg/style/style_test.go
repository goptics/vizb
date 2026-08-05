package style

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ThemeSuite struct {
	suite.Suite
}

func (s *ThemeSuite) TestValidateAndNormalizeTheme() {
	for _, value := range []string{"default", "Westeros", "#F00,#00ff00"} {
		s.NoError(ValidateTheme(NormalizeTheme(value)))
	}
	s.Equal("westeros", NormalizeTheme(" Westeros "))
	s.Equal("#F00,#00ff00", NormalizeTheme(" #F00, #00ff00 "))
}

func (s *ThemeSuite) TestRejectsInvalidTheme() {
	for _, value := range []string{"", "unknown", "#F00", "#F00,not-a-color"} {
		s.Error(ValidateTheme(NormalizeTheme(value)))
	}
}

func (s *ThemeSuite) TestValidateStructuredTheme() {
	valid := []string{
		"ocean:colors=#0ff,#00f",
		"ocean:colors=#0ff,#00f;visualMapColors=#0ff,#00f",
		"ocean:visualMapColors=#0ff,#00f;colors=#0ff,#00f,#0f0",
		" My Theme :colors= #abc , #def ",
	}
	for _, value := range valid {
		s.NoError(ValidateTheme(NormalizeTheme(value)), value)
	}
	invalid := []string{
		"ocean:visualMapColors=#0ff,#00f", // colors required
		"ocean:colors=#0ff",               // need ≥2 colors
		"ocean:colors=#0ff,#ggg",
		"ocean:colors=#0ff,#00f;visualMapColors=#0ff", // visual map needs 2
		"ocean:colors=#0ff,#00f;unknown=#0ff",
		":colors=#0ff,#00f", // empty name
	}
	for _, value := range invalid {
		s.Error(ValidateTheme(NormalizeTheme(value)), value)
	}
}

func (s *ThemeSuite) TestNormalizeStructuredTheme() {
	s.Equal(
		"ocean:colors=#0ff,#00f;visualMapColors=#abc,#def",
		NormalizeTheme(" ocean :colors= #0ff , #00f ;visualMapColors= #abc , #def "),
	)
	s.Equal(
		"ocean:colors=#0ff,#00f",
		NormalizeTheme("ocean:colors=#0ff,#00f"),
	)
}

func (s *ThemeSuite) TestParseThemeSpecBuiltIn() {
	theme, err := ParseThemeSpec("Westeros")
	s.Require().NoError(err)
	s.Equal("westeros", theme.Name)
	s.Equal(catalogColors("westeros"), theme.Colors)
	s.Equal(catalogVisualMap("westeros"), theme.VisualMapColors)
	s.Len(theme.VisualMapColors, 2)

	// All built-ins expand with curated colors.
	for name := range themeCatalog {
		t, err := ParseThemeSpec(name)
		s.Require().NoError(err, name)
		s.Equal(name, t.Name)
		s.NotEmpty(t.Colors)
		s.Len(t.VisualMapColors, 2)
	}
}

func (s *ThemeSuite) TestParseThemeSpecBareHex() {
	theme, err := ParseThemeSpec("#F00,#00ff00,#0000ff")
	s.Require().NoError(err)
	s.Equal("custom", theme.Name)
	s.Equal([]string{"#F00", "#00ff00", "#0000ff"}, theme.Colors)
	s.Equal([]string{"#F00", "#0000ff"}, theme.VisualMapColors)
}

func (s *ThemeSuite) TestParseThemeSpecStructured() {
	theme, err := ParseThemeSpec("ocean:colors=#0ff,#00f,#0f0")
	s.Require().NoError(err)
	s.Equal("ocean", theme.Name)
	s.Equal([]string{"#0ff", "#00f", "#0f0"}, theme.Colors)
	s.Equal([]string{"#0ff", "#0f0"}, theme.VisualMapColors)

	theme, err = ParseThemeSpec("ocean:visualMapColors=#111,#222;colors=#aaa,#bbb,#ccc")
	s.Require().NoError(err)
	s.Equal("ocean", theme.Name)
	s.Equal([]string{"#aaa", "#bbb", "#ccc"}, theme.Colors)
	s.Equal([]string{"#111", "#222"}, theme.VisualMapColors)
}

func (s *ThemeSuite) TestParseThemeSpecInvalid() {
	for _, value := range []string{"", "unknown", "#F00", "x:colors=#f00", "bad"} {
		_, err := ParseThemeSpec(value)
		s.Error(err, value)
	}
}

func (s *ThemeSuite) TestParseThemeSpecDefaultHasColors() {
	theme, err := ParseThemeSpec("default")
	s.Require().NoError(err)
	s.Equal("default", theme.Name)
	s.Equal([]string{
		"#5470C6", "#3BA272", "#FC8452", "#73C0DE", "#EE6666",
		"#FAC858", "#9A60B4", "#EA7CCC", "#91CC75", "#FF9F7F",
	}, theme.Colors)
	s.Equal([]string{"#91CC75", "#EE6666"}, theme.VisualMapColors)
}

func (s *ThemeSuite) TestResolveThemesSkipsDefault() {
	themes, err := ResolveThemes([]string{"default", "vintage", "roma"})
	s.Require().NoError(err)
	s.Len(themes, 2)
	s.Equal("vintage", themes[0].Name)
	s.Equal("roma", themes[1].Name)
	s.Equal(catalogColors("vintage"), themes[0].Colors)
}

func (s *ThemeSuite) TestResolveThemesOnlyDefaultReturnsEmpty() {
	themes, err := ResolveThemes([]string{"default", "DEFAULT"})
	s.Require().NoError(err)
	s.Empty(themes)
}

func (s *ThemeSuite) TestResolveThemesAnonymousCustomNames() {
	themes, err := ResolveThemes([]string{"#f00,#0f0", "#111,#222,#333"})
	s.Require().NoError(err)
	s.Require().Len(themes, 2)
	s.Equal("custom", themes[0].Name)
	s.Equal([]string{"#f00", "#0f0"}, themes[0].Colors)
	s.Equal([]string{"#f00", "#0f0"}, themes[0].VisualMapColors)
	s.Equal("custom-2", themes[1].Name)
	s.Equal([]string{"#111", "#222", "#333"}, themes[1].Colors)
	s.Equal([]string{"#111", "#333"}, themes[1].VisualMapColors)
}

func (s *ThemeSuite) TestResolveThemesAnonymousAvoidsNamedCustom() {
	themes, err := ResolveThemes([]string{
		"custom:colors=#aaa,#bbb",
		"#f00,#0f0",
	})
	s.Require().NoError(err)
	s.Require().Len(themes, 2)
	s.Equal("custom", themes[0].Name)
	s.Equal("custom-2", themes[1].Name)
}

func (s *ThemeSuite) TestResolveThemesStructuredNameStartingWithHash() {
	// Structured specs must not be treated as anonymous bare-hex, even when
	// the theme name begins with '#'.
	themes, err := ResolveThemes([]string{
		"#brand:colors=#aaa,#bbb",
		"#f00,#0f0",
	})
	s.Require().NoError(err)
	s.Require().Len(themes, 2)
	s.Equal("#brand", themes[0].Name)
	s.Equal([]string{"#aaa", "#bbb"}, themes[0].Colors)
	s.Equal("custom", themes[1].Name)
	s.Equal([]string{"#f00", "#0f0"}, themes[1].Colors)
}

func (s *ThemeSuite) TestResolveThemesDedupeLastWins() {
	themes, err := ResolveThemes([]string{
		"ocean:colors=#111,#222",
		"vintage",
		"ocean:colors=#aaa,#bbb,#ccc",
	})
	s.Require().NoError(err)
	s.Require().Len(themes, 2)
	// First-seen order, last content for ocean.
	s.Equal("ocean", themes[0].Name)
	s.Equal([]string{"#aaa", "#bbb", "#ccc"}, themes[0].Colors)
	s.Equal("vintage", themes[1].Name)
}

func (s *ThemeSuite) TestResolveThemesFirstSeenIsActive() {
	themes, err := ResolveThemes([]string{"vintage", "roma"})
	s.Require().NoError(err)
	s.Equal("vintage", themes[0].Name)
	s.Equal("roma", themes[1].Name)
}

func (s *ThemeSuite) TestResolveThemesMixed() {
	themes, err := ResolveThemes([]string{
		"default",
		"#f00,#0f0",
		"westeros",
		"brand:colors=#123,#456;visualMapColors=#123,#456",
	})
	s.Require().NoError(err)
	s.Require().Len(themes, 3)
	s.Equal("custom", themes[0].Name)
	s.Equal("westeros", themes[1].Name)
	s.Equal("brand", themes[2].Name)
}

func (s *ThemeSuite) TestResolveThemesInvalid() {
	_, err := ResolveThemes([]string{"vintage", "not-a-theme"})
	s.Error(err)
}

func (s *ThemeSuite) TestCatalogMatchesUIPalettes() {
	// Spot-check a few known UI catalog entries.
	s.Equal([]string{
		"#d87c7c", "#919e8b", "#d7ab82", "#6e7074", "#61a0a8",
		"#efa18d", "#787464", "#cc7e63", "#724e58", "#4b565b",
	}, catalogColors("vintage"))
	s.Equal([]string{"#919e8b", "#d87c7c"}, catalogVisualMap("vintage"))

	s.Equal([]string{
		"#8a7ca8", "#e098c7", "#8fd3e8", "#71669e", "#cc70af",
		"#7cb4cc", "#6d6189", "#ba79aa", "#72b8cc", "#574e7d",
	}, catalogColors("purple-passion"))
	s.Equal([]string{"#8fd3e8", "#cc70af"}, catalogVisualMap("purple-passion"))
}

// helpers for tests — read from catalog via ParseThemeSpec
func catalogColors(name string) []string {
	t, err := ParseThemeSpec(name)
	if err != nil {
		return nil
	}
	return t.Colors
}

func catalogVisualMap(name string) []string {
	t, err := ParseThemeSpec(name)
	if err != nil {
		return nil
	}
	return t.VisualMapColors
}

func TestThemeSuite(t *testing.T) {
	suite.Run(t, new(ThemeSuite))
}
