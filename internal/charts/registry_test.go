package charts_test

import (
	"encoding/json"
	"sort"
	"testing"

	_ "github.com/goptics/vizb/cmd/charts/bar"
	_ "github.com/goptics/vizb/cmd/charts/heatmap"
	_ "github.com/goptics/vizb/cmd/charts/line"
	_ "github.com/goptics/vizb/cmd/charts/pie"
	_ "github.com/goptics/vizb/cmd/charts/radar"
	_ "github.com/goptics/vizb/cmd/charts/scatter"
	"github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
	heatmapchart "github.com/goptics/vizb/internal/charts/heatmap"
	linechart "github.com/goptics/vizb/internal/charts/line"
	piechart "github.com/goptics/vizb/internal/charts/pie"
	radarchart "github.com/goptics/vizb/internal/charts/radar"
	scatterchart "github.com/goptics/vizb/internal/charts/scatter"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// RegistrySuite covers the shared chart config registry in package charts.
type RegistrySuite struct {
	suite.Suite
}

func (s *RegistrySuite) TestRegistryListsChartTypes() {
	got := charts.Registered()
	sort.Strings(got)
	want := []string{"bar", "heatmap", "line", "pie", "radar", "scatter"}
	s.Equal(want, got)
}

func (s *RegistrySuite) TestRegistryRejectsDuplicate() {
	spec := charts.Spec{Type: "test_dup", Factory: func() charts.ChartConfig { return &barchart.Config{} }}
	charts.Register(spec)
	s.Panics(func() {
		charts.Register(spec)
	})
}

func (s *RegistrySuite) TestNewUnknownType() {
	_, err := charts.New("graph")
	s.Error(err)
}

func (s *RegistrySuite) TestNewKnownType() {
	cfg, err := charts.New("bar")
	s.NoError(err)
	barCfg, ok := cfg.(*barchart.Config)
	s.Require().True(ok)
	s.Equal("bar", barCfg.ChartType())
}

func (s *RegistrySuite) TestSmoothFlagIsLineOnly() {
	flagNames := func(chartType string) map[string]bool {
		out := map[string]bool{}
		for _, f := range charts.FlagsFor(chartType) {
			out[f.EffectiveKey()] = true
		}
		return out
	}

	s.True(flagNames("line")["smooth"])
	for _, chartType := range []string{"bar", "scatter", "pie", "heatmap", "radar"} {
		s.False(flagNames(chartType)["smooth"], "%s should not register smooth", chartType)
	}
}

func (s *RegistrySuite) TestHorizontalFlagIsBarOnly() {
	flagNames := func(chartType string) map[string]bool {
		out := map[string]bool{}
		for _, f := range charts.FlagsFor(chartType) {
			out[f.EffectiveKey()] = true
		}
		return out
	}

	s.True(flagNames("bar")["horizontal"])
	for _, chartType := range []string{"line", "scatter", "pie", "heatmap", "radar"} {
		s.False(flagNames(chartType)["horizontal"], "%s should not register horizontal", chartType)
	}
}

func (s *RegistrySuite) TestNewScatterKnownType() {
	cfg, err := charts.New("scatter")
	s.NoError(err)
	scatterCfg, ok := cfg.(*scatterchart.Config)
	s.Require().True(ok)
	s.Equal("scatter", scatterCfg.ChartType())
}

func (s *RegistrySuite) TestDecodeBarRoundTrip() {
	original := barchart.Config{Type: "bar", Swap: "yxn", Scale: "log"}
	raw, err := json.Marshal(original)
	s.Require().NoError(err)

	cfg, err := charts.Decode("bar", raw)
	s.NoError(err)
	got, ok := cfg.(*barchart.Config)
	s.Require().True(ok)
	s.Equal(original, *got)
}

func (s *RegistrySuite) TestDecodeUnknownType() {
	_, err := charts.Decode("graph", json.RawMessage(`{"type":"graph"}`))
	s.Error(err)
}

func (s *RegistrySuite) TestDecodeInvalidJSON() {
	_, err := charts.Decode("bar", json.RawMessage(`{invalid`))
	s.Error(err)
}

func (s *RegistrySuite) TestGetKnownAndUnknown() {
	spec, ok := charts.Get("bar")
	s.True(ok)
	s.Equal("bar", spec.Type)

	_, ok = charts.Get("graph")
	s.False(ok)
}

func (s *RegistrySuite) TestChartConfigAccessors() {
	stat := &shared.StatConfig{Enabled: true, Math: []string{"counts"}}
	cases := []charts.ChartConfig{
		&barchart.Config{Type: "bar", Swap: "yxn", Stat: stat},
		&linechart.Config{Type: "line", Swap: "nxy", Stat: stat},
		&scatterchart.Config{Type: "scatter", Swap: "xyn", Stat: stat},
		&piechart.Config{Type: "pie", Swap: "ynx", Stat: stat},
		&heatmapchart.Config{Type: "heatmap", Swap: "xy", Stat: stat},
		&radarchart.Config{Type: "radar", Swap: "yx", Stat: stat},
	}
	for _, cfg := range cases {
		s.True(cfg.StatEnabled())
		s.Equal([]string{"counts"}, cfg.StatMath())
		s.NotEmpty(cfg.SwapString())
	}
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}
