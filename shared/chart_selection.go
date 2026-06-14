package shared

import "slices"

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

// DatasetNeeds3D reports whether a dataset will render a 3D chart, matching the
// UI's is3D = hasX && hasY && hasZ combined with the bar/line-only 3D routing.
// Used to decide whether the echarts-gl (3D) chunk must ship.
func DatasetNeeds3D(ds *Dataset) bool {
	return DatasetHasZAxis(ds) && ChartsHave3DCapable(ds.Settings.Charts)
}
