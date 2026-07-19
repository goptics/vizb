import type { DataSet } from '../types'

export type DatasetCatalogEntry = {
  id: string
  name: string
}

export type RemotePayload =
  | { mode: 'full'; datasets: DataSet[] }
  | { mode: 'catalog'; entries: DatasetCatalogEntry[] }

type Fetcher = (input: string | URL, init?: RequestInit) => Promise<Response>

const detailRequests = new Map<string, Promise<DataSet>>()

const isObject = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value)

const payloadShapeError = (message: string) =>
  new Error(
    `${message} Expected one full dataset object, an array of full dataset objects, ` +
      'or a catalog array of { id, name } entries that omit data and settings.'
  )

export const classifyRemotePayload = (payload: unknown): RemotePayload => {
  if (isObject(payload)) {
    return { mode: 'full', datasets: [payload as DataSet] }
  }
  if (!Array.isArray(payload)) {
    throw payloadShapeError('Invalid data-url response.')
  }
  if (payload.length === 0) {
    return { mode: 'full', datasets: [] }
  }

  let summaryCount = 0
  let fullCount = 0
  for (const [index, entry] of payload.entries()) {
    if (!isObject(entry)) {
      throw payloadShapeError(`Invalid entry at index ${index}.`)
    }
    const hasData = Object.hasOwn(entry, 'data')
    const hasSettings = Object.hasOwn(entry, 'settings')
    if (hasData !== hasSettings) {
      throw payloadShapeError(
        `Invalid entry at index ${index}: full datasets must contain both data and settings.`
      )
    }
    if (hasData) fullCount++
    else summaryCount++
  }

  if (summaryCount > 0 && fullCount > 0) {
    throw payloadShapeError('Invalid mixed array of catalog summaries and full datasets.')
  }
  if (fullCount > 0) {
    return { mode: 'full', datasets: payload as DataSet[] }
  }

  const ids = new Set<string>()
  const entries = payload.map((entry, index) => {
    const candidate = entry as Record<string, unknown>
    if (typeof candidate.id !== 'string' || candidate.id.trim() === '') {
      throw payloadShapeError(`Invalid catalog entry at index ${index}: id must be non-empty.`)
    }
    if (typeof candidate.name !== 'string' || candidate.name.trim() === '') {
      throw payloadShapeError(`Invalid catalog entry at index ${index}: name must be non-empty.`)
    }
    if (ids.has(candidate.id)) {
      throw payloadShapeError(`Invalid catalog: duplicate dataset id "${candidate.id}".`)
    }
    ids.add(candidate.id)
    return { id: candidate.id, name: candidate.name }
  })

  return { mode: 'catalog', entries }
}

export const buildDatasetDetailUrl = (baseUrl: string, id: string): string => {
  const url = new URL(baseUrl)
  const basePath = url.pathname.replace(/\/+$/, '')
  url.pathname = `${basePath}/dataset/${encodeURIComponent(id)}`
  url.hash = ''
  return url.toString()
}

const validateDetail = (payload: unknown, requestedId: string): DataSet => {
  if (!isObject(payload)) {
    throw new Error(
      `Invalid detail response for dataset "${requestedId}": expected one full dataset object.`
    )
  }
  if (!Array.isArray(payload.data) || !Array.isArray(payload.settings)) {
    throw new Error(
      `Invalid detail response for dataset "${requestedId}": ` +
        'the object must contain data and settings arrays.'
    )
  }
  if (typeof payload.name !== 'string' || payload.name.trim() === '') {
    throw new Error(`Invalid detail response for dataset "${requestedId}": name must be non-empty.`)
  }
  if (payload.id === undefined) {
    return { ...payload, id: requestedId } as DataSet
  }
  if (payload.id !== requestedId) {
    throw new Error(
      `Dataset detail ID mismatch: requested "${requestedId}" but received "${String(payload.id)}".`
    )
  }
  return payload as DataSet
}

export const fetchDatasetDetail = (
  baseUrl: string,
  id: string,
  fetcher: Fetcher = fetch
): Promise<DataSet> => {
  const pending = detailRequests.get(id)
  if (pending) return pending

  const request = (async () => {
    const url = buildDatasetDetailUrl(baseUrl, id)
    const response = await fetcher(url)
    if (!response.ok) {
      throw new Error(
        `Failed to load dataset "${id}": ${response.status} ${response.statusText}`.trim()
      )
    }
    return validateDetail(await response.json(), id)
  })()

  detailRequests.set(id, request)
  const removeRequest = () => {
    if (detailRequests.get(id) === request) detailRequests.delete(id)
  }
  void request.then(removeRequest, removeRequest)
  return request
}
