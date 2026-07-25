import { describe, it, expect, beforeEach, vi } from 'vitest'
import { computed, ref, type Ref } from 'vue'
import type { Dataset, ChartType } from '../types'
import { ds } from '@/test-utils'

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

describe('useActiveChartShape', () => {
  beforeEach(() => {
    vi.resetModules()
    holder.activeIndex = 0
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
})
