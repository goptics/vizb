package shared

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AggregateSuite struct {
	suite.Suite
}

func statVal(stats []Stat, typ string) (float64, bool) {
	for _, s := range stats {
		if s.Type == typ {
			if s.Value == nil {
				return 0, true
			}
			return *s.Value, true
		}
	}
	return 0, false
}

func (s *AggregateSuite) TestAggregateDataPointsSumsDuplicateKeys() {
	in := []DataPoint{
		{Name: "sales", XAxis: "2024-01-01", YAxis: "East", Stats: []Stat{{Type: "amount", Value: F64(100)}}},
		{Name: "sales", XAxis: "2024-01-01", YAxis: "East", Stats: []Stat{{Type: "amount", Value: F64(250)}}},
		{Name: "sales", XAxis: "2024-01-01", YAxis: "West", Stats: []Stat{{Type: "amount", Value: F64(40)}}},
	}

	out := AggregateDataPoints(in)
	s.Require().Len(out, 2)

	v, ok := statVal(out[0].Stats, "amount")
	s.Require().True(ok)
	s.Equal(350.0, v)
	s.Equal("East", out[0].YAxis)
}

func (s *AggregateSuite) TestAggregateDataPointsPreservesUniqueKeys() {
	in := []DataPoint{
		{XAxis: "A", Stats: []Stat{{Type: "v", Value: F64(1)}}},
		{XAxis: "B", Stats: []Stat{{Type: "v", Value: F64(2)}}},
	}

	out := AggregateDataPoints(in)
	s.Require().Len(out, 2)
}

func (s *AggregateSuite) TestAggregateDataPointsSumsPerStatType() {
	in := []DataPoint{
		{XAxis: "A", Stats: []Stat{{Type: "amount", Value: F64(10)}, {Type: "tax", Value: F64(1)}}},
		{XAxis: "A", Stats: []Stat{{Type: "amount", Value: F64(20)}, {Type: "tax", Value: F64(3)}}},
	}

	out := AggregateDataPoints(in)
	s.Require().Len(out, 1)
	v, _ := statVal(out[0].Stats, "amount")
	s.Equal(30.0, v)
	v, _ = statVal(out[0].Stats, "tax")
	s.Equal(4.0, v)
}

func (s *AggregateSuite) TestAggregateDataPointsDoesNotMutateInput() {
	in := []DataPoint{
		{XAxis: "A", Stats: []Stat{{Type: "v", Value: F64(5)}}},
		{XAxis: "A", Stats: []Stat{{Type: "v", Value: F64(7)}}},
	}

	AggregateDataPoints(in)
	s.Equal(5.0, *in[0].Stats[0].Value)
}

func TestAggregateSuite(t *testing.T) {
	suite.Run(t, new(AggregateSuite))
}
