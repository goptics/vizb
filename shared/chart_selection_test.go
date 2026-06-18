package shared

import (
	"testing"

	config_charts "github.com/goptics/vizb/config/charts"
	"github.com/stretchr/testify/suite"
)

type stubChartConfig struct {
	typ string
}

func (s stubChartConfig) ChartType() string { return s.typ }

func dsWith(types []string, zValues ...string) *Dataset {
	ds := &Dataset{}
	for _, t := range types {
		ds.Settings = append(ds.Settings, stubChartConfig{typ: t})
	}
	for _, z := range zValues {
		ds.Data = append(ds.Data, DataPoint{XAxis: "x", YAxis: "y", ZAxis: z})
	}
	return ds
}

func dsWithThreeDOption(chartType string) *Dataset {
	cfg, err := config_charts.Decode(chartType, []byte(`{"type":"`+chartType+`","threeD":true}`))
	if err != nil {
		panic(err)
	}
	return &Dataset{Settings: []config_charts.ChartConfig{cfg}}
}

type ChartSelectionSuite struct {
	suite.Suite
}

func (s *ChartSelectionSuite) TestChartsHave3DCapable() {
	cases := []struct {
		name  string
		types []string
		want  bool
	}{
		{"bar is 3D-capable", []string{"bar"}, true},
		{"line is 3D-capable", []string{"line"}, true},
		{"pie is not", []string{"pie"}, false},
		{"heatmap is not", []string{"heatmap"}, false},
		{"pie+heatmap only", []string{"pie", "heatmap"}, false},
		{"mixed with bar", []string{"pie", "bar"}, true},
		{"empty", nil, false},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			s.Equal(c.want, ChartsHave3DCapable(c.types))
		})
	}
}

func (s *ChartSelectionSuite) TestDatasetNeeds3D() {
	cases := []struct {
		name string
		ds   *Dataset
		want bool
	}{
		{"z + bar => needs 3D", dsWith([]string{"bar"}, "1"), true},
		{"z + line => needs 3D", dsWith([]string{"line"}, "2"), true},
		{"z but pie only => no", dsWith([]string{"pie"}, "1"), false},
		{"z but heatmap only => no", dsWith([]string{"heatmap"}, "1"), false},
		{"bar but no z => no", dsWith([]string{"bar"}, "", ""), false},
		{"bar but empty data => no", dsWith([]string{"bar"}), false},
		{"mixed charts with z, one z point", dsWith([]string{"pie", "bar"}, "", "3"), true},
		{"threeD bar without z => needs 3D", dsWithThreeDOption("bar"), true},
		{"threeD line without z => needs 3D", dsWithThreeDOption("line"), true},
		{"threeD pie without z => no", dsWithThreeDOption("pie"), false},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			s.Equal(c.want, DatasetNeeds3D(c.ds))
		})
	}
}

var _ config_charts.ChartConfig = stubChartConfig{}

func TestChartSelectionSuite(t *testing.T) {
	suite.Run(t, new(ChartSelectionSuite))
}
