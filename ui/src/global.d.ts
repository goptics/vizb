import type { DataSet, ChartType } from './types'

declare global {
  interface Window {
    VIZB_DATA: DataSet[]
    VIZB_VERSION: string
    VIZB_DATA_URL?: string
    // Charts bundled at generation time (--charts). The chunk pruner drops
    // renderer chunks for charts not listed here, so the UI must not surface a
    // tab whose chunk was pruned — setCharts() intersects against this.
    VIZB_CHARTS?: ChartType[]
  }
}

export {}
