import { describe, it, expect, vi, beforeEach } from 'vitest'
import { computed, ref } from 'vue'
import { mount } from '@vue/test-utils'
import type { BarConfig, ChartConfig, ChartType, LineConfig } from '@/types'

const { controlStub } = vi.hoisted(() => {
  const controlStub = (name: string) => ({
    default: {
      name,
      props: ['modelValue', 'disabled'],
      emits: ['update:modelValue'],
      template: `<div data-testid="${name}" :data-disabled="String(!!disabled)" :data-value="modelValue" />`,
    },
  })
  return { controlStub }
})

vi.mock('@/components/settings/SortControl.vue', () => controlStub('SortControl'))
vi.mock('@/components/settings/ScaleControl.vue', () => controlStub('ScaleControl'))
vi.mock('@/components/settings/StackControl.vue', () => controlStub('StackControl'))
vi.mock('@/components/settings/ShowLabelsControl.vue', () => controlStub('ShowLabelsControl'))
vi.mock('@/components/settings/SmoothControl.vue', () => controlStub('SmoothControl'))
vi.mock('@/components/settings/HorizontalControl.vue', () => controlStub('HorizontalControl'))
vi.mock('@/components/settings/ThreeDRotateControl.vue', () => controlStub('ThreeDRotateControl'))
vi.mock('@/components/settings/ThreeDControl.vue', () => controlStub('ThreeDControl'))
vi.mock('@/components/settings/ThreeDVisualMapControl.vue', () =>
  controlStub('ThreeDVisualMapControl')
)
vi.mock('@/components/settings/VisualMapControl.vue', () => controlStub('VisualMapControl'))
vi.mock('@/components/settings/SwapControl.vue', () => controlStub('SwapControl'))

vi.mock('./Selector.vue', () => ({
  default: {
    name: 'Selector',
    props: ['items', 'activeId'],
    emits: ['select'],
    template:
      '<button data-testid="chart-type-selector" @click="$emit(\'select\', 1)">{{ activeId }}</button>',
  },
}))

vi.mock('./SettingHeader.vue', () => ({
  default: {
    name: 'SettingHeader',
    props: ['label', 'description'],
    template: '<div data-testid="setting-header">{{ label }}</div>',
  },
}))

const holder = vi.hoisted(() => ({
  barConfig: {
    type: 'bar',
    sort: { enabled: false, order: 'asc' },
    scale: 'linear',
    stack: false,
    showLabels: false,
    swap: 'x/y',
  } as BarConfig,
  chartType: 'bar' as ChartType,
  settings: [] as ChartConfig[],
  data: [{ xAxis: 'a', yAxis: 'b', zAxis: '', stats: [{ type: 'v', value: 1 }] }],
  axes: undefined as { key: string; type?: string }[] | undefined,
  dimension: '2D' as '1D' | '2D' | '3D',
  arrangement: { targetString: 'x/y' },
  arrangementMap: undefined as string | undefined,
  groupId: 0,
  isValueMode: false,
  chartMode: 'grouped' as 'grouped' | 'value' | 'mixed',
  setChartType: vi.fn(),
  setSort: vi.fn(),
  setScale: vi.fn(),
  setStack: vi.fn(),
  setShowLabels: vi.fn(),
  setSmooth: vi.fn(),
  setHorizontal: vi.fn(),
  setThreeDRotate: vi.fn(),
  setSwap: vi.fn(),
  setThreeD: vi.fn(),
  setThreeDVisualMap: vi.fn(),
  setVisualMap: vi.fn(),
  setArrangement: vi.fn(),
  getArrangement: vi.fn(() => holder.arrangementMap),
  resetColor: vi.fn(),
}))

holder.settings = [holder.barConfig]

const barConfigRef = ref(holder.barConfig)
const chartTypeRef = ref(holder.chartType)
const groupIdRef = ref(holder.groupId)

vi.mock('@/composables/useSettingsStore', () => ({
  useSettingsStore: () => ({
    activeConfig: barConfigRef,
    chartType: chartTypeRef,
    setChartType: (t: ChartType) => {
      holder.setChartType(t)
      chartTypeRef.value = t
    },
    setSort: holder.setSort,
    setScale: holder.setScale,
    setStack: holder.setStack,
    setShowLabels: holder.setShowLabels,
    setSmooth: holder.setSmooth,
    setHorizontal: holder.setHorizontal,
    setThreeDRotate: holder.setThreeDRotate,
    setSwap: holder.setSwap,
    setThreeD: holder.setThreeD,
    setThreeDVisualMap: holder.setThreeDVisualMap,
    setVisualMap: holder.setVisualMap,
  }),
}))

