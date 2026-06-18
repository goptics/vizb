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
}))

vi.mock('./useSettingsStore', () => ({
  useSettingsStore: () => ({
    activeConfig: computed(() => holder.ref?.value?.settings[holder.activeIndex]),
  }),
}))

const ds = (settings: ChartConfig[]): DataSet => ({
  name: 'test',
  settings,
  data: [],
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
