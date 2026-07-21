import { watch, computed } from 'vue'
import type {
  ChartType,
  ChartConfig,
  SortOrder,
  ScaleType,
  BarConfig,
  LineConfig,
  ScatterConfig,
  Dataset,
} from '../types'
import { SORT_ORDERS, SCALE_TYPES } from '../types'
import { useSettingsStore } from './useSettingsStore'
import { useDataPoint } from './useDataPoint'
import { activeDataset } from './useDataPoint'
import { ALL_CHART_TYPES } from './constants'
import { isValidIndex } from '../lib/utils'

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

const resolveDatasetIndex = (
  params: Record<string, string | undefined>,
  datasets: Dataset[],
  pathDatasetId: string | null
): number => {
  if (pathDatasetId) return 0
  const idParam = params.id?.trim()
  if (idParam) {
    const idx = datasets.findIndex((ds) => ds.id === idParam)
    if (idx >= 0) return idx
  }
  if (params.d !== undefined) {
    const n = parseInt(params.d, 10)
    if (!isNaN(n) && isValidIndex(n, datasets.length)) return n
  }
  return 0
}

// Field update payload accepted by `applyConfigUpdate`. Only the keys present
// in the URL get touched; missing fields stay at whatever the config already
// holds. `scale` and 3D fields are skipped for pie/heatmap/radar (those
// types don't carry the field) so a malformed URL can't widen the config.
type ConfigUpdate = {
  sort?: { enabled: boolean; order: SortOrder }
  showLabels?: boolean
  scale?: ScaleType
  threeD?: boolean
  threeDRotate?: boolean
  threeDVisualMap?: boolean
  visualMap?: boolean
  horizontal?: boolean
}

// Find the first config of the given chart type and apply a partial update in
// place. The dataset slice is the source of truth — Vue reactivity propagates
// the mutation to every consumer (chart pipeline, panel, etc.). Returns true
// when an update was applied.
const applyConfigUpdate = (type: ChartType, update: ConfigUpdate): boolean => {
  const settings = activeDataset.value?.settings
  if (!settings) return false
  const cfg = settings.find((s) => s.type === type)
  if (!cfg) return false

  if (update.sort) cfg.sort = update.sort
  if (update.showLabels !== undefined) cfg.showLabels = update.showLabels
  if (cfg.type === 'bar' || cfg.type === 'line' || cfg.type === 'scatter') {
    const cartesian = cfg as BarConfig | LineConfig | ScatterConfig
    if (update.scale) cartesian.scale = update.scale
    if (update.threeD !== undefined) cartesian.threeD = update.threeD
    if (update.threeDRotate !== undefined) cartesian.threeDRotate = update.threeDRotate
    if (update.threeDVisualMap !== undefined) cartesian.threeDVisualMap = update.threeDVisualMap
    if (cfg.type === 'scatter' && update.visualMap !== undefined) {
      ;(cartesian as ScatterConfig).visualMap = update.visualMap
    }
    if (cfg.type === 'bar' && update.horizontal !== undefined) {
      ;(cfg as BarConfig).horizontal = update.horizontal
    }
  }
  return true
}

