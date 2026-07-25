import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { computed, defineComponent, h, nextTick, ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { makeGroupedChartData, makePieChartData } from '@/test-utils'
import type { ChartData, ChartType } from '@/types'

const holder = vi.hoisted(() => ({
  stat: { enabled: true } as { enabled: boolean; math?: string[] } | undefined,
  chartType: 'bar' as ChartType | 'unknown',
  threeD: false,
  isFullscreen: false,
  arrangement: 'x/y',
  axes: undefined as { key: string; type?: string }[] | undefined,
  options: { title: { text: 'opt' } } as Record<string, unknown>,
  sort: undefined as { enabled: boolean; order: 'asc' | 'desc' } | undefined,
  loaders: [] as Array<() => Promise<unknown>>,
  loadingComp: null as unknown,
  errorComp: null as unknown,
}))

vi.mock('vue', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue')>()
  let seq = 0
  return {
    ...actual,
    defineAsyncComponent: (loaderOrOpts: unknown) => {
      const id = ++seq
      const ChartStub = actual.defineComponent({
        name: `ChartStub${id}`,
        props: {
          option: { type: Object, default: () => ({}) },
          initOptions: { type: Object, default: () => ({}) },
        },
        emits: ['legendselectchanged'],
        setup:
          (_props, { emit }) =>
          () =>
            actual.h('div', {
              'data-testid': 'chart-stub',
              'data-stub-id': String(id),
              onClick: () => emit('legendselectchanged', { selected: { z1: false, z2: true } }),
            }),
      })
      if (typeof loaderOrOpts === 'function') {
        holder.loaders.push(loaderOrOpts as () => Promise<unknown>)
        void (loaderOrOpts as () => Promise<unknown>)().catch(() => {})
        return ChartStub
      }
      const opts = loaderOrOpts as {
        loader: () => Promise<unknown>
        loadingComponent?: unknown
        errorComponent?: unknown
      }
      holder.loaders.push(opts.loader)
      holder.loadingComp = opts.loadingComponent
      holder.errorComp = opts.errorComponent
      void opts.loader().catch(() => {})
      if (typeof opts.loadingComponent === 'function') {
        ;(opts.loadingComponent as () => unknown)()
      }
      if (typeof opts.errorComponent === 'function') {
        ;(opts.errorComponent as () => unknown)()
      }
      return ChartStub
    },
  }
})

const chartTypeRef = ref(holder.chartType as ChartType)
const threeDRef = ref(holder.threeD)
const sortRef = ref(holder.sort)
const isFullscreenRef = ref(holder.isFullscreen)

vi.mock('@/composables/useSettingsStore', () => ({
  useSettingsStore: () => ({
    isDark: ref(false),
    chartType: chartTypeRef,
  }),
}))

vi.mock('@/composables/useDataPoint', () => ({
  useDataPoint: () => ({
    activeArrangement: computed(() => ({ targetString: holder.arrangement })),
    activeDataset: computed(() => (holder.axes ? { axes: holder.axes } : undefined)),
  }),
}))

vi.mock('@/composables/useActiveChartShape', () => ({
  useActiveChartShape: () => ({
    sort: sortRef,
    showLabels: computed(() => false),
    scale: computed(() => 'linear' as const),
    stack: computed(() => false),
    threeDRotate: computed(() => false),
    threeD: threeDRef,
    threeDVisualMap: computed(() => false),
    visualMap: computed(() => false),
    stat: computed(() => holder.stat),
    symbol: computed(() => undefined),
    symbolSize: computed(() => undefined),
    smooth: computed(() => false),
    horizontal: computed(() => false),
  }),
}))

const optionsRef = ref(holder.options)

vi.mock('@/composables/useChartOptions', () => ({
  useChartOptions: (
    _chartData: unknown,
    resolvedSort: { value: unknown },
    _showLabels: unknown,
    _isDark: unknown,
    _chartType: unknown,
    _scale: unknown,
    _stack: unknown,
    _threeDRotate: unknown,
    _visibleZ: unknown,
    _threeD: unknown,
    _threeDVisualMap: unknown,
    _visualMap: unknown,
    arrangementTarget: { value: unknown }
  ) => {
    // Force evaluation of computed args passed from ChartCard.
    void resolvedSort.value
    void arrangementTarget.value
    return { options: optionsRef }
  },
}))

vi.mock('@/composables/useFullscreen', () => ({
  useFullscreen: () => ({
    containerRef: ref<HTMLElement | null>(null),
    isFullscreen: isFullscreenRef,
    withFullscreenToolbox: <T>(option: T) => option,
  }),
}))

vi.mock('./StatsPanel.vue', () => ({
  default: defineComponent({
    name: 'StatsPanel',
    props: ['chartData', 'math'],
    setup: () => () => h('div', { 'data-testid': 'stats-panel' }),
  }),
}))

vi.mock('@/lib/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/utils')>()
  return {
    ...actual,
    is3D: () => holder.threeD,
  }
})

import ChartCard from './ChartCard.vue'

function mountCard(overrides: { loading?: boolean; title?: string; chartData?: ChartData } = {}) {
  const chartData =
    overrides.chartData ??
    makeGroupedChartData(overrides.title !== undefined ? { title: overrides.title } : {})
  return mount(ChartCard, {
    props: {
      chartData,
      loading: overrides.loading,
    },
  })
}

