import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const detail = {
  name: 'Go amd64',
  data: [],
  settings: [{ type: 'bar' as const }],
}

describe('useDataPoint path loading', () => {
  beforeEach(() => vi.resetModules())
  afterEach(() => vi.unstubAllGlobals())

  it('fetches only the encoded detail URL in path mode', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/go-1.25%2Famd64', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(fetcher).toHaveBeenCalledTimes(1)
    expect(fetcher).toHaveBeenCalledWith('https://api.example.com/catalog/dataset/go-1.25%2Famd64')
    expect(state.datasets.value).toHaveLength(1)
    expect(state.datasets.value[0]?.id).toBe('go-1.25/amd64')
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
      VIZB_DATA_URL: 'https://api.example.com/catalog',
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

  it('uses the base URL instead of path mode for file pages', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/my-id', protocol: 'file:' },
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

  it('uses the base URL for a trailing-slash mount root', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify(detail), { status: 200 }))
    vi.stubGlobal('fetch', fetcher)
    vi.stubGlobal('window', {
      location: { pathname: '/repo/', protocol: 'https:' },
      VIZB_DATA_URL: 'https://api.example.com/catalog',
      VIZB_DATA: [],
    })

    const { useDataPoint } = await import('./useDataPoint')
    const state = useDataPoint()

    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.pathDatasetId).toBeNull()
    expect(fetcher).toHaveBeenCalledWith('https://api.example.com/catalog')
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
