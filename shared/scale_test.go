package shared_test

import (
	"encoding/json"
	"testing"

	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/internal/specparse"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// ScaleSuite covers the hybrid string|object Scale wire type and bag encode.
type ScaleSuite struct {
	suite.Suite
	fields []flags.ObjectField
}

func (s *ScaleSuite) SetupTest() {
	s.fields = []flags.ObjectField{
		{Name: "type", Kind: flags.KindString},
		{Name: "axes", Kind: flags.KindString},
		{Name: "base", Kind: flags.KindFloat},
		{Name: "baseX", Kind: flags.KindFloat},
		{Name: "baseY", Kind: flags.KindFloat},
		{Name: "baseZ", Kind: flags.KindFloat},
	}
}

func (s *ScaleSuite) TestMarshalStringForm() {
	raw, err := json.Marshal(shared.ScaleLog)
	s.Require().NoError(err)
	s.JSONEq(`"log"`, string(raw))

	raw, err = json.Marshal(shared.ScaleLinear)
	s.Require().NoError(err)
	s.JSONEq(`"linear"`, string(raw))
}

func (s *ScaleSuite) TestMarshalObjectWhenAxesSet() {
	sc := shared.Scale{Type: "log", Axes: []string{"x"}}
	raw, err := json.Marshal(sc)
	s.Require().NoError(err)
	s.JSONEq(`{"type":"log","axes":["x"],"base":10}`, string(raw))
}

func (s *ScaleSuite) TestMarshalObjectWhenNonDefaultBase() {
	base := 5.0
	sc := shared.Scale{Type: "log", Base: &base}
	raw, err := json.Marshal(sc)
	s.Require().NoError(err)
	s.JSONEq(`{"type":"log","base":5}`, string(raw))
}

func (s *ScaleSuite) TestUnmarshalStringAndObject() {
	var sc shared.Scale
	s.Require().NoError(json.Unmarshal([]byte(`"log"`), &sc))
	s.Equal(shared.ScaleLog, sc)

	s.Require().NoError(json.Unmarshal([]byte(`{"type":"log","axes":["x"],"base":10}`), &sc))
	s.Equal("log", sc.Type)
	s.Equal([]string{"x"}, sc.Axes)
	s.Require().NotNil(sc.Base)
	s.Equal(10.0, *sc.Base)
}

func (s *ScaleSuite) TestEncodeScaleValueStringBackCompat() {
	s.Equal("log", shared.EncodeScaleValue("log", s.fields))
	s.Equal("log", shared.EncodeScaleValue("LOG", s.fields))
	s.Equal("linear", shared.EncodeScaleValue("linear", s.fields))
}

func (s *ScaleSuite) TestEncodeScaleValueTypeLogNoAxesIsString() {
	s.Equal("log", shared.EncodeScaleValue("type=log", s.fields))
	s.Equal("log", shared.EncodeScaleValue("type=log;base=10", s.fields))
}

func (s *ScaleSuite) TestEncodeScaleValueAxesObject() {
	got := shared.EncodeScaleValue("type=log;axes=x", s.fields)
	s.Equal(map[string]any{
		"type": "log",
		"axes": []string{"x"},
		"base": 10.0,
	}, got)
}

func (s *ScaleSuite) TestEncodeScaleValuePerAxisBase() {
	got := shared.EncodeScaleValue("type=log;axes=x,y;baseX=5;baseY=10", s.fields)
	s.Equal(map[string]any{
		"type":  "log",
		"axes":  []string{"x", "y"},
		"base":  10.0,
		"baseX": 5.0,
	}, got)
}

func (s *ScaleSuite) TestEncodeScaleValueInvalidBaseWarnsAndDefaults() {
	for _, raw := range []string{
		"type=log;axes=x;base=1",
		"type=log;axes=x;base=0",
		"type=log;axes=x;base=-2",
	} {
		var got any
		out := testutil.CaptureStderr(func() {
			got = shared.EncodeScaleValue(raw, s.fields)
		})
		s.Contains(out, "Invalid scale base", raw)
		s.Equal(map[string]any{
			"type": "log",
			"axes": []string{"x"},
			"base": 10.0,
		}, got, raw)
	}
}

func (s *ScaleSuite) TestEncodeScaleValueUnknownAxisWarnsAndSkips() {
	var got any
	out := testutil.CaptureStderr(func() {
		got = shared.EncodeScaleValue("type=log;axes=x,q", s.fields)
	})
	s.Contains(out, "Invalid scale axis")
	s.Contains(out, "Skipping")
	s.Equal(map[string]any{
		"type": "log",
		"axes": []string{"x"},
		"base": 10.0,
	}, got)
}

func (s *ScaleSuite) TestIsScaleBag() {
	s.True(shared.IsScaleBag("type=log;axes=x"))
	s.True(shared.IsScaleBag("type=log"))
	s.False(shared.IsScaleBag("log"))
	s.False(shared.IsScaleBag("linear"))
	s.False(shared.IsScaleBag("log:axes=x"))
	s.True(shared.IsScaleBag("{type=log;axes=x}"))
}

func (s *ScaleSuite) TestParseObjectBagThenPayload() {
	props := []specparse.Prop{
		{Key: "type", Value: "log", HasValue: true},
		{Key: "axes", Value: "x", HasValue: true},
		{Key: "base", Value: "10", HasValue: true},
	}
	bag, err := shared.ParseObjectBag(props, s.fields)
	s.Require().NoError(err)
	got := shared.ScaleFromBag(bag).Payload()
	s.Equal(map[string]any{
		"type": "log",
		"axes": []string{"x"},
		"base": 10.0,
	}, got)
}

func (s *ScaleSuite) TestMarshalPerAxisBases() {
	baseY, baseZ := 2.0, 8.0
	sc := shared.Scale{Type: "log", Axes: []string{"y", "z"}, BaseY: &baseY, BaseZ: &baseZ}
	raw, err := json.Marshal(sc)
	s.Require().NoError(err)
	s.JSONEq(`{"type":"log","axes":["y","z"],"base":10,"baseY":2,"baseZ":8}`, string(raw))
}

func (s *ScaleSuite) TestUnmarshalJSONNullAndObjectFields() {
	var sc shared.Scale
	s.Require().NoError(json.Unmarshal([]byte(`null`), &sc))
	s.Equal(shared.Scale{}, sc)

	s.Require().NoError(json.Unmarshal([]byte(`{}`), &sc))
	s.Equal("log", sc.Type)

	s.Require().NoError(json.Unmarshal([]byte(`{"type":"log","baseY":2,"baseZ":8}`), &sc))
	s.Require().NotNil(sc.BaseY)
	s.Equal(2.0, *sc.BaseY)
	s.Require().NotNil(sc.BaseZ)
	s.Equal(8.0, *sc.BaseZ)
}

func (s *ScaleSuite) TestUnmarshalJSONErrors() {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"invalid string type", `"foo"`, `scale value "foo" is invalid`},
		{"number", `1`, "must be a string or object"},
		{"array", `[]`, "must be a string or object"},
		{"type not string", `{"type":1}`, "scale.type:"},
		{"type invalid", `{"type":"expo"}`, `scale.type "expo" is invalid`},
		{"axes not array", `{"axes":"x"}`, "must be an array of axis names"},
		{"base not number", `{"base":"10"}`, "scale.base:"},
		{"unknown field", `{"nope":1}`, `unknown field "nope"`},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			var sc shared.Scale
			err := json.Unmarshal([]byte(c.raw), &sc)
			s.Require().Error(err, c.raw)
			s.Contains(err.Error(), c.want, c.raw)
		})
	}
}

