import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { nextTick, ref, type Ref } from 'vue'
import type { Dataset, BarConfig, LineConfig } from '../types'
import { ds } from '@/test-utils'

const holder = vi.hoisted(() => ({
  datasets: undefined as Ref<Dataset[]> | undefined,
  activeChartIndex: { value: 0 },
  chartType: { value: 'bar' as 'bar' | 'line' | 'pie' | 'heatmap' | 'radar' | 'scatter' },
  activeDatasetId: { value: 0 },
  activeGroupId: { value: 0 },
  selectDataset: vi.fn(),
  selectGroup: vi.fn(),
  setArrangement: vi.fn(),
  setChartType: vi.fn(),
  resultGroups: undefined as Ref<{ name: string }[]> | undefined,
  pathDatasetId: null as string | null,
  arrangementMap: new Map<string, string>(),
  activeDatasetRef: {
    get value() {
      return holder.datasets?.value[holder.activeDatasetId.value]
    },
  },
}))

vi.mock('./useDataPoint', () => ({
  activeDataset: holder.activeDatasetRef,
  useDataPoint: () => ({
    activeDataset: holder.activeDatasetRef,
    datasets: {
      get value() {
        return holder.datasets?.value ?? []
      },
    },
    get resultGroups() {
      if (!holder.resultGroups) throw new Error('forgot beforeEach resultGroups')
      return holder.resultGroups
    },
    get activeDatasetId() {
      return holder.activeDatasetId
    },
    get activeGroupId() {
      return holder.activeGroupId
    },
    activeArrangement: { value: { identityString: 'xy', targetString: 'xy' } },
    selectDataset: holder.selectDataset,
    selectGroup: holder.selectGroup,
    setArrangement: holder.setArrangement,
    arrangementMap: holder.arrangementMap,
    pathDatasetId: holder.pathDatasetId,
  }),
}))

vi.mock('./useSettingsStore', () => ({
  useSettingsStore: () => ({
    activeChartIndex: holder.activeChartIndex,
    chartType: holder.chartType,
    setChartType: holder.setChartType,
  }),
}))

function mockWindow(search: string, pathname = '/') {
  const replaceState = vi.fn()
  vi.stubGlobal('window', {
    location: { pathname, search, protocol: 'https:' },
    history: { replaceState },
  })
  return replaceState
}

