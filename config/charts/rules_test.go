package charts_test

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/config/flags"
	"github.com/stretchr/testify/suite"
)

// testRuleConfig is a minimal ChartConfig used by ApplyRules tests that need
// flags with Rule fields (no production chart has non-nil Rule in Phase B).
type testRuleConfig struct {
	Type      string `json:"type"`
	XFlag     string `json:"xFlag,omitempty"`
	YFlag     string `json:"yFlag,omitempty"`
	BothFlag  string `json:"bothFlag,omitempty"`
	XFlagR    string `json:"xFlagR,omitempty"`
	YFlagR    string `json:"yFlagR,omitempty"`
	BothFlagR string `json:"bothFlagR,omitempty"`
}

func (testRuleConfig) ChartType() string  { return "_test_rules_chart" }
func (testRuleConfig) StatEnabled() bool  { return false }
func (testRuleConfig) StatMath() []string { return nil }
func (testRuleConfig) SwapString() string { return "" }

// RulesSuite covers rule builders (RequiresAxes, RequiresZAxis, OnlyScatter2D)
// and the ApplyRules pipeline pass.
type RulesSuite struct {
	suite.Suite
}

func (s *RulesSuite) SetupSuite() {
	charts.Register(charts.Spec{
		Type:    "_test_rules_chart",
		Factory: func() charts.ChartConfig { return &testRuleConfig{} },
		Flags: []flags.Flag{
			{Name: "x-flag", JSONKey: "xFlag"},
			{Name: "y-flag", JSONKey: "yFlag"},
			{Name: "both-flag", JSONKey: "bothFlag"},
			{Name: "x-flag-r", JSONKey: "xFlagR", Rule: []flags.RuleFn{charts.RequiresAxes("x")}},
			{Name: "y-flag-r", JSONKey: "yFlagR", Rule: []flags.RuleFn{charts.RequiresAxes("y")}},
			{
				Name:    "both-flag-r",
				JSONKey: "bothFlagR",
				Rule:    []flags.RuleFn{charts.RequiresAxes("x"), charts.RequiresAxes("y")},
			},
		},
	})
}

// --- RequiresAxes builder ---

func (s *RulesSuite) TestRequiresAxes_KeepWhenAllPresent() {
	rule := charts.RequiresAxes("x", "y")
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{{Key: "x"}, {Key: "y"}},
	})
	s.Equal(flags.Keep, out)
	s.Empty(msg)
}

func (s *RulesSuite) TestRequiresAxes_SkipWhenMissingAxis() {
	rule := charts.RequiresAxes("x", "y")
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{{Key: "x"}},
	})
	s.Equal(flags.Skip, out)
	s.Contains(msg, "requires axis \"y\"")
}

func (s *RulesSuite) TestRequiresAxes_SkipWhenZMissing() {
	rule := charts.RequiresAxes("z")
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{{Key: "x"}, {Key: "y"}},
	})
	s.Equal(flags.Skip, out)
	s.Contains(msg, "requires axis \"z\"")
}

func (s *RulesSuite) TestRequiresAxes_PanicsOnEmptyKeys() {
	s.Panics(func() {
		charts.RequiresAxes()
	})
}

// --- RequiresZAxis convenience ---

func (s *RulesSuite) TestRequiresZAxis_KeepWhenZPresent() {
	rule := charts.RequiresZAxis()
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{{Key: "z"}},
	})
	s.Equal(flags.Keep, out)
	s.Empty(msg)
}

func (s *RulesSuite) TestRequiresZAxis_SkipWhenZMissing() {
	rule := charts.RequiresZAxis()
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{{Key: "x"}, {Key: "y"}},
	})
	s.Equal(flags.Skip, out)
	s.Contains(msg, "requires axis \"z\"")
}

// --- Requires3DMode ---

func (s *RulesSuite) TestRequires3DMode_KeepWhenZPresent() {
	rule := charts.Requires3DMode()
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{{Key: "z", Type: ""}},
	})
	s.Equal(flags.Keep, out)
	s.Empty(msg)
}

func (s *RulesSuite) TestRequires3DMode_KeepWhenXYZValueMode() {
	rule := charts.Requires3DMode()
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{
			{Key: "x", Type: "value"},
			{Key: "y", Type: "value"},
			{Key: "z", Type: "value"},
		},
	})
	s.Equal(flags.Keep, out)
	s.Empty(msg)
}

func (s *RulesSuite) TestRequires3DMode_SkipWhenNo3DData() {
	rule := charts.Requires3DMode()
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{
			{Key: "x"},
			{Key: "y"},
		},
	})
	s.Equal(flags.Skip, out)
	s.Contains(msg, "requires z-axis in data")
}

// --- OnlyScatter2D ---