func (s *ScaleSuite) TestScaleFromBagInvalidTypeWarnsAndDefaults() {
	var got shared.Scale
	out := testutil.CaptureStderr(func() {
		got = shared.ScaleFromBag(map[string]any{"type": "expo", "axes": "x"})
	})
	s.Contains(out, "Invalid scale type")
	s.Equal("linear", got.Type)
	s.Equal([]string{"x"}, got.Axes)
}

func (s *ScaleSuite) TestScaleFromBagMissingTypeDefaultsToLog() {
	got := shared.ScaleFromBag(map[string]any{"axes": "x"})
	s.Equal("log", got.Type)
	s.Equal([]string{"x"}, got.Axes)
}

func (s *ScaleSuite) TestEncodeScaleValueDuplicateAxisSkipped() {
	got := shared.EncodeScaleValue("type=log;axes=x,x", s.fields)
	s.Equal(map[string]any{
		"type": "log",
		"axes": []string{"x"},
		"base": 10.0,
	}, got)
}

func (s *ScaleSuite) TestScaleFromBagInvalidBaseValues() {
	for _, raw := range []any{"nope", true} {
		var got shared.Scale
		out := testutil.CaptureStderr(func() {
			got = shared.ScaleFromBag(map[string]any{"type": "log", "base": raw})
		})
		s.Contains(out, "Invalid scale base", raw)
		s.Require().NotNil(got.Base, raw)
		s.Equal(10.0, *got.Base, raw)
	}
}

func TestScaleSuite(t *testing.T) {
	suite.Run(t, new(ScaleSuite))
}
