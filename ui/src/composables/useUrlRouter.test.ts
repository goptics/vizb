import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { nextTick, ref, type Ref } from 'vue'
import type { Dataset, BarConfig, LineConfig } from '../types'

const holder = vi.hoisted(() => ({
  datasets: undefined as Ref<Dataset[]> | undefined,
  activeChartIndex: { value: 0 },
  chartType: { value: 'bar' as 'bar' | 'line' | 'pie' | 'heatmap' | 'radar' },
  activeDatasetId: { value: 0 },
  selectDataset: vi.fn(),
  selectGroup: vi.fn(),
  setArrangement: vi.fn(),
  setChartType: vi.fn(),
  resultGroups: { value: [{ name: 'first' }, { name: 'second' }] },
  pathDatasetId: null as string | null,
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
    resultGroups: holder.resultGroups,
    get activeDatasetId() {
      return holder.activeDatasetId
    },
    activeGroupId: { value: 0 },
    activeArrangement: { value: { identityString: 'xy', targetString: 'xy' } },
    selectDataset: holder.selectDataset,
    selectGroup: holder.selectGroup,
    setArrangement: holder.setArrangement,
    arrangementMap: new Map<string, string>(),
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

const ds = (settings: Dataset['settings']): Dataset => ({
  name: 'test',
  settings,
  data: [],
})

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
    holder.chartType.value = 'bar'
    holder.pathDatasetId = null
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

  it('applies label mode after legacy label parameters', async () => {
    mockWindow('?l=true&bar.l=false&bar.lm=percentage')
    const { useUrlRouter } = await import('./useUrlRouter')
    await useUrlRouter().initFromUrl()

    const settings = holder.datasets!.value[0]!.settings
    expect(settings[0]).toMatchObject({ showLabels: false, labelMode: 'percentage' })
    expect(settings[1]).toMatchObject({ showLabels: true })
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

  it('syncs labelMode and keeps legacy-only configs on the old parameter', async () => {
    holder.datasets = ref([
      ds([
        { type: 'bar', showLabels: true, labelMode: 'percentage' },
        { type: 'line', showLabels: true },
      ]),
    ])
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    useUrlRouter().syncUrlToState()

    expect(replaceState).toHaveBeenCalledWith(null, '', '/?bar.lm=percentage&line.l=true')
  })
})
