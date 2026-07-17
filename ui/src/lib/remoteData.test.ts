import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { DataSet } from '../types'
import {
  LazyDatasetSelection,
  buildDatasetDetailUrl,
  classifyRemotePayload,
  clearDatasetDetailCache,
  fetchDatasetDetail,
} from './remoteData'

const detail = (id?: string): DataSet => ({
  ...(id === undefined ? {} : { id }),
  name: id ?? 'Requested dataset',
  data: [],
  settings: [{ type: 'bar' }],
})

const jsonResponse = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), {
    status,
    statusText: status === 200 ? 'OK' : 'Server Error',
    headers: { 'Content-Type': 'application/json' },
  })

describe('remote data payloads', () => {
  beforeEach(clearDatasetDetailCache)
  afterEach(() => vi.restoreAllMocks())

  it('keeps full object and full-array payloads in eager mode', () => {
    const one = detail('one')
    expect(classifyRemotePayload(one)).toEqual({ mode: 'full', datasets: [one] })
    expect(classifyRemotePayload([one, detail('two')])).toEqual({
      mode: 'full',
      datasets: [one, detail('two')],
    })
  })

  it('detects a valid id/name catalog', () => {
    expect(
      classifyRemotePayload([
        { id: 'one', name: 'One' },
        { id: 'two', name: 'Two' },
      ])
    ).toEqual({
      mode: 'catalog',
      entries: [
        { id: 'one', name: 'One' },
        { id: 'two', name: 'Two' },
      ],
    })
  })

  it.each([
    {
      name: 'missing ID',
      payload: [{ name: 'One' }],
      message: 'id must be non-empty',
    },
    {
      name: 'empty ID',
      payload: [{ id: ' ', name: 'One' }],
      message: 'id must be non-empty',
    },
    {
      name: 'duplicate ID',
      payload: [
        { id: 'same', name: 'One' },
        { id: 'same', name: 'Two' },
      ],
      message: 'duplicate dataset id "same"',
    },
    {
      name: 'mixed catalog and datasets',
      payload: [{ id: 'one', name: 'One' }, detail('two')],
      message: 'mixed array',
    },
    {
      name: 'partial full dataset',
      payload: [{ id: 'one', name: 'One', data: [] }],
      message: 'both data and settings',
    },
  ])('rejects $name with the expected-shape guidance', ({ payload, message }) => {
    expect(() => classifyRemotePayload(payload)).toThrow(message)
    expect(() => classifyRemotePayload(payload)).toThrow('Expected one full dataset object')
  })

  it('builds one encoded detail path segment and preserves only the query', () => {
    expect(
      buildDatasetDetailUrl(
        'https://example.com/api/catalog///?token=a%20b#section',
        'suite/hello world?'
      )
    ).toBe('https://example.com/api/catalog/dataset/suite%2Fhello%20world%3F?token=a%20b')
    expect(buildDatasetDetailUrl('https://example.com/', 'one')).toBe(
      'https://example.com/dataset/one'
    )
  })

  it('fills an omitted detail ID and caches a successful response', async () => {
    const fetcher = vi.fn(async () => jsonResponse(detail()))
    const first = await fetchDatasetDetail('https://example.com/catalog', 'requested', fetcher)
    const second = await fetchDatasetDetail('https://example.com/catalog', 'requested', fetcher)

    expect(first.id).toBe('requested')
    expect(second).toBe(first)
    expect(fetcher).toHaveBeenCalledTimes(1)
  })

  it('shares one in-flight detail promise', async () => {
    let resolveResponse!: (response: Response) => void
    const fetcher = vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          resolveResponse = resolve
        })
    )

    const first = fetchDatasetDetail('https://example.com/catalog', 'one', fetcher)
    const second = fetchDatasetDetail('https://example.com/catalog', 'one', fetcher)
    expect(second).toBe(first)
    expect(fetcher).toHaveBeenCalledTimes(1)

    resolveResponse(jsonResponse(detail('one')))
    await expect(first).resolves.toMatchObject({ id: 'one' })
  })

  it('does not cache failures and allows retry', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ error: true }, 500))
      .mockResolvedValueOnce(jsonResponse(detail('one')))

    await expect(fetchDatasetDetail('https://example.com/catalog', 'one', fetcher)).rejects.toThrow(
      '500 Server Error'
    )
    await expect(
      fetchDatasetDetail('https://example.com/catalog', 'one', fetcher)
    ).resolves.toMatchObject({ id: 'one' })
    expect(fetcher).toHaveBeenCalledTimes(2)
  })

  it.each([
    { body: [detail('one')], message: 'expected one full dataset object' },
    { body: { id: 'one', name: 'One', data: [] }, message: 'data and settings arrays' },
    { body: detail('different'), message: 'ID mismatch' },
  ])('rejects malformed or inconsistent detail responses', async ({ body, message }) => {
    const fetcher = vi.fn(async () => jsonResponse(body))
    await expect(fetchDatasetDetail('https://example.com/catalog', 'one', fetcher)).rejects.toThrow(
      message
    )
  })

  it('marks a late response for an earlier selection as stale', async () => {
    const responses = new Map<string, (response: Response) => void>()
    const fetcher = vi.fn(
      (input: string | URL) =>
        new Promise<Response>((resolve) => {
          responses.set(String(input), resolve)
        })
    )
    const selection = new LazyDatasetSelection('https://example.com/catalog', fetcher)

    const a = selection.load('a')
    const b = selection.load('b')
    responses.get('https://example.com/catalog/dataset/b')!(jsonResponse(detail('b')))
    await expect(b).resolves.toMatchObject({ ok: true, dataset: { id: 'b' }, current: true })

    responses.get('https://example.com/catalog/dataset/a')!(jsonResponse(detail('a')))
    await expect(a).resolves.toMatchObject({ ok: true, dataset: { id: 'a' }, current: false })
  })
})
