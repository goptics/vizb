package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ExpandSuite struct {
	suite.Suite
}

func TestExpandSuite(t *testing.T) {
	suite.Run(t, new(ExpandSuite))
}

func (s *ExpandSuite) TestExpandMultiStatOntoX() {
	in := []DataPoint{{
		YAxis: "100",
		Stats: []Stat{
			{Type: "default", Value: F64(3103.62)},
			{Type: "chi", Value: F64(3103.73)},
		},
	}}

	out := ExpandStatsOntoAxis(in, DimensionXAxis)
	s.Require().Len(out, 2)
	s.Equal("100", out[0].YAxis)
	s.Equal("default", out[0].XAxis)
	s.Require().Len(out[0].Stats, 1)
	s.Empty(out[0].Stats[0].Type)
	s.Equal(3103.62, *out[0].Stats[0].Value)
	s.Equal("chi", out[1].XAxis)
	s.Equal(3103.73, *out[1].Stats[0].Value)
}

func (s *ExpandSuite) TestExpandOntoYNameZ() {
	base := DataPoint{
		XAxis: "load",
		Stats: []Stat{{Type: "a", Value: F64(1)}, {Type: "b", Value: F64(2)}},
	}

	yOut := ExpandStatsOntoAxis([]DataPoint{base}, DimensionYAxis)
	s.Equal("a", yOut[0].YAxis)
	s.Equal("load", yOut[0].XAxis)

	nOut := ExpandStatsOntoAxis([]DataPoint{base}, DimensionName)
	s.Equal("a", nOut[0].Name)
	s.Equal("load", nOut[0].XAxis)

	zOut := ExpandStatsOntoAxis([]DataPoint{base}, DimensionZAxis)
	s.Equal("a", zOut[0].ZAxis)
}

func (s *ExpandSuite) TestExpandEmptyStatsPassthrough() {
	in := []DataPoint{{XAxis: "1", YAxis: "2", Stats: nil}}
	out := ExpandStatsOntoAxis(in, DimensionXAxis)
	s.Require().Len(out, 1)
	s.Equal("1", out[0].XAxis)
	s.Empty(out[0].Stats)
}

func (s *ExpandSuite) TestExpandNilStatValue() {
	out := ExpandStatsOntoAxis([]DataPoint{{
		Stats: []Stat{{Type: "a", Value: nil}},
	}}, DimensionXAxis)
	s.Require().Len(out, 1)
	s.Equal("a", out[0].XAxis)
	s.Require().Len(out[0].Stats, 1)
	s.Nil(out[0].Stats[0].Value)
}

func (s *ExpandSuite) TestExpandEmptyInput() {
	s.Empty(ExpandStatsOntoAxis(nil, DimensionXAxis))
	s.Empty(ExpandStatsOntoAxis([]DataPoint{}, DimensionXAxis))
}

func (s *ExpandSuite) TestExpandDoesNotMutateInput() {
	in := []DataPoint{{
		YAxis: "100",
		Stats: []Stat{{Type: "default", Value: F64(1)}},
	}}
	_ = ExpandStatsOntoAxis(in, DimensionXAxis)
	s.Equal("default", in[0].Stats[0].Type)
	s.Equal("100", in[0].YAxis)
	s.Empty(in[0].XAxis)
}

func (s *ExpandSuite) TestEmptyTypeOmitsFromJSON() {
	st := Stat{Value: F64(1.5)}
	b, err := json.Marshal(st)
	s.Require().NoError(err)
	s.JSONEq(`{"value":1.5}`, string(b))
	s.NotContains(string(b), "type")
}

func (s *ExpandSuite) TestEnsureAxisInsertsMissing() {
	axes := []Axis{{Key: "y", Label: "load"}}
	out := EnsureAxis(axes, DimensionXAxis)
	s.Equal([]string{"x", "y"}, axisKeys(out))
}

func (s *ExpandSuite) TestEnsureAxisIdempotent() {
	axes := []Axis{{Key: "x"}, {Key: "y"}}
	out := EnsureAxis(axes, DimensionXAxis)
	s.Equal([]string{"x", "y"}, axisKeys(out))
}