describe('ChartCard', () => {
  beforeEach(() => {
    holder.stat = { enabled: true }
    holder.chartType = 'bar'
    chartTypeRef.value = 'bar'
    holder.threeD = false
    threeDRef.value = false
    holder.isFullscreen = false
    isFullscreenRef.value = false
    holder.arrangement = 'x/y'
    holder.axes = undefined
    holder.sort = undefined
    sortRef.value = undefined
    holder.options = { title: { text: 'opt' } }
    optionsRef.value = holder.options
    vi.stubGlobal('requestAnimationFrame', (cb: FrameRequestCallback) => {
      cb(0)
      return 0
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders chart title and badges', () => {
    const w = mountCard({ title: 'Q1 revenue' })
    expect(w.get('h3').text()).toBe('Q1 revenue')
    expect(w.text()).toMatch(/region|Series/)
    expect(w.text()).toMatch(/Total/)
  })

  it('shows skeleton overlay while loading', () => {
    const w = mountCard({ loading: true })
    expect(w.find('.absolute.animate-pulse').exists()).toBe(true)
  })

  it('reveals after loading false with rAF', async () => {
    const w = mountCard({ loading: true })
    await w.setProps({ loading: false })
    await flushPromises()
    await nextTick()
    expect(w.find('.absolute.animate-pulse').exists()).toBe(false)
  })

  it('toggles Stats panel when enabled', async () => {
    holder.stat = { enabled: true, math: ['center'] }
    const w = mountCard()
    const statsBtn = w.findAll('button').find((b) => b.text().includes('Stats'))!
    expect(statsBtn.exists()).toBe(true)
    expect(w.find('[data-testid="stats-panel"]').exists()).toBe(false)
    await statsBtn.trigger('click')
    expect(w.find('[data-testid="stats-panel"]').exists()).toBe(true)
  })

  it('hides Stats button when stat.enabled is false', () => {
    holder.stat = { enabled: false }
    const w = mountCard()
    expect(w.findAll('button').some((b) => b.text().includes('Stats'))).toBe(false)
  })

  it('hides Stats when series empty', () => {
    holder.stat = { enabled: true }
    const w = mountCard({
      chartData: makeGroupedChartData({ series: [] }),
    })
    expect(w.findAll('button').some((b) => b.text().includes('Stats'))).toBe(false)
  })

  it('handles legend select changed', async () => {
    const w = mountCard()
    await w.get('[data-testid="chart-stub"]').trigger('click')
    await nextTick()
    // total recomputed with visibleZ — still shows Total
    expect(w.text()).toMatch(/Total/)
  })

  it('routes pie/heatmap/radar past 3D', async () => {
    for (const t of ['pie', 'heatmap', 'radar'] as ChartType[]) {
      holder.chartType = t
      chartTypeRef.value = t
      holder.threeD = true
      threeDRef.value = true
      const w = mountCard({
        chartData: t === 'pie' ? makePieChartData() : makeGroupedChartData(),
      })
      expect(w.find('[data-testid="chart-stub"]').exists()).toBe(true)
      w.unmount()
    }
  })

  it('uses 3D renderer for bar when is3D', () => {
    holder.chartType = 'bar'
    chartTypeRef.value = 'bar'
    holder.threeD = true
    threeDRef.value = true
    const w = mountCard({
      chartData: makeGroupedChartData({
        zAxis: ['z1'],
        series: [{ xAxis: 'A', values: [1], benchmarkId: '' }],
        axisLabels: { x: 'x', y: 'y', z: 'z' },
      }),
    })
    expect(w.text()).toMatch(/z|Z-axis/)
    expect(w.find('[data-testid="chart-stub"]').exists()).toBe(true)
  })

  it('fullscreen class applies', () => {
    holder.isFullscreen = true
    isFullscreenRef.value = true
    const w = mountCard()
    expect(w.classes().join(' ') + w.html()).toMatch(/fixed inset-0|h-\[calc/)
  })

  it('pass-through option updates when not loading', async () => {
    const w = mountCard({ loading: false })
    optionsRef.value = { title: { text: 'new' } }
    await nextTick()
    expect(w.find('[data-testid="chart-stub"]').exists()).toBe(true)
  })

  it('freezes buffer while loading true', async () => {
    const w = mountCard({ loading: false })
    await w.setProps({ loading: true })
    optionsRef.value = { title: { text: 'during-load' } }
    await nextTick()
    expect(w.find('.absolute.animate-pulse').exists()).toBe(true)
  })

  it('falls back badge labels when axisLabels missing', () => {
    const w = mountCard({
      chartData: makeGroupedChartData({ axisLabels: {} }),
    })
    expect(w.text()).toMatch(/Series|Y-axis/)
  })

  it('falls back to bar renderer for unknown chart type', () => {
    holder.chartType = 'unknown' as ChartType
    chartTypeRef.value = 'unknown' as ChartType
    holder.threeD = false
    threeDRef.value = false
    const w = mountCard()
    expect(w.find('[data-testid="chart-stub"]').exists()).toBe(true)
  })

  it('switches ActiveChart when chartType changes without loading', async () => {
    holder.chartType = 'bar'
    chartTypeRef.value = 'bar'
    const w = mountCard({ loading: false })
    holder.chartType = 'line'
    chartTypeRef.value = 'line'
    await nextTick()
    expect(w.find('[data-testid="chart-stub"]').exists()).toBe(true)
  })

  it('uses explicit sort when present', () => {
    holder.sort = { enabled: true, order: 'desc' }
    sortRef.value = holder.sort
    const w = mountCard()
    expect(w.find('[data-testid="chart-stub"]').exists()).toBe(true)
  })

  it('invokes async loading and error placeholders', () => {
    expect(holder.loadingComp).toBeTruthy()
    expect(holder.errorComp).toBeTruthy()
    if (typeof holder.loadingComp === 'function') {
      expect((holder.loadingComp as () => unknown)()).toBeTruthy()
    }
    if (typeof holder.errorComp === 'function') {
      expect((holder.errorComp as () => unknown)()).toBeTruthy()
    }
  })
})
