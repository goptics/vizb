import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const detail = {
  name: 'Dataset 2',
  data: [],
  settings: [{ type: 'bar' as const }],
}

describe('useDataPoint remote loading', () => {
  beforeEach(() => vi.resetModules())
  afterEach(() => vi.unstubAllGlobals())

  it('loads a catalog entry on selection and reuses its detail', async () => {
    const catalog = [
      { id: 'dataset-1', name: 'Dataset 1' },
      { id: 'dataset/2', name: 'Dataset 2' },
    ]
    const selectedDetail = {
      id: 'dataset/2',
      name: 'Dataset 2',
      data: [{ name: 'value', value: 2 }],
      settings: [{ type: 'bar' as const }],
    }
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(catalog), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(selectedDetail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog?format=full',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.lazyCatalog.value).toBe(true)
    expect(state.datasets.value).toMatchObject([
      { id: 'dataset-1', name: 'Dataset 1', data: [], settings: [] },
      { id: 'dataset/2', name: 'Dataset 2', data: [], settings: [] },
    ])

    await expect(state.selectDataset(1)).resolves.toBe(true)
    expect(fetcher).toHaveBeenNthCalledWith(1, 'https://api.example.com/catalog?format=full')
    expect(fetcher).toHaveBeenNthCalledWith(
      2,
      'https://api.example.com/catalog/dataset/dataset%2F2?format=full'
    )
    expect(state.activeDatasetId.value).toBe(1)
    expect(state.activeDataset.value).toMatchObject(selectedDetail)

    await state.selectDataset(1)
    expect(fetcher).toHaveBeenCalledTimes(2)
    expect(state.activeDataset.value).toMatchObject(selectedDetail)
  })

  it('retries a failed catalog detail request and replaces the summary', async () => {
    const catalog = [{ id: 'dataset-1', name: 'Dataset 1' }]
    const selectedDetail = {
      id: 'dataset-1',
      name: 'Dataset 1',
      data: [{ name: 'value', value: 1 }],
      settings: [{ type: 'line' as const }],
    }
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(catalog), { status: 200 }))
      .mockResolvedValueOnce(new Response('', { status: 503, statusText: 'Service Unavailable' }))
      .mockResolvedValueOnce(new Response(JSON.stringify(selectedDetail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    await expect(state.selectDataset(0)).resolves.toBe(false)
    expect(state.detailError.value).toContain('503 Service Unavailable')
    expect(state.activeDataset.value).toMatchObject({
      id: 'dataset-1',
      data: [],
      settings: [],
    })

    await expect(state.retryActiveDataset()).resolves.toBe(true)
    expect(state.detailError.value).toBeNull()
    expect(state.detailLoading.value).toBe(false)
    expect(state.activeDataset.value).toMatchObject(selectedDetail)
    expect(fetcher).toHaveBeenCalledTimes(3)
  })

  it('fetches only the encoded detail URL in path mode', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/dataset%2F2', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/dataset/?format=full#latest',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(fetcher).toHaveBeenCalledWith('https://api.example.com/dataset/dataset%2F2?format=full')
    expect(state.pathDatasetId).toBe('dataset/2')
    expect(state.loadError.value).toBeNull()
    expect(state.datasets.value).toHaveLength(1)
    expect(state.datasets.value[0]?.id).toBe('dataset/2')
    expect(state.lazyCatalog.value).toBe(false)
  })

  it('retries a failed path detail request', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response('', { status: 404, statusText: 'Not Found' }))
      .mockResolvedValueOnce(new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/my-id', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/dataset',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loadError.value).toContain('404 Not Found'))
    await state.retryActiveDataset()
    expect(fetcher).toHaveBeenCalledTimes(2)
    expect(state.loadError.value).toBeNull()
    expect(state.datasets.value[0]?.id).toBe('my-id')
  })

  it('disables path mode when the data URL does not end in dataset', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/ignored-id', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.pathDatasetId).toBeNull()
    expect(fetcher).toHaveBeenCalledOnce()
    expect(fetcher).toHaveBeenCalledWith('https://api.example.com/catalog')
  })

  it('uses the base URL instead of path mode for file pages', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/my-id', protocol: 'file:' },
      VIZB_DATA_URL: 'https://api.example.com/dataset',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.pathDatasetId).toBeNull()
    expect(fetcher).toHaveBeenCalledOnce()
    expect(fetcher).toHaveBeenCalledWith('https://api.example.com/dataset')
  })

  it('uses the base URL for a trailing-slash mount root', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/repo/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/dataset',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.pathDatasetId).toBeNull()
    expect(fetcher).toHaveBeenCalledWith('https://api.example.com/dataset')
  })

  it('ignores the path without a data URL', async () => {
    const fetcher = vi.fn()
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/ignored-id', protocol: 'https:' },
      VIZB_DATA: [detail],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.pathDatasetId).toBeNull()
    expect(fetcher).not.toHaveBeenCalled()
  })
})