func (s *RulesSuite) TestOnlyScatter2D_KeepWhenNoXYZValueMode() {
	rule := charts.OnlyScatter2D()
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{
			{Key: "x", Type: "value"},
			{Key: "y", Type: "value"},
		},
	})
	s.Equal(flags.Keep, out)
	s.Empty(msg)
}

func (s *RulesSuite) TestOnlyScatter2D_SkipWhenXYZAllValue() {
	rule := charts.OnlyScatter2D()
	out, msg := rule(charts.RuleContext{
		Axes: []charts.AxisInfo{
			{Key: "x", Type: "value"},
			{Key: "y", Type: "value"},
			{Key: "z", Type: "value"},
		},
	})
	s.Equal(flags.Skip, out)
	s.Contains(msg, "visualmap skipped")
}

// --- ApplyRules basic wiring ---

func (s *RulesSuite) TestApplyRules_EmptyConfigs() {
	warnings, fatal := charts.ApplyRules(charts.RuleContext{}, nil)
	s.Nil(warnings)
	s.Nil(fatal)
}

func (s *RulesSuite) TestApplyRules_NoRulesOnDescriptors() {
	// Phase B state: all production flags have nil Rule — no-op.
	raw, err := json.Marshal(barchart.Config{Type: "bar", Scale: "log"})
	s.Require().NoError(err)

	cfg, err := charts.Decode("bar", raw)
	s.Require().NoError(err)

	configs := []charts.ChartConfig{cfg}
	warnings, fatal := charts.ApplyRules(charts.RuleContext{Axes: []charts.AxisInfo{{Key: "x"}}}, configs)
	s.Nil(warnings)
	s.Nil(fatal)
}

// --- ApplyRules with synthetic flags ---

func (s *RulesSuite) TestApplyRules_KeepWhenAxesMet() {
	raw, err := json.Marshal(testRuleConfig{
		Type:   "_test_rules_chart",
		XFlagR: "set",
	})
	s.Require().NoError(err)

	cfg, err := charts.Decode("_test_rules_chart", raw)
	s.Require().NoError(err)

	ctx := charts.RuleContext{Axes: []charts.AxisInfo{{Key: "x"}}}
	configs := []charts.ChartConfig{cfg}
	warnings, fatal := charts.ApplyRules(ctx, configs)
	s.Nil(fatal)
	s.Empty(warnings)

	// JSON round-trip: xFlagR should still be present.
	out, err := json.Marshal(configs[0])
	s.Require().NoError(err)
	var result testRuleConfig
	s.Require().NoError(json.Unmarshal(out, &result))
	s.Equal("set", result.XFlagR)
}

func (s *RulesSuite) TestApplyRules_SkipWhenAxisMissing() {
	raw, err := json.Marshal(testRuleConfig{
		Type:   "_test_rules_chart",
		XFlagR: "set",
		YFlagR: "set",
	})
	s.Require().NoError(err)

	cfg, err := charts.Decode("_test_rules_chart", raw)
	s.Require().NoError(err)

	ctx := charts.RuleContext{Axes: []charts.AxisInfo{{Key: "x"}}}
	configs := []charts.ChartConfig{cfg}
	warnings, fatal := charts.ApplyRules(ctx, configs)
	s.Nil(fatal)
	s.Len(warnings, 1)
	s.Contains(warnings[0], `"y-flag-r" skipped`)

	// JSON round-trip: xFlagR kept, yFlagR dropped.
	out, err := json.Marshal(configs[0])
	s.Require().NoError(err)
	var result testRuleConfig
	s.Require().NoError(json.Unmarshal(out, &result))
	s.Equal("set", result.XFlagR)
	s.Empty(result.YFlagR)
}

func (s *RulesSuite) TestApplyRules_MultipleRulesWorstWins() {
	raw, err := json.Marshal(testRuleConfig{
		Type:      "_test_rules_chart",
		BothFlagR: "set",
	})
	s.Require().NoError(err)

	cfg, err := charts.Decode("_test_rules_chart", raw)
	s.Require().NoError(err)

	ctx := charts.RuleContext{Axes: []charts.AxisInfo{{Key: "x"}}}
	configs := []charts.ChartConfig{cfg}
	warnings, fatal := charts.ApplyRules(ctx, configs)
	s.Nil(fatal)
	s.Len(warnings, 1)
	s.Contains(warnings[0], `"both-flag-r" skipped`)
	s.Contains(warnings[0], `requires axis "y"`)

	// JSON round-trip: bothFlagR dropped.
	out, err := json.Marshal(configs[0])
	s.Require().NoError(err)
	var result testRuleConfig
	s.Require().NoError(json.Unmarshal(out, &result))
	s.Empty(result.BothFlagR)
}

func TestRulesSuite(t *testing.T) {
	suite.Run(t, new(RulesSuite))
}
