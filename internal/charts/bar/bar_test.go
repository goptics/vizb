package bar

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type BarConfigSuite struct {
	suite.Suite
}

func (s *BarConfigSuite) TestBorderRadiusDefaultsNil() {
	cfg := New().(*Config)
	s.Nil(cfg.BorderRadius)
}

func (s *BarConfigSuite) TestBorderRadiusField() {
	r := shared.BorderRadius{8}
	cfg := &Config{Type: Type, BorderRadius: &r}
	s.Equal("bar", cfg.ChartType())
	s.Equal(shared.BorderRadius{8}, *cfg.BorderRadius)

	r = shared.BorderRadius{8, 8, 0, 0}
	cfg.BorderRadius = &r
	s.Equal(shared.BorderRadius{8, 8, 0, 0}, *cfg.BorderRadius)
}

func (s *BarConfigSuite) TestBackgroundDefaultsNil() {
	cfg := New().(*Config)
	s.Nil(cfg.Background)

	// Omitted --bg leaves the marshalled JSON without a background field.
	raw, err := json.Marshal(cfg)
	s.Require().NoError(err)
	s.NotContains(string(raw), "background")
}

func (s *BarConfigSuite) TestBackgroundFieldRoundTrip() {
	width := 0.0
	radius := shared.BorderRadius{8, 8, 0, 0}
	cfg := &Config{Type: Type, Background: &shared.Background{
		Active:       true,
		Color:        "rgba(180, 180, 180, 0.2)",
		BorderColor:  "#000",
		BorderWidth:  &width,
		BorderRadius: &radius,
	}}
	raw, err := json.Marshal(cfg)
	s.Require().NoError(err)
	s.Contains(string(raw), `"background":{"active":true`)
	s.Contains(string(raw), `"borderWidth":0`)

	decoded := New().(*Config)
	s.Require().NoError(json.Unmarshal(raw, decoded))
	s.Require().NotNil(decoded.Background)
	s.True(decoded.Background.Active)
	s.Equal("rgba(180, 180, 180, 0.2)", decoded.Background.Color)
	s.Equal(float64(0), *decoded.Background.BorderWidth)
	s.Equal(shared.BorderRadius{8, 8, 0, 0}, *decoded.Background.BorderRadius)
}

func (s *BarConfigSuite) TestBackgroundStrictDecodeRejectsInvalid() {
	raws := []string{
		`{"type":"bar","background":{"decal":1}}`,
		`{"type":"bar","background":{"borderType":"dash"}}`,
		`{"type":"bar","background":{"opacity":2}}`,
	}
	for _, raw := range raws {
		cfg := New().(*Config)
		s.Error(json.Unmarshal([]byte(raw), cfg), raw)
	}
}

func TestBarConfigSuite(t *testing.T) {
	suite.Run(t, new(BarConfigSuite))
}
