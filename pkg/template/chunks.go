package template

import (
	"encoding/json"
	htmlTemplate "html/template"
)

// defaultCharts is the full chart set, used when no selection is supplied so
// nothing is over-pruned.
var defaultCharts = []string{"bar", "line", "pie", "heatmap", "radar"}

// gatedRoots returns the set of renderer chunk keys the pruner gates: a gated
// chunk is followed during reachability only when it is explicitly enabled.
func gatedRoots() map[string]bool {
	gated := make(map[string]bool, len(VizbChartRoots))
	for _, key := range VizbChartRoots {
		gated[key] = true
	}
	return gated
}

// reachableChunks walks the chunk reference graph from the enabled roots,
// following every edge except those into a gated renderer chunk that is not
// itself enabled. This keeps shared chunks (echarts core, vendor) whenever any
// consumer is kept, while dropping unselected chart renderers and — when the 3D
// root is not enabled — the echarts-gl stack reachable only through it.
func reachableChunks(enabled map[string]bool) map[string]bool {
	gated := gatedRoots()
	visited := make(map[string]bool, len(VizbChunks))
	queue := make([]string, 0, len(enabled))
	for key := range enabled {
		if !visited[key] {
			visited[key] = true
			queue = append(queue, key)
		}
	}

	for len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]
		for _, dep := range VizbChunkImports[key] {
			if visited[dep] {
				continue
			}
			// Skip a gated renderer chunk unless it is explicitly enabled.
			if gated[dep] && !enabled[dep] {
				continue
			}
			visited[dep] = true
			queue = append(queue, dep)
		}
	}

	return visited
}

// SelectChunks returns the gzip+base64 chunk map (as a JS object literal) holding
// only the chunks reachable from the selected charts. needs3D additionally keeps
// the 3D renderer + echarts-gl. An empty selection is treated as all charts.
func SelectChunks(charts []string, needs3D bool) htmlTemplate.JS {
	if len(charts) == 0 {
		charts = defaultCharts
	}

	enabled := map[string]bool{VizbEntryKey: true}
	for _, c := range charts {
		if key, ok := VizbChartRoots[c]; ok {
			enabled[key] = true
		}
	}
	if needs3D {
		if key, ok := VizbChartRoots["3d"]; ok {
			enabled[key] = true
		}
	}

	reachable := reachableChunks(enabled)
	out := make(map[string]string, len(reachable))
	for key := range reachable {
		if blob, ok := VizbChunks[key]; ok {
			out[key] = blob
		}
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		// VizbChunks values are base64 strings; marshaling cannot realistically
		// fail. Fall back to an empty object rather than emitting invalid JS.
		return htmlTemplate.JS("{}")
	}
	return htmlTemplate.JS(encoded)
}
