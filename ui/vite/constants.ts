// Dynamic-import chunk filename prefix → logical chart name. Rollup names a
// dynamic-import chunk after its module, so "ChartBar-<hash>.js" → "bar". These
// are the only chunks the Go pruner gates; everything else (shared echarts core,
// vendor) is always kept when reachable.
export const CHART_ROOT_PREFIX: Record<string, string> = {
  ChartBar: 'bar',
  ChartLine: 'line',
  ChartPie: 'pie',
  ChartHeatmap: 'heatmap',
  ChartRadar: 'radar',
  Chart3D: '3d',
}