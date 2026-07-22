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
