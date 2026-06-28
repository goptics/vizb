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
	"github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
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

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}