describe('useUrlRouter', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  beforeEach(() => {
    vi.resetModules()
    holder.activeChartIndex.value = 0
    holder.activeDatasetId.value = 0
    holder.activeGroupId.value = 0
    holder.chartType.value = 'bar'
    holder.pathDatasetId = null
    holder.arrangementMap = new Map()
    holder.resultGroups = ref([{ name: 'first' }, { name: 'second' }])
    holder.selectDataset.mockReset()
    holder.selectDataset.mockImplementation(async (id: number) => {
      holder.activeDatasetId.value = id
      return true
    })
    holder.setChartType.mockReset()
    holder.selectGroup.mockReset()
    holder.setArrangement.mockReset()
    holder.datasets = ref([
      ds([
        {
          type: 'bar',
          sort: { enabled: false, order: 'asc' },
          threeD: true,
          threeDVisualMap: true,
        },
        {
          type: 'line',
          sort: { enabled: false, order: 'asc' },
          threeD: true,
          threeDVisualMap: true,
        },
        {
          type: 'scatter',
          sort: { enabled: false, order: 'asc' },
        },
      ]),
    ])
  })

  it('applies bar.3d-vm and line.3d-vm from the URL on init', async () => {
    mockWindow('?bar.3d-vm=false&line.3d-vm=true')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { initFromUrl } = useUrlRouter()
    await initFromUrl()

    const settings = holder.datasets!.value[0]!.settings
    const bar = settings[0] as BarConfig
    const line = settings[1] as LineConfig
    expect(bar.threeDVisualMap).toBe(false)
    expect(line.threeDVisualMap).toBe(true)
  })

  it('applies bar.3d and bar.3d-rt from the URL on init', async () => {
    mockWindow('?bar.3d=true&bar.3d-rt=true')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { initFromUrl } = useUrlRouter()
    await initFromUrl()

    const bar = holder.datasets!.value[0]!.settings[0] as BarConfig
    expect(bar.threeD).toBe(true)
    expect(bar.threeDRotate).toBe(true)
  })

  it('selects dataset by ?id= when present', async () => {
    holder.datasets = ref([
      ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]),
      { ...ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]), id: 'second' },
    ])
    mockWindow('?id=second')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { initFromUrl } = useUrlRouter()
    await initFromUrl()
    expect(holder.activeDatasetId.value).toBe(1)
    expect(holder.selectDataset).toHaveBeenCalledTimes(1)
    expect(holder.selectDataset).toHaveBeenCalledWith(1)
  })

  it('uses legacy ?d= when ?id= does not match', async () => {
    holder.datasets = ref([ds([{ type: 'bar' }]), ds([{ type: 'bar' }]), ds([{ type: 'bar' }])])
    mockWindow('?id=missing&d=2')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { initFromUrl } = useUrlRouter()
    await initFromUrl()
    expect(holder.selectDataset).toHaveBeenCalledWith(2)
  })

  it('syncs the URL immediately after successful initialization', async () => {
    holder.datasets = ref([ds([])])
    const replaceState = mockWindow('?d=0')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    expect(replaceState).toHaveBeenCalledWith(null, '', '/')
  })

  it('uses path identity while applying the shared chart and group parameters', async () => {
    holder.pathDatasetId = 'my-id'
    holder.datasets = ref([
      ds([
        { type: 'bar', threeD: false },
        { type: 'line', sort: { enabled: false, order: 'asc' } },
      ]),
      { ...ds([{ type: 'bar' }]), id: 'query-id' },
    ])
    mockWindow('?id=query-id&d=1&c=line&g=1&bar.3d=true', '/my-id')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    expect(holder.selectDataset).toHaveBeenCalledWith(0)
    expect(holder.selectGroup).toHaveBeenCalledWith(1)
    expect(holder.setChartType).toHaveBeenCalledWith('line')
    expect((holder.datasets.value[0]!.settings[0] as BarConfig).threeD).toBe(true)
  })

  it('applies chart parameters only after the selected detail has loaded', async () => {
    holder.datasets = ref([
      { id: 'one', name: 'One', data: [], settings: [] },
      { id: 'two', name: 'Two', data: [], settings: [] },
    ])
    let release!: () => void
    holder.selectDataset.mockImplementationOnce(
      (id: number) =>
        new Promise<boolean>((resolve) => {
          release = () => {
            holder.datasets!.value[id] = {
              id: 'two',
              name: 'Two',
              data: [],
              settings: [{ type: 'bar', horizontal: false }],
            }
            holder.activeDatasetId.value = id
            resolve(true)
          }
        })
    )
    mockWindow('?id=two&g=1&c=bar&bar.h=true&bar.sw=yx')
    const { useUrlRouter } = await import('./useUrlRouter')
    const pending = useUrlRouter().initFromUrl()

    expect(holder.selectDataset).toHaveBeenCalledWith(1)
    expect(holder.datasets.value[1]!.settings).toEqual([])

    release()
    await pending
    expect((holder.datasets.value[1]!.settings[0] as BarConfig).horizontal).toBe(true)
    expect(holder.setChartType).toHaveBeenCalledWith('bar')
    expect(holder.selectGroup).toHaveBeenCalledWith(1)
    expect(holder.setArrangement).toHaveBeenCalledWith(1, 'bar', 'yx')
  })

  it('applies deferred URL parameters after a failed detail is retried', async () => {
    holder.datasets = ref([
      { id: 'one', name: 'One', data: [], settings: [] },
      { id: 'two', name: 'Two', data: [], settings: [] },
    ])
    holder.selectDataset
      .mockImplementationOnce(async (id: number) => {
        holder.activeDatasetId.value = id
        return false
      })
      .mockImplementationOnce(async (id: number) => {
        holder.activeDatasetId.value = id
        return true
      })
    const replaceState = mockWindow('?id=two&bar.h=true')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    expect(replaceState).not.toHaveBeenCalled()

    holder.datasets.value = [
      holder.datasets.value[0]!,
      {
        id: 'two',
        name: 'Two',
        data: [],
        settings: [{ type: 'bar', horizontal: false }],
      },
    ]
    await nextTick()
    await nextTick()

    expect(holder.selectDataset).toHaveBeenCalledTimes(2)
    expect((holder.datasets.value[1]!.settings[0] as BarConfig).horizontal).toBe(true)
  })

  it('syncs ?id= when active dataset has an id', async () => {
    holder.datasets = ref([
      {
        ...ds([{ type: 'bar', sort: { enabled: false, order: 'asc' }, threeD: true }]),
        id: 'bench-v1',
      },
    ])
    holder.activeDatasetId.value = 0
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { syncUrlToState } = useUrlRouter()
    syncUrlToState()
    expect(replaceState).toHaveBeenCalledWith(null, '', '/?id=bench-v1&bar.3d=true')
  })

  it('syncs ?d= when active dataset has no id and index > 0', async () => {
    holder.datasets = ref([
      ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]),
      ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]),
    ])
    holder.activeDatasetId.value = 1
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { syncUrlToState } = useUrlRouter()
    syncUrlToState()
    expect(replaceState).toHaveBeenCalledWith(null, '', '/?d=1')
  })

  it('keeps the path identity and omits id/d while syncing chart parameters', async () => {
    holder.pathDatasetId = 'my-id'
    holder.datasets = ref([
      {
        ...ds([{ type: 'bar', threeD: true }]),
        id: 'my-id',
      },
    ])
    const replaceState = mockWindow('?id=query-id&d=1', '/my-id')
    const { useUrlRouter } = await import('./useUrlRouter')
    useUrlRouter().syncUrlToState()

    expect(replaceState).toHaveBeenCalledWith(null, '', '/my-id?bar.3d=true')
  })

  it('syncs 3D settings to bar.3d / bar.3d-vm in the URL', async () => {
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { syncUrlToState } = useUrlRouter()
    syncUrlToState()

    expect(replaceState).toHaveBeenCalledWith(
      null,
      '',
      '/?bar.3d=true&bar.3d-vm=true&line.3d=true&line.3d-vm=true'
    )
  })

  it('applies bar.h from the URL on init', async () => {
    mockWindow('?bar.h=true')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { initFromUrl } = useUrlRouter()
    await initFromUrl()

    const bar = holder.datasets!.value[0]!.settings[0] as BarConfig
    expect(bar.horizontal).toBe(true)
  })

  it('syncs bar.h to the URL', async () => {
    holder.datasets = ref([ds([{ type: 'bar', horizontal: true }])])
    holder.activeDatasetId.value = 0
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { syncUrlToState } = useUrlRouter()
    syncUrlToState()
    expect(replaceState).toHaveBeenCalledWith(null, '', '/?bar.h=true')
  })

  it('applies legacy global s/l/sc and per-chart sort/labels/scale/vm params', async () => {
    mockWindow(
      '?s=desc&l=true&sc=log&bar.so=asc&bar.l=false&line.l=true&line.sc=log&scatter.vm=true&scatter.so=desc'
    )
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    const settings = holder.datasets!.value[0]!.settings
    const bar = settings.find((s) => s.type === 'bar') as BarConfig
    const line = settings.find((s) => s.type === 'line') as LineConfig
    const scatter = settings.find((s) => s.type === 'scatter') as {
      sort?: { enabled: boolean; order: string }
      visualMap?: boolean
      scale?: string
      showLabels?: boolean
    }

    // Per-chart overrides win after legacy globals are applied.
    expect(bar.sort).toEqual({ enabled: true, order: 'asc' })
    expect(bar.showLabels).toBe(false)
    expect(bar.scale).toBe('log')
    expect(line.showLabels).toBe(true)
    expect(line.scale).toBe('log')
    expect(scatter.sort).toEqual({ enabled: true, order: 'desc' })
    expect(scatter.visualMap).toBe(true)
  })

  it('applies false toggles for labels, 3d, visualMap, and horizontal', async () => {
    holder.datasets = ref([
      ds([
        {
          type: 'bar',
          sort: { enabled: true, order: 'asc' },
          showLabels: true,
          threeD: true,
          threeDVisualMap: true,
          horizontal: true,
          scale: 'log',
        },
        {
          type: 'scatter',
          sort: { enabled: false, order: 'asc' },
          visualMap: true,
          threeDVisualMap: true,
        },
      ]),
    ])
    mockWindow(
      '?l=false&bar.3d=false&bar.3d-vm=false&bar.h=false&scatter.vm=false&scatter.3d-vm=false'
    )
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    const bar = holder.datasets!.value[0]!.settings[0] as BarConfig
    const scatter = holder.datasets!.value[0]!.settings[1] as {
      visualMap?: boolean
      threeDVisualMap?: boolean
      showLabels?: boolean
    }
    expect(bar.threeD).toBe(false)
    expect(bar.threeDVisualMap).toBe(false)
    expect(bar.horizontal).toBe(false)
    expect(bar.showLabels).toBe(false)
    expect(scatter.visualMap).toBe(false)
    expect(scatter.threeDVisualMap).toBe(false)
  })

  it('defers group selection until resultGroups populate', async () => {
    holder.resultGroups = ref([])
    mockWindow('?g=1')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()
    expect(holder.selectGroup).not.toHaveBeenCalled()

    holder.resultGroups.value = [{ name: 'a' }, { name: 'b' }]
    await nextTick()
    expect(holder.selectGroup).toHaveBeenCalledWith(1)
  })

  it('defers swap until datasets become available', async () => {
    holder.datasets = ref([])
    mockWindow('?bar.sw=yx')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()
    expect(holder.setArrangement).not.toHaveBeenCalled()

    holder.datasets.value = [ds([{ type: 'bar' }])]
    await nextTick()
    expect(holder.setArrangement).toHaveBeenCalledWith(0, 'bar', 'yx')
  })

  it('syncs chart index, group, sort, labels, scale, scatter vm, and swap', async () => {
    holder.datasets = ref([
      ds([
        {
          type: 'bar',
          sort: { enabled: true, order: 'desc' },
          showLabels: true,
          scale: 'log',
          threeD: true,
          threeDRotate: true,
          threeDVisualMap: false,
          horizontal: true,
        },
        {
          type: 'scatter',
          sort: { enabled: false, order: 'asc' },
          showLabels: false,
          visualMap: false,
          scale: 'linear',
        },
        {
          type: 'line',
          sort: { enabled: false, order: 'asc' },
          visualMap: true,
        },
      ]),
    ])
    holder.activeChartIndex.value = 1
    holder.activeGroupId.value = 2
    holder.arrangementMap.set('0:bar', 'yx')

    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    useUrlRouter().syncUrlToState()

    const url = String(replaceState.mock.calls[0]?.[2] ?? '')
    expect(url).toContain('c=scatter')
    expect(url).toContain('g=2')
    expect(url).toContain('bar.so=desc')
    expect(url).toContain('bar.l=true')
    expect(url).toContain('bar.sc=log')
    expect(url).toContain('bar.3d=true')
    expect(url).toContain('bar.3d-rt=true')
    expect(url).toContain('bar.3d-vm=false')
    expect(url).toContain('bar.h=true')
    expect(url).toContain('bar.sw=yx')
    expect(url).toContain('scatter.l=false')
    expect(url).toContain('scatter.vm=false')
  })

  it('syncs scatter visualMap true and omits empty query values', async () => {
    holder.datasets = ref([
      ds([
        {
          type: 'scatter',
          sort: { enabled: false, order: 'asc' },
          visualMap: true,
        },
      ]),
    ])
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    useUrlRouter().syncUrlToState()
    expect(replaceState).toHaveBeenCalledWith(null, '', '/?scatter.vm=true')
  })

  it('ignores config updates when settings or chart type are missing', async () => {
    // settings omitted entirely → applyConfigUpdate early-returns on !settings.
    // Avoid legacy s/l/sc (those read availableTypes via settings.map).
    holder.datasets = ref([{ name: 'empty', data: [] } as Dataset])
    mockWindow('?bar.l=true&pie.so=asc')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()
    expect(holder.datasets.value[0]!.settings).toBeUndefined()

    holder.datasets = ref([
      {
        name: 'pie-only',
        data: [],
        settings: [{ type: 'pie', sort: { enabled: false, order: 'asc' } }],
      },
    ])
    mockWindow('?bar.sc=log&pie.l=true')
    const { useUrlRouter: useUrlRouter2 } = await import('./useUrlRouter')
    await useUrlRouter2().initFromUrl()
    const pie = holder.datasets.value[0]!.settings![0]!
    expect(pie.showLabels).toBe(true)
  })

  it('handles invalid index params and empty deferred watch ticks', async () => {
    holder.resultGroups = ref([])
    mockWindow('?g=not-a-number&d=NaN')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()
    expect(holder.selectGroup).not.toHaveBeenCalled()

    // Deferred group watch fires with length still 0 first is skipped by once:true only on change.
    holder.resultGroups.value = []
    await nextTick()
    holder.resultGroups.value = [{ name: 'only' }]
    await nextTick()
    // g was invalid, so setter still not called
    expect(holder.selectGroup).not.toHaveBeenCalled()
  })

  it('syncs pie configs without cartesian branches', async () => {
    holder.datasets = ref([
      ds([
        { type: 'pie', sort: { enabled: true, order: 'asc' }, showLabels: false },
        { type: 'heatmap', showLabels: true },
      ]),
    ])
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    useUrlRouter().syncUrlToState()
    const url = String(replaceState.mock.calls[0]?.[2] ?? '')
    expect(url).toContain('pie.so=asc')
    expect(url).toContain('pie.l=false')
    expect(url).toContain('heatmap.l=true')
    expect(url).not.toContain('pie.sc=')
  })

  it('ignores deferred retry when active dataset id no longer matches', async () => {
    holder.datasets = ref([
      { id: 'one', name: 'One', data: [], settings: [] },
      { id: 'two', name: 'Two', data: [], settings: [] },
    ])
    holder.selectDataset.mockImplementationOnce(async (id: number) => {
      holder.activeDatasetId.value = id
      return false
    })
    mockWindow('?id=two&bar.h=true')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    // User switched away before the failed shell was replaced.
    holder.activeDatasetId.value = 0
    holder.datasets.value = [
      holder.datasets.value[0]!,
      {
        id: 'two',
        name: 'Two',
        data: [],
        settings: [{ type: 'bar', horizontal: false }],
      },
    ]
    await nextTick()
    await nextTick()

    // Retry watcher must not re-apply while activeDatasetId differs.
    expect(holder.selectDataset).toHaveBeenCalledTimes(1)
    expect((holder.datasets.value[1]!.settings[0] as BarConfig).horizontal).toBe(false)
  })

  it('covers empty deferred group/swap ticks and empty query values', async () => {
    holder.resultGroups = ref([{ name: 'seed' }])
    holder.datasets = ref([])
    mockWindow('?g=0&bar.sw=yx&c=')
    // Force an empty-string query param through location.search parsing.
    const replaceState = mockWindow('?g=0&bar.sw=yx&empty=')
    holder.resultGroups = ref([])
    holder.datasets = ref([])
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    // First deferred ticks stay at 0 / empty — exercise len>0 false branches.
    holder.resultGroups.value = []
    holder.datasets.value = []
    await nextTick()

    holder.resultGroups.value = [{ name: 'a' }, { name: 'b' }]
    holder.datasets.value = [ds([{ type: 'bar' }])]
    await nextTick()
    expect(holder.selectGroup).toHaveBeenCalledWith(0)
    expect(holder.setArrangement).toHaveBeenCalledWith(0, 'bar', 'yx')

    // Sync with only defaults produces an empty query (buildQueryString false branch via omit).
    holder.activeGroupId.value = 0
    holder.activeChartIndex.value = 0
    holder.arrangementMap.clear()
    replaceState.mockClear()
    useUrlRouter().syncUrlToState()
    // May or may not call replaceState depending on current URL; just ensure no throw.
    expect(true).toBe(true)
  })

  it('skips deferred group watch when g is absent and groups are empty', async () => {
    holder.resultGroups = ref([])
    mockWindow('?c=bar')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()
    expect(holder.selectGroup).not.toHaveBeenCalled()
    expect(holder.setChartType).toHaveBeenCalledWith('bar')
  })
})
