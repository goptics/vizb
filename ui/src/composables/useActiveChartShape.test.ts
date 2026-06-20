import { describe, it, expect, beforeEach, vi } from 'vitest'
import { computed, ref, type Ref } from 'vue'
import type { DataSet, ChartConfig, ChartType } from '../types'

// The holder is set in beforeEach and read by the mock factories below.
const holder = vi.hoisted(() => ({
  ref: undefined as Ref<DataSet | undefined> | undefined,
  activeIndex: 0,
}))

vi.mock('./useDataPoint', () => ({
  get activeDataSet() {
    if (!holder.ref) throw new Error('forgot beforeEach')
    return holder.ref
  },
  useDataPoint: () => ({
    activeDataSet: holder.ref,
    activeDataSetId: { value: 0 },
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

const ds = (settings: ChartConfig[], data: DataSet['data'] = []): DataSet => ({
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
    const { scale, threeDRotate, showLabels } = useActiveChartShape()
    expect(scale.value).toBe('linear')
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

  it('reads set values from the active config', async () => {
    holder.ref = ref(
      ds([
        {
          type: 'bar' as ChartType,
          scale: 'log',
          threeDRotate: true,
          showLabels: true,
        },
      ])
    )
    const { useActiveChartShape } = await import('./useActiveChartShape')
    const { scale, threeDRotate, showLabels } = useActiveChartShape()
    expect(scale.value).toBe('log')
    expect(threeDRotate.value).toBe(true)
    expect(showLabels.value).toBe(true)
  })
})
