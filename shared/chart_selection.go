package shared

import (
	"encoding/json"
	"slices"

	config_charts "github.com/goptics/vizb/config/charts"
)

// ValidChartTypes is every chart type the CLI accepts via --charts.
var ValidChartTypes = []string{"bar", "line", "pie", "heatmap", "radar"}

// DefaultChartTypes is the --charts default when the user does not pass -c.
var DefaultChartTypes = []string{"bar", "line", "pie"}

// chartNeeds3D reports whether a selected chart type can ever render in 3D.
// Only bar and line have a 3D form; pie and heatmap always render 2D (pie folds
// dimensions into per-dimension pies, heatmap folds z onto the legend). This
// mirrors the UI routing in ui/src/components/ChartCard.vue.
func chartNeeds3D(chart string) bool {
	return chart == "bar" || chart == "line"
}

// ChartsHave3DCapable reports whether any selected chart type has a 3D form.
func ChartsHave3DCapable(charts []string) bool {
	return slices.ContainsFunc(charts, chartNeeds3D)
}

// DatasetHasZAxis reports whether any data point carries a z dimension. Mirrors
// the UI's chartHasZAxis check (ui/src/lib/utils.ts).
func DatasetHasZAxis(ds *Dataset) bool {
	for i := range ds.Data {
		if ds.Data[i].ZAxis != "" {
			return true
		}
	}
	return false
}

// SettingsHasThreeDOption reports whether any bar/line config was baked with
// threeD (via --3d), which unlocks value-mode 3D for x+y-only data.
func SettingsHasThreeDOption(settings []config_charts.ChartConfig) bool {
	for _, c := range settings {
		if configHasThreeDOption(c) {
			return true
		}
	}
	return false
}

func configHasThreeDOption(c config_charts.ChartConfig) bool {
	raw, err := json.Marshal(c)
	if err != nil {
		return false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	_, ok := m["threeD"]
	return ok
}

// DatasetNeeds3D reports whether a dataset will render a 3D chart: grouped 3D
// when z-axis data is present, or value 3D when threeD is baked on bar/line.
// Used to decide whether the echarts-gl (3D) chunk must ship.
func DatasetNeeds3D(ds *Dataset) bool {
	types := make([]string, len(ds.Settings))
	for i, c := range ds.Settings {
		types[i] = c.ChartType()
	}
	if !ChartsHave3DCapable(types) {
		return false
	}
	return DatasetHasZAxis(ds) || SettingsHasThreeDOption(ds.Settings)
}
