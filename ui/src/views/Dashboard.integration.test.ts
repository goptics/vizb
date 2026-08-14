import { describe, it, expect, vi, beforeEach } from 'vitest'
import { computed, defineComponent, h, nextTick, ref, type Ref } from 'vue'
import { mount } from '@vue/test-utils'
import { makeGroupedChartData } from '@/test-utils'
import type { Axis, ChartData, Dataset } from '@/types'

const holder = vi.hoisted(() => ({
  datasets: undefined as Ref<{ name: string }[]> | undefined,
  activeDataset: undefined as Ref<Dataset | undefined> | undefined,
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
  // Author theme count gates selector visibility (0–1 hidden, 2+ shown).
  authorThemeCount: 0,
  themeNames: ['default'] as string[],
  sort: { enabled: false, order: 'asc' as const } as
    | { enabled: boolean; order: 'asc' | 'desc' }
    | undefined,
  showLabels: false,
  scale: 'linear' as const,
  threeD: false,
  charts: [] as { key: string; data?: ChartData; pending: boolean }[],
  groupNames: ['g0', 'g1'] as string[],
  datasetInitializing: false,
  initFromUrl: vi.fn(async () => {}),
  lastPipelineLabels: undefined as unknown,
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
    get datasets() {
      if (!holder.datasets) throw new Error('forgot beforeEach datasets')
      return holder.datasets
    },
    get activeDataset() {
      if (!holder.activeDataset) throw new Error('forgot beforeEach activeDataset')
      return holder.activeDataset
    },
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
    const labels = args[2]
    if (labels && typeof labels === 'object' && labels !== null && 'value' in labels) {
      holder.lastPipelineLabels = (labels as { value: unknown }).value
    }
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

vi.mock('../composables/useUrlRouter', () => ({
  useUrlRouter: () => ({
    initFromUrl: holder.initFromUrl,
  }),
}))

vi.mock('../lib/themes', () => ({
  listAvailableThemeNames: () => holder.themeNames,
  shouldShowThemeSelector: () => holder.authorThemeCount >= 2,
  authorThemeCount: () => holder.authorThemeCount,
  isThemeName: (v?: string) =>
    !!v && (v === 'default' || holder.themeNames.some((n) => n.toLowerCase() === v.toLowerCase())),
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
    // Declare props so :dataset does not fall through onto a native <div>
    // (HTMLElement.dataset is read-only → Vue warn under non-production Vue).
    props: ['dataset', 'datasets', 'activeDatasetId', 'resultGroups', 'activeGroupId'],
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
    holder.datasets = ref([{ name: 'Main' }, { name: 'Other' }])
    holder.activeDataset = ref(baseDataset())
    holder.loading = false
    holder.loadError = null
    holder.detailError = null
    holder.detailLoading = false
    holder.lazyCatalog = false
    holder.datasetInitializing = false
    datasetInitializingRef.value = false
    holder.isDark = false
    holder.themeName = 'default'
    // Default fixture: no author multi-theme → selector hidden.
    holder.authorThemeCount = 0
    holder.themeNames = ['default']
    holder.sort = { enabled: false, order: 'asc' }
    chartsRef.value = [
      { key: 'c1', data: makeGroupedChartData({ title: 'Chart One' }), pending: false },
      { key: 'c2', data: undefined, pending: true },
    ]
    groupNamesRef.value = ['g0', 'g1']
    document.title = 'Vizb'
    holder.initFromUrl.mockReset()
    holder.initFromUrl.mockResolvedValue(undefined)
    holder.lastPipelineLabels = undefined
    vi.stubGlobal('VIZB_VERSION', 'v1.2.3-test')
    vi.clearAllMocks()
  })

  it('main happy path: header, charts, footer, nav', async () => {
    const w = mount(Dashboard)
    expect(holder.initFromUrl).toHaveBeenCalledTimes(1)
    expect(document.title).toBe('Vizb | Main')
    expect(w.find('[data-testid="dataset-header"]').exists()).toBe(true)
    expect(w.find('[data-testid="chart-card"]').text()).toContain('Chart One')
    expect(w.find('[data-testid="footer"]').text()).toContain('v1.2.3-test')
    expect(w.find('[aria-label="View Package Source"]').attributes('href')).toContain(
      'github.com/org/pkg'
    )
    expect(w.find('[data-testid="settings-pop"]').exists()).toBe(true)
    // 0 author themes → theme selector hidden.
    expect(w.find('[data-testid="theme-selector"]').exists()).toBe(false)

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

  it('passes undefined chart labels when axes are missing or unlabeled', () => {
    holder.activeDataset = ref(baseDataset({ axes: undefined }))
    mount(Dashboard)
    expect(holder.lastPipelineLabels).toBeUndefined()

    holder.activeDataset = ref(
      baseDataset({
        axes: [
          { key: 'x', label: '' },
          { key: 'y', label: '' },
        ],
      })
    )
    mount(Dashboard)
    expect(holder.lastPipelineLabels).toBeUndefined()
  })

  it('shows theme selector only when author provided 2+ themes', async () => {
    holder.authorThemeCount = 2
    holder.themeNames = ['default', 'westeros', 'vintage']
    holder.themeName = 'westeros'
    const w = mount(Dashboard)
    expect(w.find('[data-testid="theme-selector"]').exists()).toBe(true)

    await w.get('[data-testid="theme-selector"]').trigger('click')
    expect(holder.setTheme).toHaveBeenCalledWith('vintage')
  })

  it('hides theme selector when author provided a single theme', () => {
    holder.authorThemeCount = 1
    holder.themeNames = ['westeros']
    holder.themeName = 'westeros'
    const w = mount(Dashboard)
    expect(w.find('[data-testid="theme-selector"]').exists()).toBe(false)
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
    holder.activeDataset!.value = undefined
    const w = mount(Dashboard)
    expect(w.find('main').exists()).toBe(false)
    expect(w.find('[data-testid="footer"]').exists()).toBe(false)
  })

  it('dark mode toggle button exists', () => {
    holder.isDark = true
    const w = mount(Dashboard)
    expect(w.find('[aria-label="Toggle dark mode"]').exists()).toBe(true)
  })

  it('custom theme name path shows selector when multi-theme report is active', () => {
    holder.authorThemeCount = 2
    holder.themeNames = ['default', 'westeros', 'vintage']
    holder.themeName = 'custom-hex'
    const w = mount(Dashboard)
    expect(w.find('[data-testid="theme-selector"]').exists()).toBe(true)
  })

  it('hides package button without pkg meta', () => {
    holder.activeDataset!.value = baseDataset({ meta: {} })
    const w = mount(Dashboard)
    expect(w.find('[aria-label="View Package Source"]').exists()).toBe(false)
  })

  it('uses default sort when shape sort undefined', () => {
    holder.sort = undefined
    const w = mount(Dashboard)
    expect(w.find('[data-testid="chart-card"]').exists()).toBe(true)
  })

  it('axes without labels', () => {
    holder.activeDataset!.value = baseDataset({
      axes: [{ key: 'x' } as Axis, { key: 'y', label: '' } as Axis],
    })
    const w = mount(Dashboard)
    expect(w.find('[data-testid="chart-card"]').exists()).toBe(true)
  })

  it('empty axes and preserveRows', () => {
    holder.activeDataset!.value = baseDataset({
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

  it('page shell keeps footer after content', () => {
    chartsRef.value = [
      { key: 'c1', data: makeGroupedChartData({ title: 'Chart One' }), pending: false },
    ]
    const w = mount(Dashboard)
    expect(w.find('[data-testid="page-shell"]').exists()).toBe(true)
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

  it('calls initFromUrl once when datasets become non-empty', async () => {
    holder.datasets!.value = []
    holder.activeDataset!.value = undefined
    mount(Dashboard)
    expect(holder.initFromUrl).not.toHaveBeenCalled()

    holder.datasets!.value = [{ name: 'Main' }]
    await nextTick()
    expect(holder.initFromUrl).toHaveBeenCalledTimes(1)

    holder.datasets!.value = [{ name: 'Main' }, { name: 'Other' }]
    await nextTick()
    expect(holder.initFromUrl).toHaveBeenCalledTimes(1)
  })

  it('sets document.title from activeDataset name', async () => {
    holder.activeDataset!.value = baseDataset({ name: 'Sales' })
    mount(Dashboard)
    await nextTick()
    expect(document.title).toBe('Vizb | Sales')
  })
})
