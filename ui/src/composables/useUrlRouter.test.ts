import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { ref, type Ref } from 'vue'
import type { DataSet, BarConfig, LineConfig } from '../types'

const holder = vi.hoisted(() => ({
  dataSets: undefined as Ref<DataSet[]> | undefined,
  activeChartIndex: { value: 0 },
  chartType: { value: 'bar' as 'bar' | 'line' | 'pie' | 'heatmap' | 'radar' },
  activeDataSetId: { value: 0 },
  activeDataSetRef: {
    get value() {
      return holder.dataSets?.value[holder.activeDataSetId.value]
    },
  },
}))

vi.mock('./useDataPoint', () => ({
  activeDataSet: holder.activeDataSetRef,
  useDataPoint: () => ({
    activeDataSet: holder.activeDataSetRef,
    dataSets: {
      get value() {
        return holder.dataSets?.value ?? []
      },
    },
    resultGroups: { value: [] },
    get activeDataSetId() {
      return holder.activeDataSetId
    },
    activeGroupId: { value: 0 },
    activeArrangement: { value: { identityString: 'xy', targetString: 'xy' } },
    selectDataSet: (id: number) => {
      holder.activeDataSetId.value = id
    },
    selectGroup: vi.fn(),
    setArrangement: vi.fn(),
    arrangementMap: new Map<string, string>(),
  }),
}))

vi.mock('./useSettingsStore', () => ({
  useSettingsStore: () => ({
    activeChartIndex: holder.activeChartIndex,
    chartType: holder.chartType,
    setChartType: vi.fn(),
  }),
}))

const ds = (settings: DataSet['settings']): DataSet => ({
  name: 'test',
  settings,
  data: [],
})

function mockWindow(search: string) {
  const replaceState = vi.fn()
  vi.stubGlobal('window', {
    location: { pathname: '/', search },
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
    holder.activeDataSetId.value = 0
    holder.chartType.value = 'bar'
    holder.dataSets = ref([
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
    initFromUrl()

    const settings = holder.dataSets!.value[0]!.settings
    const bar = settings[0] as BarConfig
    const line = settings[1] as LineConfig
    expect(bar.threeDVisualMap).toBe(false)
    expect(line.threeDVisualMap).toBe(true)
  })

  it('applies bar.3d and bar.3d-rt from the URL on init', async () => {
    mockWindow('?bar.3d=true&bar.3d-rt=true')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { initFromUrl } = useUrlRouter()
    initFromUrl()

    const bar = holder.dataSets!.value[0]!.settings[0] as BarConfig
    expect(bar.threeD).toBe(true)
    expect(bar.threeDRotate).toBe(true)
  })

  it('selects dataset by ?id= when present', async () => {
    holder.dataSets = ref([
      ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]),
      { ...ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]), id: 'second' },
    ])
    mockWindow('?id=second')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { initFromUrl } = useUrlRouter()
    initFromUrl()
    expect(holder.activeDataSetId.value).toBe(1)
  })

  it('syncs ?id= when active dataset has an id', async () => {
    holder.dataSets = ref([
      {
        ...ds([{ type: 'bar', sort: { enabled: false, order: 'asc' }, threeD: true }]),
        id: 'bench-v1',
      },
    ])
    holder.activeDataSetId.value = 0
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { syncUrlToState } = useUrlRouter()
    syncUrlToState()
    expect(replaceState).toHaveBeenCalledWith(null, '', '/?id=bench-v1&bar.3d=true')
  })

  it('syncs ?d= when active dataset has no id and index > 0', async () => {
    holder.dataSets = ref([
      ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]),
      ds([{ type: 'bar', sort: { enabled: false, order: 'asc' } }]),
    ])
    holder.activeDataSetId.value = 1
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { syncUrlToState } = useUrlRouter()
    syncUrlToState()
    expect(replaceState).toHaveBeenCalledWith(null, '', '/?d=1')
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
    initFromUrl()

    const bar = holder.dataSets!.value[0]!.settings[0] as BarConfig
    expect(bar.horizontal).toBe(true)
  })

  it('syncs bar.h to the URL', async () => {
    holder.dataSets = ref([ds([{ type: 'bar', horizontal: true }])])
    holder.activeDataSetId.value = 0
    const replaceState = mockWindow('')
    const { useUrlRouter } = await import('./useUrlRouter')
    const { syncUrlToState } = useUrlRouter()
    syncUrlToState()
    expect(replaceState).toHaveBeenCalledWith(null, '', '/?bar.h=true')
  })
})
