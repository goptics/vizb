// ECharts emphasis only lights the hovered data item, so with labels off a
// block hover would reveal one value. Highlight the block's whole series
// instead, and hand the column back to the axis highlight on leave.
//
// The group (axis) tooltip describes a whole column, which is noise while a
// single block is hovered, so it is hidden until the pointer leaves the block.
// Item tooltips (value-mode charts) are left alone: they are the only readout.
type ActionPayload = { type: string; seriesIndex?: number; dataIndex?: number }
type Chart = {
  dispatchAction: (payload: ActionPayload) => void
  setOption: (option: { tooltip: { show: boolean } }) => void
  getOption: () => object | undefined
}
type SeriesMouseEvent = { componentType?: string; seriesIndex?: number; dataIndex?: number }

const hasAxisTooltip = (chart: Chart): boolean => {
  const tooltip = (chart.getOption() as { tooltip?: unknown } | undefined)?.tooltip
  const first = (Array.isArray(tooltip) ? tooltip[0] : tooltip) as { trigger?: string } | undefined
  return first?.trigger === 'axis'
}

/** Mouse handlers for 2D category charts: block hover lights its series and hides the group tooltip. */
export function createSeriesHover(getChart: () => Chart | null | undefined) {
  const isSeries = (e: SeriesMouseEvent) =>
    e.componentType === 'series' && typeof e.seriesIndex === 'number'
  let tooltipHidden = false

  return {
    onMouseover(e: SeriesMouseEvent) {
      const chart = getChart()
      if (!chart || !isSeries(e)) return
      chart.dispatchAction({ type: 'downplay' })
      chart.dispatchAction({ type: 'highlight', seriesIndex: e.seriesIndex })
      if (hasAxisTooltip(chart)) {
        chart.dispatchAction({ type: 'hideTip' })
        chart.setOption({ tooltip: { show: false } })
        tooltipHidden = true
      }
    },
    onMouseout(e: SeriesMouseEvent) {
      const chart = getChart()
      if (!chart || !isSeries(e)) return
      chart.dispatchAction({ type: 'downplay' })
      if (typeof e.dataIndex === 'number') {
        chart.dispatchAction({ type: 'highlight', dataIndex: e.dataIndex })
      }
      if (tooltipHidden) {
        chart.setOption({ tooltip: { show: true } })
        tooltipHidden = false
      }
    },
  }
}