export function useUrlRouter() {
  const { activeChartIndex, chartType, setChartType } = useSettingsStore()

  const {
    datasets,
    resultGroups,
    activeDatasetId,
    activeGroupId,
    activeArrangement,
    selectDataset,
    selectGroup,
    setArrangement,
    arrangementMap,
    pathDatasetId,
  } = useDataPoint()

  // Chart-type list for the active dataset, derived from its `settings` array.
  const availableTypes = computed<ChartType[]>(
    () => activeDataset.value?.settings.map((s) => s.type) ?? []
  )

  const parseUrlParams = () => {
    const p = new URLSearchParams(window.location.search)
    const result: Record<string, string | undefined> = {}
    for (const [key, value] of p.entries()) {
      result[key] = value
    }
    return result
  }

  const applyParams = async (params: Record<string, string | undefined>) => {
    // 1. Dataset selection (path identity wins, otherwise ?id= wins over ?d=)
    const datasetId = resolveDatasetIndex(params, datasets.value, pathDatasetId)
    const catalogShell = datasets.value[datasetId]
    const selected = await selectDataset(datasetId)
    if (!selected) {
      // A failed lazy detail remains retryable. When Retry replaces this catalog
      // shell with a loaded detail, re-run initialization so group/chart/swap
      // parameters are applied to the real settings rather than the summary.
      watch(
        () => datasets.value[datasetId],
        (dataset) => {
          if (dataset !== catalogShell && activeDatasetId.value === datasetId) {
            void applyParams(params)
          }
        },
        { once: true }
      )
      return false
    }

    // 2. Group ID — deferred if groups not yet populated (worker populates asynchronously)
    const gParam = params.g
    if (resultGroups.value.length > 0) {
      applyIndexParam(gParam, resultGroups.value.length, selectGroup)
    } else if (gParam !== undefined) {
      watch(
        () => resultGroups.value.length,
        (len) => {
          if (len > 0) applyIndexParam(gParam, len, selectGroup)
        },
        { once: true }
      )
    }

    // 3. Active chart type
    if (params.c && ALL_CHART_TYPES.includes(params.c as ChartType)) {
      setChartType(params.c as ChartType)
    }

    // 4. Legacy global params (back-compat for old URLs with bare s/l/sc) —
    // applied to every chart config that exists in the active dataset.
    const legacyS = params.s as SortOrder | undefined
    const legacyL = params.l
    const legacySc = params.sc as ScaleType | undefined
    const globalUpdate: ConfigUpdate = {}
    if (legacyS && SORT_ORDERS.includes(legacyS.toLowerCase() as SortOrder)) {
      globalUpdate.sort = { enabled: true, order: legacyS.toLowerCase() as SortOrder }
    }
    if (legacyL === 'true') globalUpdate.showLabels = true
    else if (legacyL === 'false') globalUpdate.showLabels = false
    if (legacySc && SCALE_TYPES.includes(legacySc)) globalUpdate.scale = legacySc
    if (Object.keys(globalUpdate).length > 0) {
      for (const t of availableTypes.value) applyConfigUpdate(t, globalUpdate)
    }

    // 5. Per-chart settings
    for (const ct of ALL_CHART_TYPES) {
      const so = params[`${ct}.so`] as SortOrder | undefined
      const l = params[`${ct}.l`]
      const sc = params[`${ct}.sc`] as ScaleType | undefined
      const d3 = params[`${ct}.3d`]
      const d3rt = params[`${ct}.3d-rt`]
      const d3vm = params[`${ct}.3d-vm`]
      const vm = params[`${ct}.vm`]
      const sw = params[`${ct}.sw`]

      const update: ConfigUpdate = {}
      if (so && SORT_ORDERS.includes(so.toLowerCase() as SortOrder)) {
        update.sort = { enabled: true, order: so.toLowerCase() as SortOrder }
      }
      if (l === 'true') update.showLabels = true
      else if (l === 'false') update.showLabels = false
      if (sc && SCALE_TYPES.includes(sc)) update.scale = sc
      if (d3 === 'true') update.threeD = true
      else if (d3 === 'false') update.threeD = false
      if (d3rt === 'true') update.threeDRotate = true
      if (d3vm === 'true') update.threeDVisualMap = true
      else if (d3vm === 'false') update.threeDVisualMap = false
      if (ct === 'scatter') {
        if (vm === 'true') update.visualMap = true
        else if (vm === 'false') update.visualMap = false
      }

      const h = params[`${ct}.h`]
      if (h === 'true') update.horizontal = true
      else if (h === 'false') update.horizontal = false

      if (Object.keys(update).length > 0) {
        applyConfigUpdate(ct, update)
      }

      // Swap: apply after data loads (same defer pattern as group ID)
      if (sw) {
        const applySwap = () => setArrangement(datasetId, ct, sw)
        if (datasets.value.length > 0) {
          applySwap()
        } else {
          watch(
            () => datasets.value.length,
            (len) => {
              if (len > 0) applySwap()
            },
            { once: true }
          )
        }
      }
    }
    return true
  }

  const syncUrlToState = () => {
    const params: Record<string, string | undefined> = {}
    const identity = activeArrangement.value.identityString
    const settings: ChartConfig[] = activeDataset.value?.settings ?? []

    // Active chart tab (omit if first)
    const activeCfg = settings[activeChartIndex.value]
    if (activeChartIndex.value !== 0 && activeCfg) {
      params.c = activeCfg.type
    }

    // Dataset / group
    if (!pathDatasetId) {
      const datasetId = activeDataset.value?.id?.trim()
      if (datasetId) {
        params.id = datasetId
      } else if (activeDatasetId.value > 0) {
        params.d = activeDatasetId.value.toString()
      }
    }
    if (activeGroupId.value > 0) params.g = activeGroupId.value.toString()

    // Per-chart settings
    for (const cfg of settings) {
      const ct = cfg.type
      if (cfg.sort?.enabled) params[`${ct}.so`] = cfg.sort.order
      if (cfg.showLabels === true) params[`${ct}.l`] = 'true'
      else if (cfg.showLabels === false) params[`${ct}.l`] = 'false'
      if (cfg.type === 'bar' || cfg.type === 'line' || cfg.type === 'scatter') {
        const cartesian = cfg as BarConfig | LineConfig | ScatterConfig
        if (cartesian.scale && cartesian.scale !== 'linear') params[`${ct}.sc`] = cartesian.scale
        if (cartesian.threeD === true) params[`${ct}.3d`] = 'true'
        if (cartesian.threeDRotate === true) params[`${ct}.3d-rt`] = 'true'
        if (cartesian.threeDVisualMap === true) params[`${ct}.3d-vm`] = 'true'
        else if (cartesian.threeDVisualMap === false) params[`${ct}.3d-vm`] = 'false'
        if (cfg.type === 'scatter') {
          const scatter = cartesian as ScatterConfig
          if (scatter.visualMap === true) params[`${ct}.vm`] = 'true'
          else if (scatter.visualMap === false) params[`${ct}.vm`] = 'false'
        }
        if (cfg.type === 'bar') {
          const barCfg = cfg as BarConfig
          if (barCfg.horizontal === true) params[`${ct}.h`] = 'true'
        }
      }

      const arr = arrangementMap.get(`${activeDatasetId.value}:${ct}`)
      if (arr && arr !== identity) params[`${ct}.sw`] = arr
    }

    const queryString = buildQueryString(params)
    const newUrl = window.location.pathname + queryString
    if (newUrl !== window.location.pathname + window.location.search) {
      window.history.replaceState(null, '', newUrl)
    }
  }

  const initFromUrl = async () => {
    const params = parseUrlParams()
    const applied = await applyParams(params)

    // Re-sync the URL whenever any source of truth (active index, dataset,
    // group, per-chart swap, or per-chart config) changes. The config
    // serialization is JSON so per-field mutations propagate without manual
    // bookkeeping.
    watch(
      () => ({
        chartIndex: activeChartIndex.value,
        benchmarkId: activeDatasetId.value,
        groupId: activeGroupId.value,
        chartType: chartType.value,
        swaps: availableTypes.value
          .map((ct) => arrangementMap.get(`${activeDatasetId.value}:${ct}`) ?? '')
          .join(','),
        csStr: JSON.stringify(activeDataset.value?.settings ?? []),
      }),
      () => syncUrlToState()
    )
    if (applied) syncUrlToState()
    return applied
  }

  return { initFromUrl, syncUrlToState }
}
