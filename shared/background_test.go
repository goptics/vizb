package shared

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/internal/specparse"
	"github.com/stretchr/testify/suite"
)

// BackgroundSuite covers the Background wire type (marshal omitempty) and the
// object-bag converters shared by --chart braces and the raw --bg flag form.
type BackgroundSuite struct {
	suite.Suite
	fields []flags.ObjectField
}

func (s *BackgroundSuite) SetupTest() {
	s.fields = []flags.ObjectField{
		{Name: "color", Kind: flags.KindString},
		{Name: "borderWidth", Kind: flags.KindFloat},
		{Name: "opacity", Kind: flags.KindFloat},
	}
}

func (s *BackgroundSuite) TestUnmarshalFullWire() {
	raw := `{
		"active": true,
		"color": "rgba(180, 180, 180, 0.2)",
		"borderColor": "#000",
		"borderWidth": 0,
		"borderType": "solid",
		"borderRadius": [0],
		"shadowBlur": 10,
		"shadowColor": "rgba(0, 0, 0, 0.5)",
		"shadowOffsetX": 0,
		"shadowOffsetY": 0,
		"opacity": 1
	}`
	var bg Background
	s.Require().NoError(json.Unmarshal([]byte(raw), &bg))
	s.True(bg.Active)
	s.Equal("rgba(180, 180, 180, 0.2)", bg.Color)
	s.Equal("#000", bg.BorderColor)
	s.Require().NotNil(bg.BorderWidth)
	s.Equal(float64(0), *bg.BorderWidth)
	s.Equal("solid", bg.BorderType)
	s.Require().NotNil(bg.BorderRadius)
	s.Equal(BorderRadius{0}, *bg.BorderRadius)
	s.Require().NotNil(bg.ShadowBlur)
	s.Equal(float64(10), *bg.ShadowBlur)
	s.Equal("rgba(0, 0, 0, 0.5)", bg.ShadowColor)
	s.Require().NotNil(bg.ShadowOffsetX)
	s.Equal(float64(0), *bg.ShadowOffsetX)
	s.Require().NotNil(bg.ShadowOffsetY)
	s.Equal(float64(0), *bg.ShadowOffsetY)
	s.Require().NotNil(bg.Opacity)
	s.Equal(float64(1), *bg.Opacity)
}

func (s *BackgroundSuite) TestUnmarshalInactiveRoundTrips() {
	var bg Background
	s.Require().NoError(json.Unmarshal([]byte(`{"active": false, "opacity": 0.5}`), &bg))
	s.False(bg.Active)
	s.Equal(float64(0.5), *bg.Opacity)

	out, err := json.Marshal(bg)
	s.Require().NoError(err)
	s.JSONEq(`{"active":false,"opacity":0.5}`, string(out))
}

func (s *BackgroundSuite) TestMarshalOmitsUnsetFields() {
	width := 0.0
	out, err := json.Marshal(&Background{Active: true, BorderWidth: &width})
	s.Require().NoError(err)
	// An explicit zero survives (pointer field); unset fields are omitted.
	s.JSONEq(`{"active":true,"borderWidth":0}`, string(out))
}

func (s *BackgroundSuite) TestParseObjectBag() {
	props := []specparse.Prop{
		{Key: "color", Value: "rgba(1,2,3,0.5)", HasValue: true},
		{Key: "borderWidth", Value: "2", HasValue: true},
	}
	bag, err := ParseObjectBag(props, s.fields)
	s.Require().NoError(err)
	s.Equal(map[string]any{"color": "rgba(1,2,3,0.5)", "borderWidth": "2"}, bag)
}

func (s *BackgroundSuite) TestParseObjectBagRejectsUnknownKey() {
	_, err := ParseObjectBag([]specparse.Prop{{Key: "decal", Value: "1", HasValue: true}}, s.fields)
	s.Require().Error(err)
	s.Contains(err.Error(), `unknown object field "decal"`)
	s.Contains(err.Error(), "color") // valid fields listed in the message
}

func (s *BackgroundSuite) TestParseObjectBagRejectsBareField() {
	_, err := ParseObjectBag([]specparse.Prop{{Key: "color"}}, s.fields)
	s.Require().Error(err)
	s.Contains(err.Error(), "requires a value")
}

func (s *BackgroundSuite) TestParseObjectBagValidatesFields() {
	strict := []flags.ObjectField{{Name: "opacity", Kind: flags.KindFloat, Validate: func(v string) error {
		if v != "0.5" {
			return fmt.Errorf("opacity %q is out of range", v)
		}
		return nil
	}}}
	_, err := ParseObjectBag([]specparse.Prop{{Key: "opacity", Value: "0.9", HasValue: true}}, strict)
	s.Require().Error(err)
	s.Contains(err.Error(), "out of range")

	bag, err := ParseObjectBag([]specparse.Prop{{Key: "opacity", Value: "0.5", HasValue: true}}, strict)
	s.Require().NoError(err)
	s.Equal(map[string]any{"opacity": "0.5"}, bag)
}

func (s *BackgroundSuite) TestParseObjectBagEncodesFields() {
	encoded := []flags.ObjectField{{Name: "opacity", Kind: flags.KindFloat, Encode: func(v any) any { return 0.25 }}}
	bag, err := ParseObjectBag([]specparse.Prop{{Key: "opacity", Value: "0.25", HasValue: true}}, encoded)
	s.Require().NoError(err)
	s.Equal(map[string]any{"opacity": 0.25}, bag)
}

func (s *BackgroundSuite) TestParseObjectBagEmpty() {
	bag, err := ParseObjectBag(nil, s.fields)
	s.Require().NoError(err)
	s.Empty(bag)
}

func (s *BackgroundSuite) TestParseObjectBagString() {
	bag, err := ParseObjectBagString("color=#fff;borderWidth=2", s.fields)
	s.Require().NoError(err)
	s.Equal(map[string]any{"color": "#fff", "borderWidth": "2"}, bag)

	_, err = ParseObjectBagString("{color=#fff}", s.fields)
	s.Require().Error(err)
	s.Contains(err.Error(), "braces are not allowed")
}

func TestBackgroundSuite(t *testing.T) {
	suite.Run(t, new(BackgroundSuite))
}
