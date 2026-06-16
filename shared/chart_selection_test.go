package shared

import (
	"testing"

	config_charts "github.com/goptics/vizb/config/charts"
)

// stubChartConfig is a minimal ChartConfig used by the internal test below.
// The test only exercises DatasetNeeds3D, which reads ChartType() from each
// entry — the per-chart typed Configs are not required. The internal test
// file is in package shared, so it cannot import the per-chart subpackages
// (bar/line/pie/heatmap/radar each import shared, which would cycle).
type stubChartConfig struct {
	typ string
}

func (s stubChartConfig) ChartType() string { return s.typ }

func dsWith(types []string, zValues ...string) *Dataset {
	ds := &Dataset{}
	// Build Settings from the requested chart types. The test only cares
	// about ChartType() values, so a stub config that satisfies
	// config_charts.ChartConfig is sufficient — no need to go through the
	// registry (which lives in config/charts and imports shared, so we
	// can't import the per-chart subpackages from this internal test).
	for _, t := range types {
		ds.Settings = append(ds.Settings, stubChartConfig{typ: t})
	}
	for _, z := range zValues {
		ds.Data = append(ds.Data, DataPoint{XAxis: "x", YAxis: "y", ZAxis: z})
	}
	return ds
}

func TestChartsHave3DCapable(t *testing.T) {
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
		t.Run(c.name, func(t *testing.T) {
			if got := ChartsHave3DCapable(c.types); got != c.want {
				t.Fatalf("ChartsHave3DCapable(%v) = %v, want %v", c.types, got, c.want)
			}
		})
	}
}

func TestDatasetNeeds3D(t *testing.T) {
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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DatasetNeeds3D(c.ds); got != c.want {
				t.Fatalf("DatasetNeeds3D() = %v, want %v", got, c.want)
			}
		})
	}
}

// silence unused-import warning if Go's tooling ever flags it; the test
// references config_charts via the stub type assignment above.
var _ config_charts.ChartConfig = stubChartConfig{}
