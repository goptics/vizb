package bar

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/goptics/vizb/config/charts"
	_ "github.com/goptics/vizb/config/charts/heatmap"
	_ "github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/radar"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// BarSuite covers bar config registry, decode, and Materialise precedence.
type BarSuite struct {
	suite.Suite
}

func (s *BarSuite) TestRegistryListsChartTypes() {
	got := charts.Registered()
	sort.Strings(got)
	want := []string{"bar", "heatmap", "line", "pie", "radar"}
	s.Equal(want, got)
}

func (s *BarSuite) TestRegistryRejectsDuplicate() {
	factory := func() charts.ChartConfig { return &Config{} }
	charts.Register("test_dup", factory)
	s.Panics(func() {
		charts.Register("test_dup", factory)
	})
}

func (s *BarSuite) TestNewUnknownType() {
	_, err := charts.New("graph")
	s.Error(err)
}

func (s *BarSuite) TestNewKnownType() {
	cfg, err := charts.New("bar")
	s.NoError(err)
	barCfg, ok := cfg.(*Config)
	s.Require().True(ok)
	s.Equal("bar", barCfg.ChartType())
}

func (s *BarSuite) TestDecodeBarRoundTrip() {
	original := Config{Type: "bar", Swap: "yxn", Scale: "log"}
	raw, err := json.Marshal(original)
	s.Require().NoError(err)

	cfg, err := charts.Decode("bar", raw)
	s.NoError(err)
	got, ok := cfg.(*Config)
	s.Require().True(ok)
	s.Equal(original, *got)
}

func (s *BarSuite) TestDecodeUnknownType() {
	_, err := charts.Decode("graph", json.RawMessage(`{"type":"graph"}`))
	s.Error(err)
}

func (s *BarSuite) TestMaterialiseBarPrecedence() {
	tr := true
	fa := false

	override := &Config{Swap: "yxn", Scale: "log", ShowLabels: &tr, AutoRotate: &tr}
	got := Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: false, AutoRotate: false}, override)
	s.Equal("yxn", got.Swap)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.AutoRotate)
	s.True(*got.AutoRotate)

	got = Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: true, AutoRotate: true}, nil)
	s.Equal("xyn", got.Swap)
	s.Equal("linear", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.AutoRotate)
	s.True(*got.AutoRotate)

	partial := &Config{Swap: "n"}
	got = Materialise(Flags{Swap: "xyn", Scale: "log", ShowLabels: true, AutoRotate: true}, partial)
	s.Equal("n", got.Swap)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.AutoRotate)
	s.True(*got.AutoRotate)

	got = Materialise(Flags{}, nil)
	s.Equal("", got.Swap)
	s.Equal("linear", got.Scale)
	s.Nil(got.ShowLabels)
	s.Nil(got.AutoRotate)
	s.Nil(got.Sort)

	got = Materialise(Flags{ShowLabels: true}, &Config{ShowLabels: &fa})
	s.Require().NotNil(got.ShowLabels)
	s.False(*got.ShowLabels)

	overrideSort := &shared.Sort{Enabled: true, Order: "desc"}
	got = Materialise(Flags{}, &Config{Sort: overrideSort})
	s.Require().NotNil(got.Sort)
	s.True(got.Sort.Enabled)
	s.Equal("desc", got.Sort.Order)

	got = Materialise(Flags{Sort: "asc"}, nil)
	s.Require().NotNil(got.Sort)
	s.True(got.Sort.Enabled)
	s.Equal("asc", got.Sort.Order)
}

func TestBarSuite(t *testing.T) {
	suite.Run(t, new(BarSuite))
}
