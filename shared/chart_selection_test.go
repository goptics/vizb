package shared

import "testing"

func dsWith(types []string, zValues ...string) *Dataset {
	ds := &Dataset{}
	// Build Settings from the requested chart types via the registry. The
	// test only cares about ChartType() values, so empty zero-value Configs
	// are sufficient.
	for _, t := range types {
		if cfg, err := NewChartConfig(t); err == nil {
			ds.Settings = append(ds.Settings, cfg)
		}
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
