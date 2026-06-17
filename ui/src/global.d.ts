import type { DataSet, ChartType } from './types'

declare global {
  interface Window {
    VIZB_DATA: DataSet[]
    VIZB_VERSION: string
    VIZB_DATA_URL?: string
    // Charts bundled at generation time (--charts). The chunk pruner drops
    // renderer chunks for charts not listed here; useDataPoint intersects each
    // dataset's settings against this list on load.
    VIZB_CHARTS?: ChartType[]
  }
}

export {}
