import { watch } from 'vue'
import type { ChartType, ChartSettings, SortOrder, ScaleType } from '../types'
import { SORT_ORDERS, SCALE_TYPES } from '../types'
import { useSettingsStore } from './useSettingsStore'
import { useDataPoint } from './useDataPoint'
import { DEFAULT_SETTINGS } from './constants'
import { isValidIndex } from '../lib/utils'

const ALL_CHART_TYPES = DEFAULT_SETTINGS.charts

const buildQueryString = (params: Record<string, string | undefined>): string => {
  const searchParams = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') {
      searchParams.set(key, value)
    }
  }
  const qs = searchParams.toString()
  return qs ? `?${qs}` : ''
}

const applyIndexParam = (
  value: string | undefined,
  maxLength: number,
  setter: (id: number) => void
) => {
  if (value === undefined) return
  const id = parseInt(value, 10)
  if (!isNaN(id) && isValidIndex(id, maxLength)) setter(id)
}

export function useUrlRouter() {
  const {
    settings,
    setSort,
    setShowLabels,
    setScale,
    setChartType,
    setChartSettingsForType,
  } = useSettingsStore()

  const {
    dataSets,
    resultGroups,
    activeDataSetId,
    activeGroupId,
    activeArrangement,
    selectDataSet,
    selectGroup,
    setArrangement,
    arrangementMap,
  } = useDataPoint()

  const parseUrlParams = () => {
    const p = new URLSearchParams(window.location.search)
    const result: Record<string, string | undefined> = {}
    for (const [key, value] of p.entries()) {
      result[key] = value
    }
    return result
  }

  const applyParams = (params: Record<string, string | undefined>) => {
    // 1. Dataset ID
    applyIndexParam(params.d, dataSets.value.length, selectDataSet)

    // 2. Group ID — deferred if groups not yet populated (worker populates asynchronously)
    const gParam = params.g
    if (resultGroups.value.length > 0) {
      applyIndexParam(gParam, resultGroups.value.length, selectGroup)
    } else if (gParam !== undefined) {
      watch(
        () => resultGroups.value.length,
        (len) => { if (len > 0) applyIndexParam(gParam, len, selectGroup) },
        { once: true }
      )
    }

    // 3. Active chart type
    if (params.c && DEFAULT_SETTINGS.charts.includes(params.c as ChartType)) {
      setChartType(params.c as ChartType)
    }

    // 4. Legacy global params (back-compat for old URLs with bare s/l/sc)
    const legacyS = params.s as SortOrder | undefined
    if (legacyS && SORT_ORDERS.includes(legacyS.toLowerCase() as SortOrder)) {
      setSort({ enabled: true, order: legacyS.toLowerCase() as SortOrder })
    }
    const legacyL = params.l
    if (legacyL === 'true') setShowLabels(true)
    else if (legacyL === 'false') setShowLabels(false)
    const legacySc = params.sc as ScaleType | undefined
    if (legacySc && SCALE_TYPES.includes(legacySc)) setScale(legacySc)

    // 5. Per-chart settings
    const datasetId = params.d !== undefined ? (parseInt(params.d, 10) || 0) : 0
    for (const ct of ALL_CHART_TYPES) {
      const so = params[`${ct}.so`] as SortOrder | undefined
      const l = params[`${ct}.l`]
      const sc = params[`${ct}.sc`] as ScaleType | undefined
      const rt = params[`${ct}.rt`]
      const sw = params[`${ct}.sw`]

      const update: Partial<Omit<ChartSettings, 'swap'>> = {}
      if (so && SORT_ORDERS.includes(so.toLowerCase() as SortOrder)) {
        update.sort = { enabled: true, order: so.toLowerCase() as SortOrder }
      }
      if (l === 'true') update.showLabels = true
      else if (l === 'false') update.showLabels = false
      if (sc && SCALE_TYPES.includes(sc)) update.scale = sc
      if (rt === 'true') update.autoRotate = true

      if (Object.keys(update).length > 0) {
        setChartSettingsForType(ct as ChartType, update)
      }

      // Swap: apply after data loads (same defer pattern as group ID)
      if (sw) {
        const applySwap = () => setArrangement(datasetId, ct as ChartType, sw)
        if (dataSets.value.length > 0) {
          applySwap()
        } else {
          watch(
            () => dataSets.value.length,
            (len) => { if (len > 0) applySwap() },
            { once: true }
          )
        }
      }
    }
  }

  const syncUrlToState = () => {
    const params: Record<string, string | undefined> = {}
    const identity = activeArrangement.value.identityString

    // Active chart tab (omit if first)
    if (settings.activeChartIndex !== 0) {
      params.c = settings.charts[settings.activeChartIndex]
    }

    // Dataset / group (omit if first)
    if (activeDataSetId.value > 0) params.d = activeDataSetId.value.toString()
    if (activeGroupId.value > 0) params.g = activeGroupId.value.toString()

    // Per-chart settings
    for (const ct of settings.charts) {
      const cs = settings.chartSettings[ct as ChartType]
      if (cs?.sort?.enabled) params[`${ct}.so`] = cs.sort.order
      if (cs?.showLabels === true) params[`${ct}.l`] = 'true'
      else if (cs?.showLabels === false) params[`${ct}.l`] = 'false'
      if (cs?.scale && cs.scale !== 'linear') params[`${ct}.sc`] = cs.scale
      if (cs?.autoRotate === true) params[`${ct}.rt`] = 'true'

      const arr = arrangementMap.get(`${activeDataSetId.value}:${ct}`)
      if (arr && arr !== identity) params[`${ct}.sw`] = arr
    }

    const queryString = buildQueryString(params)
    const newUrl = window.location.pathname + queryString
    if (newUrl !== window.location.pathname + window.location.search) {
      window.history.replaceState(null, '', newUrl)
    }
  }

  const initFromUrl = () => {
    const params = parseUrlParams()
    applyParams(params)

    watch(
      () => ({
        chartIndex: settings.activeChartIndex,
        benchmarkId: activeDataSetId.value,
        groupId: activeGroupId.value,
        swaps: settings.charts
          .map((ct) => arrangementMap.get(`${activeDataSetId.value}:${ct}`) ?? '')
          .join(','),
        csStr: Object.entries(settings.chartSettings)
          .map(([k, v]) => `${k}:${JSON.stringify(v)}`)
          .join('|'),
      }),
      () => syncUrlToState()
    )
  }

  return { initFromUrl, syncUrlToState }
}