const allowNullDataset = { value: false }
vi.mock('@/composables/useDataPoint', () => ({
  useDataPoint: () => ({
    activeDataset: computed(() =>
      allowNullDataset.value
        ? undefined
        : {
            settings: holder.settings,
            data: holder.data,
            axes: holder.axes,
          }
    ),
    activeDatasetId: ref('ds-1'),
    activeDataDimension: computed(() => holder.dimension),
    activeArrangement: computed(() => holder.arrangement),
    getArrangement: () => holder.getArrangement(),
    setArrangement: holder.setArrangement,
    activeGroupId: groupIdRef,
    isValueMode: computed(() => holder.isValueMode),
    chartMode: computed(() => holder.chartMode),
  }),
}))

vi.mock('@/lib/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/utils')>()
  return {
    ...actual,
    resetColor: () => holder.resetColor(),
    canOfferValue3D: () => true,
  }
})

vi.mock('@/lib/swap', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/swap')>()
  return {
    ...actual,
    arrangementHasChartZ: (t: string) => t.includes('z'),
  }
})

import SettingsPanel from './SettingsPanel.vue'

describe('SettingsPanel', () => {
  beforeEach(() => {
    holder.barConfig = {
      type: 'bar',
      sort: { enabled: false, order: 'asc' },
      scale: 'linear',
      stack: false,
      showLabels: false,
      swap: 'x/y',
    }
    barConfigRef.value = holder.barConfig
    holder.chartType = 'bar'
    chartTypeRef.value = 'bar'
    holder.settings = [holder.barConfig]
    holder.dimension = '2D'
    holder.arrangement = { targetString: 'x/y' }
    holder.arrangementMap = undefined
    holder.chartMode = 'grouped'
    holder.isValueMode = false
    holder.groupId = 0
    groupIdRef.value = 0
    vi.clearAllMocks()
    allowNullDataset.value = false
  })

  it('renders Settings card title', () => {
    const w = mount(SettingsPanel)
    expect(w.text()).toContain('Settings')
  })

  it('mounts getRenderableFields-driven bar 2D controls', () => {
    const w = mount(SettingsPanel)
    for (const name of [
      'SortControl',
      'ScaleControl',
      'StackControl',
      'ShowLabelsControl',
      'HorizontalControl',
      'SwapControl',
    ]) {
      expect(w.findComponent({ name }).exists()).toBe(true)
    }
  })

  it('shows chart type selector for multi-type datasets and switches', async () => {
    holder.settings = [
      holder.barConfig,
      {
        type: 'line',
        sort: { enabled: false, order: 'asc' },
        scale: 'linear',
        stack: false,
        showLabels: false,
        smooth: false,
        swap: 'x/y',
      } as LineConfig,
    ]
    const w = mount(SettingsPanel)
    expect(w.text()).toContain('Chart type')
    await w.get('[data-testid="chart-type-selector"]').trigger('click')
    expect(holder.setChartType).toHaveBeenCalledWith('line')
  })

  it('hides chart type when single setting', () => {
    const w = mount(SettingsPanel)
    expect(w.text()).not.toContain('Chart type')
  })

  it('filters sort in value mode and keeps swap', () => {
    holder.chartMode = 'value'
    holder.isValueMode = true
    const w = mount(SettingsPanel)
    expect(w.findComponent({ name: 'SortControl' }).exists()).toBe(false)
    expect(w.findComponent({ name: 'SwapControl' }).exists()).toBe(true)
  })

  it('filters swap in mixed mode', () => {
    holder.chartMode = 'mixed'
    holder.isValueMode = false
    const w = mount(SettingsPanel)
    expect(w.findComponent({ name: 'SwapControl' }).exists()).toBe(false)
  })

  it('disables scale when stack enabled and forces linear value', () => {
    holder.barConfig = { ...holder.barConfig, stack: true, scale: 'log' }
    barConfigRef.value = holder.barConfig
    holder.settings = [holder.barConfig]
    const w = mount(SettingsPanel)
    const scale = w.findComponent({ name: 'ScaleControl' })
    expect(scale.exists()).toBe(true)
    expect(scale.props('disabled')).toBe(true)
    expect(scale.props('modelValue')).toBe('linear')
  })

  it('handles field updates including swap side effects', async () => {
    const w = mount(SettingsPanel)
    const sort = w.findComponent({ name: 'SortControl' })
    await sort.vm.$emit('update:modelValue', { enabled: true, order: 'desc' })
    expect(holder.setSort).toHaveBeenCalled()

    const swap = w.findComponent({ name: 'SwapControl' })
    await swap.vm.$emit('update:modelValue', 'y/x')
    expect(holder.setArrangement).toHaveBeenCalledWith('ds-1', 'bar', 'y/x')
    expect(groupIdRef.value).toBe(0)
    expect(holder.resetColor).toHaveBeenCalled()
    expect(holder.setSwap).toHaveBeenCalledWith('y/x')

    await swap.vm.$emit('update:modelValue', undefined)

    // fire generic onUpdate for other keys through remaining controls
    for (const name of ['StackControl', 'ShowLabelsControl', 'HorizontalControl', 'ScaleControl']) {
      const c = w.findComponent({ name })
      if (c.exists()) await c.vm.$emit('update:modelValue', true)
    }
  })
  it('uses arrangement map and wire swap for z detection', () => {
    holder.arrangementMap = 'x/y/z'
    holder.barConfig = { ...holder.barConfig, threeD: true }
    barConfigRef.value = holder.barConfig
    const w = mount(SettingsPanel)
    expect(w.text()).toContain('Settings')
  })

  it('falls back wire swap when map empty', () => {
    holder.arrangementMap = undefined
    holder.barConfig = { ...holder.barConfig, swap: 'x/z/y', threeD: true }
    barConfigRef.value = holder.barConfig
    const w = mount(SettingsPanel)
    expect(w.text()).toContain('Settings')
    // Wire swap still drives a SwapControl when map is empty.
    expect(w.findComponent({ name: 'SwapControl' }).exists()).toBe(true)
  })

  it('empty activeConfig yields no fields', () => {
    // @ts-expect-error empty config branch
    barConfigRef.value = undefined
    const w = mount(SettingsPanel)
    expect(w.findComponent({ name: 'SortControl' }).exists()).toBe(false)
  })

  it('chart type select ignores invalid index', async () => {
    holder.settings = [
      holder.barConfig,
      {
        type: 'line',
        sort: { enabled: false, order: 'asc' },
        scale: 'linear',
        stack: false,
        showLabels: false,
        smooth: false,
        swap: 'x/y',
      } as LineConfig,
    ]
    const w = mount(SettingsPanel)
    const sel = w.findComponent({ name: 'Selector' })
    await sel.vm.$emit('select', 99)
    expect(holder.setChartType).not.toHaveBeenCalled()
  })
  it('valueFor returns undefined when config empty via setupState', () => {
    // @ts-expect-error empty
    barConfigRef.value = undefined
    const w = mount(SettingsPanel)
    const setup = (
      w.vm as unknown as {
        $: {
          setupState: {
            valueFor: (k: string) => unknown
            onUpdate: (k: string, v: unknown) => void
          }
        }
      }
    ).$.setupState
    expect(setup.valueFor('sort')).toBeUndefined()
    // onUpdate with empty config still routes to handlers that no-op safely
    setup.onUpdate('stack', true)
  })

  it('availableTypes empty when dataset missing settings map', () => {
    holder.settings = []
    const w = mount(SettingsPanel)
    expect(w.text()).not.toContain('Chart type')
  })

  it('availableTypes uses empty fallback when dataset undefined', () => {
    allowNullDataset.value = true
    const w = mount(SettingsPanel)
    expect(w.text()).not.toContain('Chart type')
    allowNullDataset.value = false
  })

  it('unknown chart type icon fallback and arrangement targetString', async () => {
    holder.arrangementMap = undefined
    holder.arrangement = { targetString: 'x/y' }
    holder.barConfig = {
      type: 'bar',
      sort: { enabled: false, order: 'asc' },
      scale: 'linear',
      stack: false,
      showLabels: false,
      swap: undefined as unknown as string,
      threeD: true,
    } as BarConfig
    barConfigRef.value = holder.barConfig
    holder.settings = [
      holder.barConfig,
      { type: 'not-a-chart' as ChartType, sort: { enabled: false, order: 'asc' } } as ChartConfig,
    ]
    holder.dimension = '3D'
    const w = mount(SettingsPanel)
    expect(w.text()).toContain('Chart type')
    // threeD group controls
    for (const name of ['ThreeDControl', 'ThreeDRotateControl', 'ThreeDVisualMapControl']) {
      const c = w.findComponent({ name })
      if (c.exists()) await c.vm.$emit('update:modelValue', true)
    }
  })
})
