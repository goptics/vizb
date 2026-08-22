package charts_test

import (
	"testing"

	"github.com/goptics/vizb/internal/charts"
	"github.com/goptics/vizb/internal/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ChartFlagSuite struct {
	suite.Suite
}

func (s *ChartFlagSuite) TestValidateScaleValue() {
	t := s.T()
	require.NoError(t, charts.ValidateScaleValue("linear"))
	require.NoError(t, charts.ValidateScaleValue("LOG"))
	assert.Error(t, charts.ValidateScaleValue("sqrt"))
}

func (s *ChartFlagSuite) TestValidateSymbolSizeValue() {
	t := s.T()
	require.NoError(t, charts.ValidateSymbolSizeValue("12"))
	assert.Error(t, charts.ValidateSymbolSizeValue("nope"))
	assert.Error(t, charts.ValidateSymbolSizeValue("0"))
}

func (s *ChartFlagSuite) TestValidateBorderRadiusValue() {
	t := s.T()
	require.NoError(t, charts.ValidateBorderRadiusValue("0"))
	require.NoError(t, charts.ValidateBorderRadiusValue("8"))
	require.NoError(t, charts.ValidateBorderRadiusValue("100"))
	require.NoError(t, charts.ValidateBorderRadiusValue("8,8,0,0"))
	require.NoError(t, charts.ValidateBorderRadiusValue("1,2"))
	require.NoError(t, charts.ValidateBorderRadiusValue("1,2,3"))
	require.NoError(t, charts.ValidateBorderRadiusValue("1, 2, 3, 4"))
	assert.Error(t, charts.ValidateBorderRadiusValue("-5"))
	assert.Error(t, charts.ValidateBorderRadiusValue("abc"))
	assert.Error(t, charts.ValidateBorderRadiusValue("8.5"))
	assert.Error(t, charts.ValidateBorderRadiusValue(""))
	assert.Error(t, charts.ValidateBorderRadiusValue("8,8,0,0,1"))
	assert.Error(t, charts.ValidateBorderRadiusValue("8,-1,0,0"))
	assert.Error(t, charts.ValidateBorderRadiusValue("8,8.5,0,0"))
	assert.Error(t, charts.ValidateBorderRadiusValue(","))
}

func (s *ChartFlagSuite) TestEncodeBorderRadius() {
	t := s.T()
	assert.Equal(t, []int{8}, charts.EncodeBorderRadius("8"))
	assert.Equal(t, []int{0}, charts.EncodeBorderRadius("0"))
	assert.Equal(t, []int{8, 8, 0, 0}, charts.EncodeBorderRadius("8,8,0,0"))
	assert.Equal(t, []int{1, 2}, charts.EncodeBorderRadius("1,2"))
	assert.Equal(t, []int{1, 2, 3, 4}, charts.EncodeBorderRadius("1, 2, 3, 4"))
	// Non-string input is returned unchanged.
	assert.Equal(t, 42, charts.EncodeBorderRadius(42))
}

func (s *ChartFlagSuite) TestBgFlagDescriptor() {
	t := s.T()
	assert.Equal(t, "bg", charts.BgFlag.Name)
	assert.Equal(t, "bg", charts.BgFlag.EffectiveKey())
	assert.Equal(t, "background", charts.BgFlag.JSONKey)
	assert.Equal(t, flags.KindObject, charts.BgFlag.Kind)
	assert.False(t, charts.BgFlag.MultiValue, "object flags are never MultiValue")
	assert.NotNil(t, charts.BgFlag.Encode, "bg injects active:true via Encode")
	assert.Len(t, charts.BgFlag.Rule, 1, "bg is 2D-only")
	assert.Len(t, charts.BgFlag.ObjectFields, 10)

	known := map[string]bool{}
	for _, field := range charts.BgFlag.ObjectFields {
		known[field.Name] = true
	}
	for _, name := range []string{
		"color", "borderColor", "borderWidth", "borderType", "borderRadius",
		"shadowBlur", "shadowColor", "shadowOffsetX", "shadowOffsetY", "opacity",
	} {
		assert.True(t, known[name], "field %s", name)
	}
	assert.False(t, known["active"], "active is not a user field")
}

func (s *ChartFlagSuite) TestEncodeBgObject() {
	t := s.T()
	assert.Equal(t, map[string]any{"active": true}, charts.EncodeBgObject(map[string]any{}))
	assert.Equal(t,
		map[string]any{"active": true, "opacity": 0.2},
		charts.EncodeBgObject(map[string]any{"opacity": 0.2}),
	)
	// Non-map input degrades to the bare on-switch payload.
	assert.Equal(t, map[string]any{"active": true}, charts.EncodeBgObject("junk"))
}

func (s *ChartFlagSuite) TestValidateNumberValue() {
	t := s.T()
	require.NoError(t, charts.ValidateNumberValue("0"))
	require.NoError(t, charts.ValidateNumberValue("-1.5"))
	require.NoError(t, charts.ValidateNumberValue("1e3"))
	assert.Error(t, charts.ValidateNumberValue("nope"))
	assert.Error(t, charts.ValidateNumberValue(""))
}

func (s *ChartFlagSuite) TestValidateNonNegativeNumberValue() {
	t := s.T()
	require.NoError(t, charts.ValidateNonNegativeNumberValue("0"))
	require.NoError(t, charts.ValidateNonNegativeNumberValue("1.5"))
	assert.Error(t, charts.ValidateNonNegativeNumberValue("-1"))
	assert.Error(t, charts.ValidateNonNegativeNumberValue("nope"))
}

func (s *ChartFlagSuite) TestValidateOpacityValue() {
	t := s.T()
	require.NoError(t, charts.ValidateOpacityValue("0"))
	require.NoError(t, charts.ValidateOpacityValue("1"))
	require.NoError(t, charts.ValidateOpacityValue("0.25"))
	assert.Error(t, charts.ValidateOpacityValue("-0.1"))
	assert.Error(t, charts.ValidateOpacityValue("1.5"))
	assert.Error(t, charts.ValidateOpacityValue("nope"))
}

func (s *ChartFlagSuite) TestNumericValidatorsRejectNonFinite() {
	t := s.T()
	for _, v := range []string{"NaN", "+Inf", "-Inf"} {
		assert.Error(t, charts.ValidateNumberValue(v), v)
		assert.Error(t, charts.ValidateNonNegativeNumberValue(v), v)
		assert.Error(t, charts.ValidateOpacityValue(v), v)
	}
}

func (s *ChartFlagSuite) TestValidateBorderTypeValue() {
	t := s.T()
	require.NoError(t, charts.ValidateBorderTypeValue("solid"))
	require.NoError(t, charts.ValidateBorderTypeValue("dashed"))
	require.NoError(t, charts.ValidateBorderTypeValue("dotted"))
	assert.Error(t, charts.ValidateBorderTypeValue("dash"))
	assert.Error(t, charts.ValidateBorderTypeValue(""))
}

func (s *ChartFlagSuite) TestEncodeNumber() {
	t := s.T()
	assert.Equal(t, 0.5, charts.EncodeNumber("0.5"))
	assert.Equal(t, float64(-3), charts.EncodeNumber("-3"))
	// Unparseable / non-string input passes through unchanged.
	assert.Equal(t, "junk", charts.EncodeNumber("junk"))
	assert.Equal(t, 7, charts.EncodeNumber(7))
}

func TestChartFlagSuite(t *testing.T) {
	suite.Run(t, new(ChartFlagSuite))
}
