import type { DataSet, ChartType } from './types'

declare global {
  interface Window {
    // One dataset object (common single-chart HTML) or an array for multi-tab.
    VIZB_DATA: DataSet | DataSet[]
    VIZB_VERSION: string
    VIZB_DATA_URL?: string
    // Charts bundled at generation time (--charts). The chunk pruner drops
    // renderer chunks for charts not listed here; useDataPoint intersects each
    // dataset's settings against this list on load.
    VIZB_CHARTS?: ChartType[]
  }
}

export {}
