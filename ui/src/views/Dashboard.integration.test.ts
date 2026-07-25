import { describe, it, expect, vi, beforeEach } from 'vitest'
import { computed, defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { makeGroupedChartData } from '@/test-utils'
import type { Axis, ChartData, Dataset } from '@/types'

const holder = vi.hoisted(() => ({
  datasets: [{ name: 'Main' }, { name: 'Other' }],
  activeDataset: undefined as Dataset | undefined,
  activeDatasetId: 0,
  selectDataset: vi.fn(),
  activeArrangement: { identityString: 'x/y', targetString: 'x/y' },
  resultGroups: [{ name: 'g0' }, { name: 'g1' }],
  activeGroupId: 0,
  selectGroup: vi.fn(),
  setGroupNames: vi.fn(),
  loading: false,
  loadError: null as string | null,
  lazyCatalog: false,
  detailLoading: false,
  detailError: null as string | null,
  retryActiveDataset: vi.fn(),
  isDark: false,
  toggleDark: vi.fn(),
  chartType: 'bar' as const,
  themeName: 'default',
  setTheme: vi.fn(),
  sort: { enabled: false, order: 'asc' as const } as
    | { enabled: boolean; order: 'asc' | 'desc' }
    | undefined,
  showLabels: false,
  scale: 'linear' as const,
  threeD: false,
  charts: [] as { key: string; data?: ChartData; pending: boolean }[],
  groupNames: ['g0', 'g1'] as string[],
  datasetInitializing: false,
  dashInit: vi.fn(),
}))

function baseDataset(overrides: Partial<Dataset> = {}): Dataset {
  return {
    name: 'Main',
    description: 'd',
    meta: { pkg: 'github.com/org/pkg', cpu: { name: 'M', cores: 2 }, os: 'linux' },
    timestamp: '2024-01-01T00:00:00.000Z',
    settings: [
      {
        type: 'bar',
        sort: { enabled: false, order: 'asc' },
        scale: 'linear',
        stack: false,
        showLabels: false,
        swap: 'x/y',
      },
    ],
    data: [{ xAxis: 'a', yAxis: 'b', zAxis: '', stats: [{ type: 'v', value: 1 }] }],
    axes: [
      { key: 'x', label: 'X' },
      { key: 'y', label: 'Y' },
    ],
    ...overrides,
  }
}

const groupNamesRef = ref(holder.groupNames)
const chartsRef = ref(holder.charts)
const datasetInitializingRef = ref(holder.datasetInitializing)

vi.mock('../composables/useDataPoint', () => ({
  useDataPoint: () => ({
    datasets: computed(() => holder.datasets),
    activeDataset: computed(() => holder.activeDataset),
    activeDatasetId: computed(() => holder.activeDatasetId),
    selectDataset: holder.selectDataset,
    activeArrangement: computed(() => holder.activeArrangement),
    resultGroups: computed(() => holder.resultGroups),
    activeGroupId: computed(() => holder.activeGroupId),
    selectGroup: holder.selectGroup,
    setGroupNames: holder.setGroupNames,
    loading: computed(() => holder.loading),
    loadError: computed(() => holder.loadError),
    lazyCatalog: computed(() => holder.lazyCatalog),
    detailLoading: computed(() => holder.detailLoading),
    detailError: computed(() => holder.detailError),
    retryActiveDataset: holder.retryActiveDataset,
  }),
}))

vi.mock('../composables/useSettingsStore', () => ({
  useSettingsStore: () => ({
    isDark: computed(() => holder.isDark),
    toggleDark: holder.toggleDark,
    chartType: computed(() => holder.chartType),
    themeName: computed(() => holder.themeName),
    setTheme: holder.setTheme,
  }),
}))

vi.mock('../composables/useActiveChartShape', () => ({
  useActiveChartShape: () => ({
    sort: computed(() => holder.sort),
    showLabels: computed(() => holder.showLabels),
    scale: computed(() => holder.scale),
    threeD: computed(() => holder.threeD),
  }),
}))

vi.mock('../composables/useChartPipeline', () => ({
  useChartPipeline: (...args: unknown[]) => {
    for (const a of args) {
      if (a && typeof a === 'object' && a !== null && 'value' in a) {
        void (a as { value: unknown }).value
      }
    }
    return {
      charts: chartsRef,
      groupNames: groupNamesRef,
      datasetInitializing: datasetInitializingRef,
    }
  },
}))

vi.mock('../composables/useDashboardInit', () => ({
  useDashboardInit: () => holder.dashInit(),
}))

vi.mock('../lib/themes', () => ({
  THEME_NAMES: ['default', 'vintage'],
  isThemeName: (v?: string) => v === 'default' || v === 'vintage',
}))

vi.mock('../lib/swap', () => ({
  swapAxisLabels: (_i: string, _t: string, labels?: Record<string, string>) => labels ?? {},
}))

vi.mock('../components/ChartSettingsPopover.vue', () => ({
  default: defineComponent({
    name: 'ChartSettingsPopover',
    setup: () => () => h('div', { 'data-testid': 'settings-pop' }),
  }),
}))

vi.mock('../components/ChartCard.vue', () => ({
  default: defineComponent({
    name: 'ChartCard',
    props: ['chartData', 'loading'],
    setup: (p) => () => {
      const data = p.chartData as ChartData | undefined
      return h('div', { 'data-testid': 'chart-card' }, data?.title ?? '')
    },
  }),
}))

vi.mock('../components/DatasetHeader.vue', () => ({
  default: defineComponent({
    name: 'DatasetHeader',
    emits: ['selectDataset', 'selectGroup'],
    setup:
      (_, { emit }) =>
      () =>
        h('div', { 'data-testid': 'dataset-header' }, [
          h('button', { 'data-ds': '1', onClick: () => emit('selectDataset', 1) }, 'ds'),
          h('button', { 'data-g': '1', onClick: () => emit('selectGroup', 1) }, 'g'),
        ]),
  }),
}))

vi.mock('../components/LoadingSkeleton.vue', () => ({
  default: defineComponent({
    name: 'LoadingSkeleton',
    props: {
      contentOnly: { type: Boolean, default: false },
    },
    setup: (p) => () =>
      h('div', {
        'data-testid': p.contentOnly ? 'detail-skeleton' : 'page-skeleton',
      }),
  }),
}))

vi.mock('../components/LoadError.vue', () => ({
  default: defineComponent({
    name: 'LoadError',
    props: {
      message: String,
      inline: { type: Boolean, default: false },
      retry: Function,
    },
    setup: (p) => () =>
      h(
        'div',
        {
          'data-testid': p.inline ? 'detail-error' : 'page-error',
          onClick: () => {
            if (typeof p.retry === 'function') p.retry()
          },
        },
        String(p.message)
      ),
  }),
}))

vi.mock('../components/AppFooter.vue', () => ({
  default: defineComponent({
    name: 'AppFooter',
    props: ['version'],
    setup: (p) => () => h('footer', { 'data-testid': 'footer' }, String(p.version)),
  }),
}))

vi.mock('../components/IconButton.vue', () => ({
  default: defineComponent({
    name: 'IconButton',
    props: ['href'],
    setup:
      (p, { slots, attrs }) =>
      () =>
        h(
          p.href ? 'a' : 'button',
          {
            href: p.href as string | undefined,
            'aria-label': attrs['aria-label'] as string | undefined,
            onClick: attrs.onClick as (() => void) | undefined,
          },
          slots.default?.()
        ),
  }),
}))

vi.mock('../components/Selector.vue', () => ({
  default: defineComponent({
    name: 'Selector',
    props: ['items', 'activeValue'],
    emits: ['selectValue'],
    setup:
      (p, { emit }) =>
      () =>
        h(
          'button',
          {
            'data-testid': 'theme-selector',
            onClick: () => emit('selectValue', 'vintage'),
          },
          String(p.activeValue)
        ),
  }),
}))

import Dashboard from './Dashboard.vue'

describe('Dashboard', () => {
  beforeEach(() => {
    holder.activeDataset = baseDataset()
    holder.datasets = [{ name: 'Main' }, { name: 'Other' }]
    holder.loading = false
    holder.loadError = null
    holder.detailError = null
    holder.detailLoading = false
    holder.lazyCatalog = false
    holder.datasetInitializing = false
    datasetInitializingRef.value = false
    holder.isDark = false
    holder.themeName = 'default'
    holder.sort = { enabled: false, order: 'asc' }
    chartsRef.value = [
      { key: 'c1', data: makeGroupedChartData({ title: 'Chart One' }), pending: false },
      { key: 'c2', data: undefined, pending: true },
    ]
    groupNamesRef.value = ['g0', 'g1']
    vi.stubGlobal('VIZB_VERSION', 'v1.2.3-test')
    vi.clearAllMocks()
  })

  it('main happy path: header, charts, footer, nav', async () => {
    const w = mount(Dashboard)
    expect(holder.dashInit).toHaveBeenCalled()
    expect(w.find('[data-testid="dataset-header"]').exists()).toBe(true)
    expect(w.find('[data-testid="chart-card"]').text()).toContain('Chart One')
    expect(w.find('[data-testid="footer"]').text()).toContain('v1.2.3-test')
    expect(w.find('[aria-label="View Package Source"]').attributes('href')).toContain(
      'github.com/org/pkg'
    )
    expect(w.find('[data-testid="settings-pop"]').exists()).toBe(true)

    await w.get('[data-testid="theme-selector"]').trigger('click')
    expect(holder.setTheme).toHaveBeenCalledWith('vintage')

    await w.get('[aria-label="Toggle dark mode"]').trigger('click')
    expect(holder.toggleDark).toHaveBeenCalled()

    await w.get('[data-ds]').trigger('click')
    expect(holder.selectDataset).toHaveBeenCalledWith(1)
    await w.get('[data-g]').trigger('click')
    expect(holder.selectGroup).toHaveBeenCalledWith(1)

    groupNamesRef.value = ['a', 'b']
    await nextTick()
    expect(holder.setGroupNames).toHaveBeenCalledWith(['a', 'b'])
  })

  it('loadError branch', () => {
    holder.loadError = 'boom'
    const w = mount(Dashboard)
    expect(w.get('[data-testid="page-error"]').text()).toContain('boom')
    expect(w.find('[data-testid="footer"]').exists()).toBe(false)
  })

  it('page skeleton while loading', () => {
    holder.loading = true
    const w = mount(Dashboard)
    expect(w.find('[data-testid="page-skeleton"]').exists()).toBe(true)
  })

  it('detail error inline', async () => {
    holder.detailError = 'detail fail'
    const w = mount(Dashboard)
    expect(w.get('[data-testid="detail-error"]').text()).toContain('detail fail')
    await w.get('[data-testid="detail-error"]').trigger('click')
    expect(holder.retryActiveDataset).toHaveBeenCalled()
  })

  it('detail skeleton when lazy catalog loading', () => {
    holder.lazyCatalog = true
    holder.detailLoading = true
    const w = mount(Dashboard)
    expect(w.find('[data-testid="detail-skeleton"]').exists()).toBe(true)
  })

  it('detail skeleton when datasetInitializing', () => {
    holder.lazyCatalog = true
    datasetInitializingRef.value = true
    const w = mount(Dashboard)
    expect(w.find('[data-testid="detail-skeleton"]').exists()).toBe(true)
  })

  it('detail skeleton when data present but charts empty', () => {
    holder.lazyCatalog = true
    chartsRef.value = []
    const w = mount(Dashboard)
    expect(w.find('[data-testid="detail-skeleton"]').exists()).toBe(true)
  })

  it('no active dataset renders nav only', () => {
    holder.activeDataset = undefined
    const w = mount(Dashboard)
    expect(w.find('main').exists()).toBe(false)
    expect(w.find('[data-testid="footer"]').exists()).toBe(false)
  })

  it('dark mode toggle button exists', () => {
    holder.isDark = true
    const w = mount(Dashboard)
    expect(w.find('[aria-label="Toggle dark mode"]').exists()).toBe(true)
  })

  it('custom theme name path', () => {
    holder.themeName = 'custom-hex'
    const w = mount(Dashboard)
    expect(w.find('[data-testid="theme-selector"]').exists()).toBe(true)
  })

  it('hides package button without pkg meta', () => {
    holder.activeDataset = baseDataset({ meta: {} })
    const w = mount(Dashboard)
    expect(w.find('[aria-label="View Package Source"]').exists()).toBe(false)
  })

  it('uses default sort when shape sort undefined', () => {
    holder.sort = undefined
    const w = mount(Dashboard)
    expect(w.find('[data-testid="chart-card"]').exists()).toBe(true)
  })

  it('axes without labels', () => {
    holder.activeDataset = baseDataset({
      axes: [{ key: 'x' } as Axis, { key: 'y', label: '' } as Axis],
    })
    const w = mount(Dashboard)
    expect(w.find('[data-testid="chart-card"]').exists()).toBe(true)
  })

  it('empty axes and preserveRows', () => {
    holder.activeDataset = baseDataset({
      axes: undefined,
      data: [],
      preserveRows: true,
    })
    chartsRef.value = [{ key: 'c', data: makeGroupedChartData(), pending: true }]
    const w = mount(Dashboard)
    expect(w.find('[data-testid="chart-card"]').exists()).toBe(true)
  })
  it('footer still mounts with version', () => {
    const w = mount(Dashboard)
    expect(w.find('[data-testid="footer"]').exists()).toBe(true)
  })

  it('covers version fallback via resetModules', async () => {
    vi.stubGlobal('VIZB_VERSION', undefined)
    vi.resetModules()
    // Re-import after clearing module cache so setup re-reads VIZB_VERSION.
    // Dynamic import is intentional: static import is already evaluated.
    const mod = await import('./Dashboard.vue')
    const w = mount(mod.default)
    expect(w.find('[data-testid="footer"]').text()).toContain('v0.0.0-dev')
    w.unmount()
  })
})
