import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Dataset } from '../types'
import {
  buildDatasetDetailUrl,
  classifyPayload,
  fetchDatasetDetail,
  isDatasetCollectionUrl,
} from './remoteData'

const detail = (id?: string): Dataset => ({
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
  afterEach(() => vi.restoreAllMocks())

  it('keeps full object and full-array payloads in eager mode', () => {
    const one = detail('one')
    expect(classifyPayload(one)).toEqual({ mode: 'full', datasets: [one] })
    expect(classifyPayload([one, detail('two')])).toEqual({
      mode: 'full',
      datasets: [one, detail('two')],
    })
  })

  it('detects a valid id/name catalog', () => {
    expect(
      classifyPayload([
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
    expect(() => classifyPayload(payload)).toThrow(message)
    expect(() => classifyPayload(payload)).toThrow('Expected one full dataset object')
  })

  it('detects only data URLs whose path ends in the dataset collection', () => {
    expect(isDatasetCollectionUrl('https://example.com/api/dataset')).toBe(true)
    expect(isDatasetCollectionUrl('https://example.com/api/dataset/?format=full#latest')).toBe(true)
    expect(isDatasetCollectionUrl('https://example.com/api/datasets')).toBe(false)
    expect(isDatasetCollectionUrl('https://example.com/api/dataset.json')).toBe(false)
    expect(isDatasetCollectionUrl('not a URL')).toBe(false)
  })

  it('builds one encoded detail path segment and preserves only the query', () => {
    expect(
      buildDatasetDetailUrl(
        'https://example.com/api/catalog///?filter=a%20b#section',
        'suite/hello world?'
      )
    ).toBe('https://example.com/api/catalog/dataset/suite%2Fhello%20world%3F?filter=a%20b')
    expect(buildDatasetDetailUrl('https://example.com/', 'one')).toBe(
      'https://example.com/dataset/one'
    )
    expect(
      buildDatasetDetailUrl('https://example.com/api/dataset/?format=full#latest', 'suite/one')
    ).toBe('https://example.com/api/dataset/suite%2Fone?format=full')
  })

  it('fills an omitted detail ID', async () => {
    const fetcher = vi.fn(async () => jsonResponse(detail()))
    const dataset = await fetchDatasetDetail('https://example.com/catalog', 'requested', fetcher)

    expect(dataset.id).toBe('requested')
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
})