describe('useDataPoint embedded VIZB_DATA', () => {
  beforeEach(() => {
    vi.resetModules()
    // Unit runs with import.meta.env.DEV=true; without a data URL that path
    // loads sample.json. Force the production embed branch (window.VIZB_DATA).
    vi.stubEnv('DEV', false)
  })
  afterEach(() => {
    vi.unstubAllEnvs()
    vi.unstubAllGlobals()
  })

  it('loads a single Dataset object from window.VIZB_DATA', async () => {
    const one = {
      name: 'Embedded One',
      data: [{ name: 'a', value: 1 }],
      settings: [{ type: 'bar' as const }],
    }
    const fetcher = vi.fn()
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA: one,
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.loadError.value).toBeNull()
    expect(state.lazyCatalog.value).toBe(false)
    expect(state.datasets.value).toHaveLength(1)
    expect(state.activeDataset.value?.name).toBe('Embedded One')
    expect(fetcher).not.toHaveBeenCalled()
  })

  it('applies a legacy hex theme string when the dataset has no themes[]', async () => {
    const one = {
      name: 'Legacy theme',
      theme: '#f00,#0f0,#00f',
      data: [{ name: 'a', value: 1 }],
      settings: [{ type: 'bar' as const }],
    }
    vi.stubGlobal('fetch', vi.fn())
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA: one,
    })

    const { useDataPoint } = await import('./useDataPoint')
    const { activeThemeName } = await import('../lib/themes')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.activeDataset.value?.theme).toBe('#f00,#0f0,#00f')
    expect(activeThemeName.value).toBe('#f00,#0f0,#00f')
  })

  it('loads a Dataset array with two entries', async () => {
    const payload = [
      {
        name: 'First',
        data: [{ name: 'a', value: 1 }],
        settings: [{ type: 'bar' as const }],
      },
      {
        name: 'Second',
        data: [{ name: 'b', value: 2 }],
        settings: [{ type: 'line' as const }],
      },
    ]
    vi.stubGlobal('fetch', vi.fn())
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA: payload,
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.datasets.value).toHaveLength(2)
    expect(state.datasets.value.map((d) => d.name)).toEqual(['First', 'Second'])
    expect(state.activeDataset.value?.name).toBe('First')
  })

  it('loads an empty array without error', async () => {
    vi.stubGlobal('fetch', vi.fn())
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.loadError.value).toBeNull()
    expect(state.datasets.value).toEqual([])
  })

  it('covers arrangement, groups, chart modes, and selection edge cases', async () => {
    const payload = [
      {
        name: 'Grouped',
        data: [{ name: 'a', xAxis: 'x1', yAxis: 'y1', value: 1 }],
        settings: [{ type: 'bar' as const }, { type: 'line' as const }],
        axes: [
          { key: 'name', label: 'n' },
          { key: 'x', label: 'x' },
          { key: 'metric', label: 'm' },
        ],
      },
      {
        name: 'Value',
        data: [{ xAxis: '1', yAxis: '2', value: 3 }],
        settings: [{ type: 'scatter' as const }],
        axes: [
          { key: 'x', label: 'price', type: 'value' as const },
          { key: 'y', label: 'latency', type: 'value' as const },
        ],
      },
      {
        name: 'Mixed',
        data: [{ xAxis: 'east', yAxis: '1.5', value: 4 }],
        settings: [{ type: 'scatter' as const }],
        axes: [
          { key: 'x', label: 'region' },
          { key: 'y', label: 'latency', type: 'value' as const },
        ],
      },
      {
        name: 'NoAxes',
        data: [{ name: 'only', value: 9 }],
        settings: [{ type: 'pie' as const }],
      },
    ]
    vi.stubGlobal('fetch', vi.fn())
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA: payload,
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))

    expect(state.chartMode.value).toBe('grouped')
    expect(state.isValueMode.value).toBe(false)
    expect(state.isMixedMode.value).toBe(false)
    expect(state.isValueModeDataset.value).toBe(false)
    expect(state.isMixedModeDataset.value).toBe(false)
    expect(state.activeArrangement.value.identityString).toBe('nx')
    expect(state.activeDataDimension.value).toBe('2D')

    state.setArrangement(0, 'bar', 'xn')
    expect(state.getArrangement(0, 'bar')).toBe('xn')
    expect(state.activeArrangement.value.targetString).toBe('xn')

    state.setGroupNames(['g0', 'g1'])
    expect(state.resultGroups.value).toEqual([{ name: 'g0' }, { name: 'g1' }])
    state.selectGroup(1)
    expect(state.activeGroupId.value).toBe(1)
    state.selectGroup(99)
    expect(state.activeGroupId.value).toBe(1)

    await expect(state.selectDataset(1)).resolves.toBe(true)
    expect(state.activeDatasetId.value).toBe(1)
    expect(state.activeGroupId.value).toBe(0)
    expect(state.chartMode.value).toBe('value')
    expect(state.isValueMode.value).toBe(true)
    expect(state.isValueModeDataset.value).toBe(true)
    expect(state.activeArrangement.value.identityString).toBe('xy')

    await expect(state.selectDataset(2)).resolves.toBe(true)
    expect(state.chartMode.value).toBe('mixed')
    expect(state.isMixedMode.value).toBe(true)
    expect(state.isMixedModeDataset.value).toBe(true)

    await expect(state.selectDataset(3)).resolves.toBe(true)
    expect(state.chartMode.value).toBe('grouped')
    expect(state.activeArrangement.value.identityString.length).toBeGreaterThan(0)
    expect(state.activeDataDimension.value).toBe('1D')

    await expect(state.selectDataset(99)).resolves.toBe(false)
    // Force the non-array guard in datasetsProcessed via activeDataset read.
    ;(state.datasets as { value: unknown }).value = {
      name: 'single',
      data: [],
      settings: [{ type: 'bar' as const }],
    }
    expect(state.activeDataset.value?.name).toBe('single')
    expect(Array.isArray(state.datasets.value)).toBe(true)
  })

  it('stringifies non-Error load failures', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => {
        throw 'boom'
      })
    )
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.loadError.value).toBe('boom')
  })

  it('falls back to empty VIZB_DATA when undefined', async () => {
    vi.stubGlobal('fetch', vi.fn())
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA: undefined,
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.datasets.value).toEqual([])
    expect(state.loadError.value).toBeNull()
  })

  it('stringifies non-Error catalog detail failures', async () => {
    const catalog = [{ id: 'dataset-1', name: 'Dataset 1' }]
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(catalog), { status: 200 }))
      .mockRejectedValueOnce('detail-boom')
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    await expect(state.selectDataset(0)).resolves.toBe(false)
    expect(state.detailError.value).toBe('detail-boom')
  })

  it('aborts stale catalog detail when selection changes mid-flight', async () => {
    const catalog = [
      { id: 'dataset-1', name: 'Dataset 1' },
      { id: 'dataset-2', name: 'Dataset 2' },
    ]
    let resolveDetail!: (value: Response) => void
    const detailPromise = new Promise<Response>((resolve) => {
      resolveDetail = resolve
    })
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(catalog), { status: 200 }))
      .mockReturnValueOnce(detailPromise)
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            id: 'dataset-2',
            name: 'Dataset 2',
            data: [{ name: 'b', value: 2 }],
            settings: [{ type: 'bar' as const }],
          }),
          { status: 200 }
        )
      )
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))

    const first = state.selectDataset(0)
    await expect(state.selectDataset(1)).resolves.toBe(true)
    resolveDetail(
      new Response(
        JSON.stringify({
          id: 'dataset-1',
          name: 'Dataset 1',
          data: [{ name: 'a', value: 1 }],
          settings: [{ type: 'line' as const }],
        }),
        { status: 200 }
      )
    )
    await expect(first).resolves.toBe(false)
    expect(state.activeDatasetId.value).toBe(1)
    expect(state.activeDataset.value?.id).toBe('dataset-2')
  })

  it('throws on non-ok catalog fetch responses', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('', { status: 500, statusText: 'Server Error' }))
    )
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.loadError.value).toContain('500 Server Error')
  })

  it('normalizes null settings/data on prepare and rejects catalog shells without id', async () => {
    const payload = [
      {
        name: 'NullFields',
        data: null,
        settings: null,
      },
    ]
    vi.stubGlobal('fetch', vi.fn())
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA: payload,
      VIZB_CHARTS: undefined,
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))

    expect(state.datasets.value[0]?.settings).toEqual([])
    expect(state.datasets.value[0]?.data).toEqual([])
  })

  it('returns false when a catalog shell has no id', async () => {
    // Bypass classify by injecting a lazy catalog shell directly after load.
    const catalog = [{ id: 'ok', name: 'Ok' }]
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(catalog), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))

    state.datasets.value = [{ name: 'NoId', data: [], settings: [] } as never]
    await expect(state.selectDataset(0)).resolves.toBe(false)
  })

  it('aborts catalog detail error when selection changed before catch', async () => {
    const catalog = [
      { id: 'dataset-1', name: 'Dataset 1' },
      { id: 'dataset-2', name: 'Dataset 2' },
    ]
    let rejectDetail!: (reason: unknown) => void
    const detailPromise = new Promise<Response>((_, reject) => {
      rejectDetail = reject
    })
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify(catalog), { status: 200 }))
      .mockReturnValueOnce(detailPromise)
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            id: 'dataset-2',
            name: 'Dataset 2',
            data: [{ name: 'b', value: 2 }],
            settings: [{ type: 'bar' as const }],
          }),
          { status: 200 }
        )
      )
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()
    await vi.waitFor(() => expect(state.loading.value).toBe(false))

    const first = state.selectDataset(0)
    await expect(state.selectDataset(1)).resolves.toBe(true)
    rejectDetail(new Error('stale'))
    await expect(first).resolves.toBe(false)
    expect(state.detailError.value).toBeNull()
  })
})
