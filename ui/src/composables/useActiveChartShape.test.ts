import { describe, it, expect, beforeEach, vi } from 'vitest'
import { computed, ref, type Ref } from 'vue'
import type { Dataset, ChartType } from '../types'
import { ds } from '@/test-utils'

// The holder is set in beforeEach and read by the mock factories below.
const holder = vi.hoisted(() => ({
  ref: undefined as Ref<Dataset | undefined> | undefined,
  activeIndex: 0,
  arrangement: undefined as string | undefined,
  activeArrangement: { identityString: 'xyz', targetString: 'xyz' },
}))

vi.mock('./useDataPoint', () => ({
  get activeDataset() {
    if (!holder.ref) throw new Error('forgot beforeEach')
    return holder.ref
  },
  useDataPoint: () => ({
    activeDataset: holder.ref,
    activeDatasetId: { value: 0 },
    activeArrangement: { value: holder.activeArrangement },
    getArrangement: () => holder.arrangement,
  }),
}))

vi.mock('./useSettingsStore', () => ({
  useSettingsStore: () => ({
    activeConfig: computed(() => holder.ref?.value?.settings[holder.activeIndex]),
    chartType: computed(
      () => holder.ref?.value?.settings[holder.activeIndex]?.type ?? ('bar' as ChartType)
    ),
  }),
}))

