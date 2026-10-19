// ECharts emphasis only lights the hovered data item, so with labels off a
// block hover would reveal one value. Highlight the block's whole series
// instead, and hand the column back to the axis highlight on leave.
type Dispatcher = {
  dispatchAction: (payload: { type: string; seriesIndex?: number; dataIndex?: number }) => void
}
type SeriesMouseEvent = { componentType?: string; seriesIndex?: number; dataIndex?: number }

export function createSeriesHover(getChart: () => Dispatcher | null | undefined) {
  const isSeries = (e: SeriesMouseEvent) =>
    e.componentType === 'series' && typeof e.seriesIndex === 'number'

  return {
    onMouseover(e: SeriesMouseEvent) {
      const chart = getChart()
      if (!chart || !isSeries(e)) return
      chart.dispatchAction({ type: 'downplay' })
      chart.dispatchAction({ type: 'highlight', seriesIndex: e.seriesIndex })
    },
    onMouseout(e: SeriesMouseEvent) {
      const chart = getChart()
      if (!chart || !isSeries(e)) return
      chart.dispatchAction({ type: 'downplay' })
      if (typeof e.dataIndex === 'number') {
        chart.dispatchAction({ type: 'highlight', dataIndex: e.dataIndex })
      }
    },
  }
}
