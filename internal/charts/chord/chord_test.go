package chord_test

import (
	"encoding/json"
	"testing"

	_ "github.com/goptics/vizb/cmd/charts/chord"
	"github.com/goptics/vizb/internal/charts"
	chordchart "github.com/goptics/vizb/internal/charts/chord"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// ChordSuite covers the chord chart Config: its factory, JSON round-trip,
// applicable fields, and the no-inapplicable-fields JSON contract.
type ChordSuite struct {
	suite.Suite
}

func (s *ChordSuite) TestNewReturnsZeroConfig() {
	cfg := chordchart.New()
	got, ok := cfg.(*chordchart.Config)
	s.Require().True(ok)
	s.Empty(got.Type)
	s.Empty(got.Swap)
	s.Nil(got.Sort)
	s.Nil(got.ShowLabels)
	s.Nil(got.Stat)
}

func (s *ChordSuite) TestDecodeRoundTripAllApplicableFields() {
	original := chordchart.Config{
		Type:       chordchart.Type,
		Swap:       "xyn",
		Sort:       &shared.Sort{Enabled: true, Order: "asc"},
		ShowLabels: boolPtr(true),
		Stat:       &shared.StatConfig{Enabled: true, Math: []string{"shape", "counts"}},
	}
	raw, err := json.Marshal(original)
	s.Require().NoError(err)

	cfg, err := charts.Decode(chordchart.Type, raw)
	s.Require().NoError(err)
	got, ok := cfg.(*chordchart.Config)
	s.Require().True(ok)
	s.Equal(original, *got)
	s.Equal(chordchart.Type, got.ChartType())
}

func (s *ChordSuite) TestJSONOmitsInapplicableChartFields() {
	cfg := chordchart.Config{
		Type:       chordchart.Type,
		Swap:       "xyn",
		Sort:       &shared.Sort{Enabled: true, Order: "desc"},
		ShowLabels: boolPtr(false),
		Stat:       &shared.StatConfig{Enabled: true, Math: []string{"counts"}},
	}
	raw, err := json.Marshal(cfg)
	s.Require().NoError(err)

	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))
	for _, key := range []string{"scale", "stack", "threeD", "threeDRotate", "threeDVisualMap", "visualMap", "smooth", "horizontal"} {
		_, ok := m[key]
		s.False(ok, "chord JSON must not carry %q", key)
	}
	s.Equal(chordchart.Type, m["type"])
	s.Equal("xyn", m["swap"])
}

func boolPtr(b bool) *bool { return &b }

func TestChordSuite(t *testing.T) {
	suite.Run(t, new(ChordSuite))
}