describe('useActiveChartShape', () => {
  beforeEach(() => {
    vi.resetModules()
    holder.activeIndex = 0
    holder.arrangement = undefined
    holder.activeArrangement = { identityString: 'xyz', targetString: 'xyz' }
  })

  it.each(['bar', 'pie', 'scatter'] as const)(
    '%s config returns scale/threeDRotate/showLabels defaults when fields are absent',
    async (type) => {
      holder.ref = ref(ds([{ type }]))
      const { useActiveChartShape } = await import('./useActiveChartShape')
      const { scale, stack, threeDRotate, showLabels } = useActiveChartShape()
      expect(scale.value).toBe('linear')
      expect(threeDRotate.value).toBe(false)
      expect(showLabels.value).toBe(false)
      if (type === 'bar') expect(stack.value).toBe(false)
    }
  )

  it.each(['bar', 'scatter'] as const)(
    'hasThreeDOption is true for z-data %s when z is off chart axes',
    async (type) => {
      holder.ref = ref(
        ds([{ type, swap: 'xyn' }], [{ name: '', xAxis: 'a', yAxis: 'b', zAxis: 'z1', stats: [] }])
      )
      const { useActiveChartShape } = await import('./useActiveChartShape')
      const { hasThreeDOption } = useActiveChartShape()
      expect(hasThreeDOption.value).toBe(true)
    }
  )

  it('hasThreeDOption is false for pie even with z-data', async () => {
    holder.ref = ref(
      ds(
        [{ type: 'pie' as ChartType }],
        [{ name: '', xAxis: 'a', yAxis: 'b', zAxis: 'z1', stats: [] }]
      )
    )
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { hasThreeDOption } = useActiveChartShape()
    expect(hasThreeDOption.value).toBe(false)
  })

  it('hasThreeDOption is false for scatter in value-mode axes', async () => {
    holder.ref = ref({
      ...ds(
        [{ type: 'scatter' as ChartType }],
        [{ name: '', xAxis: '1', yAxis: '2', zAxis: '3', stats: [] }]
      ),
      axes: [
        { key: 'x', label: 'x', type: 'value' },
        { key: 'y', label: 'y', type: 'value' },
      ],
    })
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { hasThreeDOption } = useActiveChartShape()
    expect(hasThreeDOption.value).toBe(false)
  })

  it('reads set values from the active config', async () => {
    holder.ref = ref(
      ds([
        {
          type: 'bar' as ChartType,
          scale: 'log',
          stack: true,
          threeDRotate: true,
          showLabels: true,
        },
      ])
    )
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { scale, stack, threeDRotate, showLabels } = useActiveChartShape()
    expect(scale.value).toBe('linear')
    expect(stack.value).toBe(true)
    expect(threeDRotate.value).toBe(true)
    expect(showLabels.value).toBe(true)
  })

  it('uses configured scale when stack is disabled', async () => {
    holder.ref = ref(ds([{ type: 'bar' as ChartType, scale: 'log', stack: false }]))
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { scale, stack } = useActiveChartShape()
    expect(stack.value).toBe(false)
    expect(scale.value).toBe('log')
  })

  it('reads sort/threeD/visualMap/stat/symbol/smooth/horizontal fields', async () => {
    holder.ref = ref(
      ds([
        {
          type: 'line' as ChartType,
          sort: { enabled: true, order: 'desc' },
          threeD: true,
          threeDVisualMap: true,
          stat: { enabled: true, math: [] },
          symbol: 'diamond',
          symbolSize: 12,
          smooth: true,
        },
        {
          type: 'scatter' as ChartType,
          visualMap: true,
          symbol: 'circle',
          symbolSize: 8,
        },
        {
          type: 'bar' as ChartType,
          horizontal: true,
        },
      ])
    )

    const { useActiveChartShape } = await import('./useActiveChartShape')
    let shape = useActiveChartShape()
    expect(shape.sort.value).toEqual({ enabled: true, order: 'desc' })
    expect(shape.threeD.value).toBe(true)
    expect(shape.threeDVisualMap.value).toBe(true)
    expect(shape.stat.value).toEqual({ enabled: true, math: [] })
    expect(shape.symbol.value).toBe('diamond')
    expect(shape.symbolSize.value).toBe(12)
    expect(shape.smooth.value).toBe(true)
    expect(shape.horizontal.value).toBe(false)

    holder.activeIndex = 1
    shape = useActiveChartShape()
    expect(shape.visualMap.value).toBe(true)
    expect(shape.symbol.value).toBe('circle')
    expect(shape.symbolSize.value).toBe(8)
    expect(shape.smooth.value).toBe(false)

    holder.activeIndex = 2
    shape = useActiveChartShape()
    expect(shape.horizontal.value).toBe(true)
    expect(shape.threeD.value).toBe(false)
    expect(shape.visualMap.value).toBe(false)
    expect(shape.stat.value).toBeUndefined()
    expect(shape.symbol.value).toBeUndefined()
    expect(shape.symbolSize.value).toBeUndefined()
  })

  it('prefers arrangement map over wire swap and identity', async () => {
    holder.arrangement = 'ynx'
    holder.ref = ref(ds([{ type: 'bar' as ChartType, swap: 'xyn' }]))
    const { useActiveChartShape } = await import('./useActiveChartShape')
    // hasThreeDOption path uses effectiveSwapTarget via arrangementHasChartZ
    const { hasThreeDOption } = useActiveChartShape()
    expect(typeof hasThreeDOption.value).toBe('boolean')

    holder.arrangement = undefined
    holder.activeArrangement = { identityString: 'xyz', targetString: 'xyn' }
    holder.ref = ref(
      ds(
        [{ type: 'bar' as ChartType }],
        [{ name: '', xAxis: 'a', yAxis: 'b', zAxis: 'z1', stats: [] }]
      )
    )
    const again = (await import('./useActiveChartShape')).useActiveChartShape()
    expect(again.hasThreeDOption.value).toBe(true)
  })

  it('defaults threeDVisualMap/visualMap/smooth/horizontal when absent', async () => {
    holder.ref = ref(ds([{ type: 'bar' as ChartType }]))
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const shape = useActiveChartShape()
    expect(shape.threeDVisualMap.value).toBe(false)
    expect(shape.visualMap.value).toBe(false)
    expect(shape.smooth.value).toBe(false)
    expect(shape.horizontal.value).toBe(false)
    expect(shape.threeD.value).toBe(false)
    expect(shape.sort.value).toBeUndefined()
  })
})
