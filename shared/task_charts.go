package shared

import "github.com/go-echarts/go-echarts/v2/charts"

// BenchCharts represents charts grouped by task name
type BenchCharts struct {
	Name             string
	NsPerOpChart     *charts.Bar
	BytesPerOpChart  *charts.Bar
	AllocsPerOpChart *charts.Bar
}
