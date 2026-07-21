import { describe, it, expect, beforeEach, vi } from 'vitest'
import { computed, ref, type Ref } from 'vue'
import type { Dataset, ChartConfig, ChartType } from '../types'

// The holder is set in beforeEach and read by the mock factories below.
const holder = vi.hoisted(() => ({
  ref: undefined as Ref<Dataset | undefined> | undefined,
  activeIndex: 0,
}))

vi.mock('./useDataPoint', () => ({
  get activeDataset() {
    if (!holder.ref) throw new Error('forgot beforeEach')
    return holder.ref
  },
  useDataPoint: () => ({
    activeDataset: holder.ref,
    activeDatasetId: { value: 0 },
    activeArrangement: { value: { identityString: 'xyz', targetString: 'xyz' } },
    getArrangement: () => undefined,
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

const ds = (settings: ChartConfig[], data: Dataset['data'] = []): Dataset => ({
  name: 'test',
  settings,
  data,
})

describe('useActiveChartShape', () => {
  beforeEach(() => {
    vi.resetModules()
    holder.activeIndex = 0
  })

  it('bar config returns scale/threeDRotate/showLabels defaults when fields are absent', async () => {
    holder.ref = ref(ds([{ type: 'bar' as ChartType }]))
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { scale, stack, threeDRotate, showLabels } = useActiveChartShape()
    expect(scale.value).toBe('linear')
    expect(stack.value).toBe(false)
    expect(threeDRotate.value).toBe(false)
    expect(showLabels.value).toBe(false)
  })

  it('pie config returns the same defaults without runtime branching', async () => {
    // PieConfig has no `scale` / `threeDRotate` field at all — `??` fallback
    // gives the same defaults. No `cfg.type === 'bar' || ...` guard needed.
    holder.ref = ref(ds([{ type: 'pie' as ChartType }]))
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { scale, threeDRotate, showLabels } = useActiveChartShape()
    expect(scale.value).toBe('linear')
    expect(threeDRotate.value).toBe(false)
    expect(showLabels.value).toBe(false)
  })

  it('hasThreeDOption is true for z-data bar when z is off chart axes', async () => {
    holder.ref = ref(
      ds(
        [{ type: 'bar' as ChartType, swap: 'xyn' }],
        [{ name: '', xAxis: 'a', yAxis: 'b', zAxis: 'z1', stats: [] }]
      )
    )
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { hasThreeDOption } = useActiveChartShape()
    expect(hasThreeDOption.value).toBe(true)
  })

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

  it('scatter config returns scale/threeDRotate/showLabels defaults when fields are absent', async () => {
    holder.ref = ref(ds([{ type: 'scatter' as ChartType }]))
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { scale, threeDRotate, showLabels } = useActiveChartShape()
    expect(scale.value).toBe('linear')
    expect(threeDRotate.value).toBe(false)
    expect(showLabels.value).toBe(false)
  })

  it('hasThreeDOption is true for z-data scatter when z is off chart axes', async () => {
    holder.ref = ref(
      ds(
        [{ type: 'scatter' as ChartType, swap: 'xyn' }],
        [{ name: '', xAxis: 'a', yAxis: 'b', zAxis: 'z1', stats: [] }]
      )
    )
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { hasThreeDOption } = useActiveChartShape()
    expect(hasThreeDOption.value).toBe(true)
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
})
