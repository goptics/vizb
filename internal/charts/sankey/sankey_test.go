package sankey_test

import (
	"encoding/json"
	"testing"

	_ "github.com/goptics/vizb/cmd/charts/sankey"
	"github.com/goptics/vizb/internal/charts"
	sankeychart "github.com/goptics/vizb/internal/charts/sankey"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// SankeySuite covers the sankey chart Config: its factory, JSON round-trip,
// and the "no 3D / no inapplicable flags" JSON contract.
type SankeySuite struct {
	suite.Suite
}

func (s *SankeySuite) TestNewReturnsZeroConfig() {
	cfg := sankeychart.New()
	got, ok := cfg.(*sankeychart.Config)
	s.Require().True(ok)
	s.Empty(got.Type)
	s.Empty(got.Swap)
	s.Nil(got.Sort)
	s.Nil(got.ShowLabels)
	s.Nil(got.Stat)
}

func (s *SankeySuite) TestDecodeRoundTripAllFields() {
	original := sankeychart.Config{
		Type:       "sankey",
		Swap:       "xyn",
		Sort:       &shared.Sort{Enabled: true, Order: "asc"},
		ShowLabels: boolPtr(true),
		Stat:       &shared.StatConfig{Enabled: true, Math: []string{"shape", "counts"}},
	}
	raw, err := json.Marshal(original)
	s.Require().NoError(err)

	cfg, err := charts.Decode("sankey", raw)
	s.Require().NoError(err)
	got, ok := cfg.(*sankeychart.Config)
	s.Require().True(ok)
	s.Equal(original, *got)
	s.Equal("sankey", got.ChartType())
}

func (s *SankeySuite) TestJSONOmitsInapplicableChartFields() {
	cfg := sankeychart.Config{
		Type:       "sankey",
		Swap:       "xyn",
		Sort:       &shared.Sort{Enabled: true, Order: "desc"},
		ShowLabels: boolPtr(false),
		Stat:       &shared.StatConfig{Enabled: true, Math: []string{"counts"}},
	}
	raw, err := json.Marshal(cfg)
	s.Require().NoError(err)

	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	// Sankey is a flow layout: scale/stack/3D/visualMap never apply and must
	// not leak into the emitted config JSON.
	for _, key := range []string{"scale", "stack", "threeD", "threeDRotate", "threeDVisualMap", "visualMap", "smooth", "horizontal"} {
		_, ok := m[key]
		s.False(ok, "sankey JSON must not carry %q", key)
	}
	s.Equal("sankey", m["type"])
	s.Equal("xyn", m["swap"])
}

func boolPtr(b bool) *bool { return &b }

func TestSankeySuite(t *testing.T) {
	suite.Run(t, new(SankeySuite))
}
