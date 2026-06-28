import type { ChartType } from '../types'

// Closed enumeration of chart types vizb knows how to render. Single source of
// truth on the UI side — the wire format's `settings[i].type` discriminator is
// the matching axis on the Go side. Keep this in sync with the registered Go
// chart configs (`internal/charts/*`).
export const ALL_CHART_TYPES: ChartType[] = ['bar', 'line', 'scatter', 'pie', 'heatmap', 'radar']
