import { describe, it, expect, vi, beforeEach } from 'vitest'
import { computed, defineComponent, h, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { makeGroupedChartData } from '@/test-utils'
import ChartCard from './ChartCard.vue'

const holder = vi.hoisted(() => ({
  stat: { enabled: true } as { enabled: boolean } | undefined,
}))

// Keep chart renderers out of the tree: ChartCard wraps them in defineAsyncComponent.
// A sync stub avoids echarts + VTU probing mock module exports for Vue internals.
vi.mock('vue', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue')>()
  const ChartStub = actual.defineComponent({
    name: 'ChartStub',
    props: {
      option: { type: Object, default: () => ({}) },
      initOptions: { type: Object, default: () => ({}) },
    },
    emits: ['legendselectchanged'],
    setup: () => () => actual.h('div', { 'data-testid': 'chart-stub' }),
  })
  return {
    ...actual,
    defineAsyncComponent: () => ChartStub,
  }
})

vi.mock('@/composables/useSettingsStore', () => ({
  useSettingsStore: () => ({
    isDark: ref(false),
    chartType: ref('bar'),
  }),
}))

vi.mock('@/composables/useDataPoint', () => ({
  useDataPoint: () => ({
    activeArrangement: computed(() => ({ targetString: 'x/y' })),
    activeDataset: computed(() => undefined),
  }),
}))

vi.mock('@/composables/useActiveChartShape', () => ({
  useActiveChartShape: () => ({
    sort: computed(() => ({ enabled: false, order: 'asc' as const })),
    showLabels: computed(() => false),
    scale: computed(() => 'linear' as const),
    stack: computed(() => false),
    threeDRotate: computed(() => false),
    threeD: computed(() => false),
    threeDVisualMap: computed(() => false),
    visualMap: computed(() => false),
    stat: computed(() => holder.stat),
    symbol: computed(() => undefined),
    symbolSize: computed(() => undefined),
    smooth: computed(() => false),
    horizontal: computed(() => false),
  }),
}))

vi.mock('@/composables/useChartOptions', () => ({
  useChartOptions: () => ({ options: ref({}) }),
}))

vi.mock('@/composables/useFullscreen', () => ({
  useFullscreen: () => ({
    containerRef: ref<HTMLElement | null>(null),
    isFullscreen: ref(false),
    withFullscreenToolbox: <T>(option: T) => option,
  }),
}))

vi.mock('./StatsPanel.vue', () => ({
  default: defineComponent({
    name: 'StatsPanel',
    setup: () => () => h('div', { 'data-testid': 'stats-panel' }),
  }),
}))

function mountCard(overrides: { loading?: boolean; title?: string } = {}) {
  const chartData = makeGroupedChartData(
    overrides.title !== undefined ? { title: overrides.title } : {}
  )
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
  })

  it('renders chart title', () => {
    const w = mountCard({ title: 'Q1 revenue' })
    expect(w.get('h3').text()).toBe('Q1 revenue')
  })

  it('shows skeleton overlay while loading', () => {
    const w = mountCard({ loading: true })
    // Overlay on the chart area (not the async loadingComponent placeholder).
    const skeleton = w.find('.absolute.animate-pulse')
    expect(skeleton.exists()).toBe(true)
  })

  it('shows Stats button when stat.enabled is true', () => {
    holder.stat = { enabled: true }
    const w = mountCard()
    const buttons = w.findAll('button')
    expect(buttons.some((b) => b.text().includes('Stats'))).toBe(true)
  })

  it('hides Stats button when stat.enabled is false', () => {
    holder.stat = { enabled: false }
    const w = mountCard()
    const buttons = w.findAll('button')
    expect(buttons.some((b) => b.text().includes('Stats'))).toBe(false)
  })
})
