package shared

import (
	"encoding/json"
	"testing"

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

// TestParseOverrides_BarSwap is the canonical example from the Task 4 spec:
// `ParseOverrides(["bar:swap=yxn"], []string{"bar"}, nil)` returns a map whose
// "bar" entry is a typed bar Config with Swap == "yxn".
//
// We verify the concrete type via JSON roundtrip to avoid an import cycle
// (config/charts/bar imports shared, so shared cannot import it back).
func (s *ChartSpecSuite) TestParseOverridesBarSwap() {
	got, err := ParseOverrides([]string{"bar:swap=yxn"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Require().NotNil(got)

	cfg, ok := got["bar"]
	s.Require().True(ok, "expected bar entry in overrides map")
	s.Equal("bar", cfg.ChartType())

	// Verify the typed field via JSON roundtrip. The Config's `swap` JSON
	// tag preserves the value, and bar has no other field with this name.
	raw, err := json.Marshal(cfg)
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	s.Equal("yxn", m["swap"])
}

// TestParseOverrides_AllFields exercises every supported key in a single spec
// to confirm the typed config receives all the values.
func (s *ChartSpecSuite) TestParseOverridesAllFields() {
	got, err := ParseOverrides(
		[]string{"bar:swap=yxn,sort=asc,scale=log,labels=true,3d-rotate=false"},
		[]string{"bar"},
		s.xynAxes,
	)
	s.Require().NoError(err)
	s.Require().Contains(got, "bar")

	raw, err := json.Marshal(got["bar"])
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	s.Equal("yxn", m["swap"])
	s.Equal("log", m["scale"])
	s.Equal("asc", m["sort"].(map[string]any)["order"])
	s.Equal(true, m["sort"].(map[string]any)["enabled"])
	s.Equal(true, m["showLabels"])
	s.Equal(false, m["threeDRotate"])
}

// TestParseOverrides_BareLabels confirms a bare flag (no =val) parses correctly.
func (s *ChartSpecSuite) TestParseOverridesBareLabels() {
	got, err := ParseOverrides([]string{"bar:labels"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)

	raw, err := json.Marshal(got["bar"])
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	s.Equal(true, m["showLabels"])
}

// TestParseOverrides_BareThreeD confirms `3d` (no =val) enables value-mode 3D.
func (s *ChartSpecSuite) TestParseOverridesBareThreeD() {
	got, err := ParseOverrides([]string{"bar:3d"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)

	raw, err := json.Marshal(got["bar"])
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	s.Equal(true, m["threeD"])
}

// TestParseOverrides_BareThreeDVisualMap confirms `3d-visualmap` enables the gradient.
func (s *ChartSpecSuite) TestParseOverridesBareThreeDVisualMap() {
	got, err := ParseOverrides([]string{"bar:3d-visualmap"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)

	raw, err := json.Marshal(got["bar"])
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	s.Equal(true, m["threeDVisualMap"])
}

// TestParseOverrides_BareRotate confirms `3d-rotate` (no =val) sets threeDRotate=true.
func (s *ChartSpecSuite) TestParseOverridesBareRotate() {
	got, err := ParseOverrides([]string{"bar:3d-rotate"}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)

	raw, err := json.Marshal(got["bar"])
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	s.Equal(true, m["threeDRotate"])
}

// TestParseOverrides_MultipleSpecsSameType confirms two specs for the same
// type are merged into a single entry in the map.
func (s *ChartSpecSuite) TestParseOverridesMultipleSpecsSameType() {
	got, err := ParseOverrides(
		[]string{"bar:sort=asc", "bar:scale=log"},
		[]string{"bar"},
		s.xynAxes,
	)
	s.Require().NoError(err)
	s.Require().Len(got, 1)

	raw, err := json.Marshal(got["bar"])
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	s.Equal("log", m["scale"])
	s.Equal("asc", m["sort"].(map[string]any)["order"])
}

// TestParseOverrides_Empty confirms the function returns (nil, nil) for empty input.
func (s *ChartSpecSuite) TestParseOverridesEmpty() {
	got, err := ParseOverrides(nil, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Nil(got)

	got, err = ParseOverrides([]string{}, []string{"bar"}, s.xynAxes)
	s.Require().NoError(err)
	s.Nil(got)
}

// TestParseOverrides_Errors covers the validation failures the new contract
// still surfaces. Per-chart Validate is deferred to a future task, so
// "pie:scale=log" is silently accepted — see TestParseOverrides_NoLimitedChartCheck.
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
		{"unknown key", []string{"bar:unknown=val"}},
		{"unknown bare flag", []string{"bar:explode"}},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			_, err := ParseOverrides(c.specs, []string{"bar"}, s.xynAxes)
			s.Error(err)
		})
	}
}

// TestParseOverrides_NoLimitedChartCheck documents the deferred-validation
// contract: pie/heatmap/radar accept keys they don't carry (scale, 3d-rotate) in
// the payload; the values are silently dropped by Decode. The per-chart
// Validate(axes) method (future task) will surface this as an error.
func (s *ChartSpecSuite) TestParseOverridesNoLimitedChartCheck() {
	got, err := ParseOverrides([]string{"pie:scale=log"}, []string{"pie"}, s.xynAxes)
	s.Require().NoError(err)
	s.Require().Contains(got, "pie")
}

func (s *ChartSpecSuite) TestParseOverridesInvalidSwap() {
	_, err := ParseOverrides([]string{"bar:swap=abc"}, []string{"bar"}, s.xynAxes)
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

func TestChartSpecSuite(t *testing.T) {
	suite.Run(t, new(ChartSpecSuite))
}
