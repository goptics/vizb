// Hover behaviour for 2D category charts (bar / line / scatter):
// - A block hover highlights the block's whole series, so with labels off its
//   values show in every column (ECharts would only emphasise the one item).
// - The group (axis) tooltip summarises a whole column, which is noise while a
//   single block is hovered. It is hidden with a CSS class on the chart
//   wrapper (see GROUP_TOOLTIP_CLASS) rather than via tooltip options: toggling
//   tooltip state makes the axis pointer drop its highlight batch, which also
//   un-emphasises the hovered block.
// Talks to the raw ECharts instance (vue-echarts' setOption needs manual-update).
type ActionPayload = { type: string; seriesIndex?: number; dataIndex?: number }
type Chart = { dispatchAction: (payload: ActionPayload) => void }
type SeriesMouseEvent = { componentType?: string; seriesIndex?: number; dataIndex?: number }

// Long enough to cross the gap between bars of one group without the card
// flashing, short enough to feel immediate when the pointer settles in a gap.
export const LEAVE_DELAY_MS = 120

/** Mouse handlers: block hover lights its series and flags block-hover for the tooltip CSS. */
export function createSeriesHover(
  getChart: () => Chart | null | undefined,
  setBlockHover: (on: boolean) => void
) {
  const isSeries = (e: SeriesMouseEvent) =>
    e.componentType === 'series' && typeof e.seriesIndex === 'number'
  let hoveredSeries: number | undefined
  let pendingLeave: ReturnType<typeof setTimeout> | undefined

  const reset = () => {
    clearTimeout(pendingLeave)
    pendingLeave = undefined
    hoveredSeries = undefined
    setBlockHover(false)
  }

  return {
    onMouseover(e: SeriesMouseEvent) {
      const chart = getChart()
      if (!chart || !isSeries(e)) return
      clearTimeout(pendingLeave)
      setBlockHover(true)
      // Highlighting redraws the bar under a still pointer (mouseout + mouseover);
      // re-dispatching for the same series would loop.
      if (hoveredSeries !== e.seriesIndex) {
        hoveredSeries = e.seriesIndex
        chart.dispatchAction({ type: 'downplay' })
        chart.dispatchAction({ type: 'highlight', seriesIndex: e.seriesIndex })
      }
    },
    onMouseout(e: SeriesMouseEvent) {
      if (!getChart() || !isSeries(e)) return
      clearTimeout(pendingLeave)
      pendingLeave = setTimeout(() => {
        const chart = getChart()
        reset()
        if (!chart) return
        // Back in the column gap: show that column's values again.
        chart.dispatchAction({ type: 'downplay' })
        if (typeof e.dataIndex === 'number') {
          chart.dispatchAction({ type: 'highlight', dataIndex: e.dataIndex })
        }
      }, LEAVE_DELAY_MS)
    },
    /** Pointer left the chart: drop any pending column restore and all highlight. */
    onGlobalout() {
      const chart = getChart()
      reset()
      chart?.dispatchAction({ type: 'downplay' })
    },
  }
}
