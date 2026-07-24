package shared

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/goptics/vizb/internal/flags"
	"github.com/stretchr/testify/suite"
)

// ChartSpecSuite covers --chart spec parsing and the shared swap validator.
type ChartSpecSuite struct {
	suite.Suite
	xynAxes   []Axis
	allCharts []string
}

func (s *ChartSpecSuite) SetupTest() {
	s.xynAxes = []Axis{{Key: "x"}, {Key: "y"}, {Key: "name"}}
	s.allCharts = []string{"bar", "line", "pie", "heatmap", "radar"}
}

// payload marshals a parsed override config back to a generic map so its typed
// fields can be asserted without an import cycle (config/charts/<chart> imports
// shared, so shared cannot import them back).
func (s *ChartSpecSuite) payload(cfg any) map[string]any {
	raw, err := json.Marshal(cfg)
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	return m
}

// TestParseOverrides_BarSwap is the canonical example: ParseOverrides for
// "bar:swap=yxn" returns a typed bar Config with Swap == "yxn".
func (s *ChartSpecSuite) TestParseOverridesBarSwap() {
	got, warnings, err := ParseOverrides([]string{"bar:swap=yxn"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Empty(warnings)
	s.Require().NotNil(got)

	cfg, ok := got["bar"]
	s.Require().True(ok, "expected bar entry in overrides map")
	s.Equal("bar", cfg.ChartType())
	s.Equal("yxn", s.payload(cfg)["swap"])
}

// TestParseOverrides_AllFields exercises every supported key in a single spec.
func (s *ChartSpecSuite) TestParseOverridesAllFields() {
	got, _, err := ParseOverrides(
		[]string{"bar:swap=yxn,sort=asc,scale=log,labels=true,3d-rotate=false"},
		[]string{"bar"},
		s.xynAxes,
	)
	s.Require().NoError(err)
	s.Require().Contains(got, "bar")

	m := s.payload(got["bar"])
	s.Equal("yxn", m["swap"])
	s.Equal("log", m["scale"])
	s.Equal("asc", m["sort"].(map[string]any)["order"])
	s.Equal(true, m["sort"].(map[string]any)["enabled"])
	s.Equal(true, m["showLabels"])
	s.Equal(false, m["threeDRotate"])
}

// TestParseOverrides_BareLabels confirms a bare flag (no =val) parses correctly.
func (s *ChartSpecSuite) TestParseOverridesBareLabels() {
	got, _, err := ParseOverrides([]string{"bar:labels"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Equal(true, s.payload(got["bar"])["showLabels"])
}

func (s *ChartSpecSuite) TestParseOverridesLabelMode() {
	got, _, err := ParseOverrides([]string{"bar:label-mode=percentage"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Equal("percentage", s.payload(got["bar"])["labelMode"])

	_, _, err = ParseOverrides([]string{"bar:label-mode=percent"}, []string{"bar"}, s.xynAxes)
	s.ErrorContains(err, "label mode")
}

func (s *ChartSpecSuite) TestParseOverridesLineSmooth() {
	got, warnings, err := ParseOverrides([]string{"line:smooth"}, []string{"line"}, s.xynAxes)
	s.Require().NoError(err)
	s.Empty(warnings)
	s.Equal(true, s.payload(got["line"])["smooth"])

	got, warnings, err = ParseOverrides([]string{"bar:smooth"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Require().NotEmpty(warnings)
	s.Contains(warnings[0], "smooth")
	s.Nil(s.payload(got["bar"])["smooth"])
}

func (s *ChartSpecSuite) TestParseOverridesBarHorizontal() {
	got, warnings, err := ParseOverrides([]string{"bar:horizontal"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Empty(warnings)
	s.Equal(true, s.payload(got["bar"])["horizontal"])
}

// TestParseOverrides_BareStack confirms a bare stack flag parses correctly.
func (s *ChartSpecSuite) TestParseOverridesBareStack() {
	got, _, err := ParseOverrides([]string{"bar:stack"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Equal(true, s.payload(got["bar"])["stack"])
}

// TestParseOverrides_BareThreeD confirms `3d` (no =val) enables value-mode 3D.
func (s *ChartSpecSuite) TestParseOverridesBareThreeD() {
	got, _, err := ParseOverrides([]string{"bar:3d"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Equal(true, s.payload(got["bar"])["threeD"])
}

// TestParseOverrides_ScatterVisualMap confirms `visualmap` applies to scatter and
// is dropped-with-warning (not an error) for charts that don't carry it.
func (s *ChartSpecSuite) TestParseOverridesScatterVisualMap() {
	got, _, err := ParseOverrides([]string{"scatter:visualmap"}, []string{"scatter"}, s.xynAxes)
	s.Require().NoError(err)
	s.Equal(true, s.payload(got["scatter"])["visualMap"])

	got, warnings, err := ParseOverrides([]string{"bar:visualmap"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Require().NotEmpty(warnings)
	s.Contains(warnings[0], "visualmap")
	s.Nil(s.payload(got["bar"])["visualMap"]) // dropped, not applied
}

// TestParseOverrides_BareThreeDVisualMap confirms `3d-visualmap` enables the gradient.
func (s *ChartSpecSuite) TestParseOverridesBareThreeDVisualMap() {
	got, _, err := ParseOverrides([]string{"bar:3d-visualmap"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Equal(true, s.payload(got["bar"])["threeDVisualMap"])
}

// TestParseOverrides_SymbolFields confirms symbol + symbol-size parse into typed config.
func (s *ChartSpecSuite) TestParseOverridesSymbolFields() {
	got, _, err := ParseOverrides(
		[]string{"scatter:symbol=triangle,symbol-size=12"},
		[]string{"scatter"},
		s.xynAxes,
	)
	s.Require().NoError(err)
	m := s.payload(got["scatter"])
	s.Equal("triangle", m["symbol"])
	s.Equal(12.0, m["symbolSize"])
}

// TestParseOverrides_BareRotate confirms `3d-rotate` (no =val) sets threeDRotate=true.
func (s *ChartSpecSuite) TestParseOverridesBareRotate() {
	got, _, err := ParseOverrides([]string{"bar:3d-rotate"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Equal(true, s.payload(got["bar"])["threeDRotate"])
}

// TestParseOverrides_MultipleSpecsSameType confirms two specs for the same
// type are merged into a single entry in the map.
func (s *ChartSpecSuite) TestParseOverridesMultipleSpecsSameType() {
	got, _, err := ParseOverrides(
		[]string{"bar:sort=asc", "bar:scale=log"},
		[]string{"bar"},
		s.xynAxes,
	)
	s.Require().NoError(err)
	s.Require().Len(got, 1)

	m := s.payload(got["bar"])
	s.Equal("log", m["scale"])
	s.Equal("asc", m["sort"].(map[string]any)["order"])
}

// TestParseOverrides_Empty confirms the function returns nil for empty input.
func (s *ChartSpecSuite) TestParseOverridesEmpty() {
	got, warnings, err := ParseOverrides(nil, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Nil(got)
	s.Nil(warnings)

	got, _, err = ParseOverrides([]string{}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Nil(got)
}

// TestParseOverrides_Errors covers the validation failures that are hard errors:
// malformed specs, unknown chart types, inactive charts, invalid values, and
// keys that are valid for no chart (typos). Keys valid for another chart are
// dropped-with-warning instead — see TestParseOverrides_CrossChartKeyDropped.
func (s *ChartSpecSuite) TestParseOverridesErrors() {
	cases := []struct {
		name  string
		specs []string
	}{
		{"malformed (no colon)", []string{"barswap=yxn"}},
		{"empty rest after colon", []string{"bar:"}},
		{"unknown chart type", []string{"graph:swap=yxn"}},
		{"chart not in --charts", []string{"pie:sort=asc"}}, // only "bar" is active
		{"bad sort value", []string{"bar:sort=invalid"}},
		{"swap not a permutation", []string{"bar:swap=abc"}},
		{"unknown key (typo)", []string{"bar:unknown=val"}},
		{"unknown bare flag (typo)", []string{"bar:explode"}},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			_, _, err := ParseOverrides(c.specs, []string{"bar"}, s.xynAxes)
			s.Error(err)
		})
	}
}

// TestParseOverrides_CrossChartKeyDropped documents the drop-with-warning
// contract: a key valid for another chart (e.g. pie:scale) is dropped with a
// warning rather than erroring or being silently ignored.
func (s *ChartSpecSuite) TestParseOverridesCrossChartKeyDropped() {
	got, warnings, err := ParseOverrides([]string{"pie:scale=log"}, []string{"pie"}, s.xynAxes)
	s.Require().NoError(err)
	s.Require().Contains(got, "pie")
	s.Require().NotEmpty(warnings)
	s.True(strings.Contains(warnings[0], "scale") && strings.Contains(warnings[0], "pie"))
}

// TestParseOverrides_BarStatBare confirms bare `stat` (no =value) enables all
// stat categories.
func (s *ChartSpecSuite) TestParseOverridesBarStatBare() {
	got, _, err := ParseOverrides([]string{"bar:stat"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	m := s.payload(got["bar"])
	stat, ok := m["stat"].(map[string]any)
	s.Require().True(ok, "expected stat to be a map")
	s.Equal(true, stat["enabled"])
	math, _ := stat["math"].([]any)
	s.Empty(math, "bare stat should produce empty math (all categories)")
}

// TestParseOverrides_BarStatValue confirms `stat=<category>` enables stats for
// the specified category.
func (s *ChartSpecSuite) TestParseOverridesBarStatValue() {
	got, _, err := ParseOverrides([]string{"bar:stat=center"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	m := s.payload(got["bar"])
	stat, ok := m["stat"].(map[string]any)
	s.Require().True(ok, "expected stat to be a map")
	s.Equal(true, stat["enabled"])
	math, _ := stat["math"].([]any)
	s.Require().Len(math, 1)
	s.Equal("center", math[0])
}

// TestParseOverrides_BarStatInvalid confirms `stat=<invalid>` is a hard error.
func (s *ChartSpecSuite) TestParseOverridesBarStatInvalid() {
	_, _, err := ParseOverrides([]string{"bar:stat=bogus"}, []string{"bar"}, s.xynAxes)
	s.Error(err)
}

func (s *ChartSpecSuite) TestParseOverridesInvalidSwap() {
	_, _, err := ParseOverrides([]string{"bar:swap=abc"}, []string{"bar"}, s.xynAxes)
	s.Error(err)
}

func (s *ChartSpecSuite) TestValidateSwap() {
	s.Run("empty swap is always valid", func() {
		s.NoError(ValidateSwap("", s.xynAxes))
	})
	s.Run("valid permutation passes", func() {
		s.NoError(ValidateSwap("yxn", s.xynAxes))
	})
	s.Run("non-permutation fails", func() {
		s.Error(ValidateSwap("abc", s.xynAxes))
	})
	s.Run("unknown axis char fails", func() {
		s.Error(ValidateSwap("xyz", s.xynAxes))
	})
	s.Run("no axes accepts any non-empty swap", func() {
		s.NoError(ValidateSwap("anything", nil))
	})
	s.Run("metric axis excluded from identity", func() {
		axes := []Axis{{Key: "x"}, {Key: "y"}, {Key: "metric"}, {Key: "name"}}
		s.NoError(ValidateSwap("yxn", axes))
		s.Error(ValidateSwap("xym", axes))
	})
}

func (s *ChartSpecSuite) TestConvertIntegerFlagValue() {
	flag := flags.Flag{Name: "limit", Kind: flags.KindInt}

	value, err := convertFlagValue(flag, "12", true, nil)
	s.Require().NoError(err)
	s.Equal(12, value)

	_, err = convertFlagValue(flag, "twelve", true, nil)
	s.EqualError(err, `--chart: key "limit" value "twelve" must be an integer`)

	_, err = convertFlagValue(flag, "", false, nil)
	s.EqualError(err, `--chart: key "limit" requires a value (e.g. limit=<value>)`)
}

func TestChartSpecSuite(t *testing.T) {
	suite.Run(t, new(ChartSpecSuite))
}
