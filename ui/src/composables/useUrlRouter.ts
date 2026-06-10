import { watch } from 'vue'
import type { SortOrder, ChartType, ScaleType } from '../types'
import { SORT_ORDERS, SCALE_TYPES } from '../types'
import { useSettingsStore } from './useSettingsStore'
import { useDataPoint } from './useDataPoint'
import { DEFAULT_SETTINGS } from './constants'
import { isValidIndex } from '../lib/utils'

type UrlParams = {
  s?: SortOrder // sort order
  l?: 'true' | 'false' // show labels
  c?: ChartType // chart type
  sc?: ScaleType // scale type
  d?: string // dataset ID
  g?: string // group ID
}

/**
 * Parse query params from the current URL
 */
const parseUrlParams = (): UrlParams => {
  const params = new URLSearchParams(window.location.search)
  return {
    s: params.get('s') as SortOrder | undefined,
    l: params.get('l') as 'true' | 'false' | undefined,
    c: params.get('c') as ChartType | undefined,
    sc: params.get('sc') as ScaleType | undefined,
    d: params.get('d') ?? undefined,
    g: params.get('g') ?? undefined,
  }
}

/**
 * Build URL query string from current state
 */
const buildQueryString = (params: Record<string, string | undefined>): string => {
  const searchParams = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') {
      searchParams.set(key, value)
    }
  }
  const queryString = searchParams.toString()
  return queryString ? `?${queryString}` : ''
}

// Helper to apply index-based params
const applyIndexParam = (
  value: string | undefined,
  maxLength: number,
  setter: (id: number) => void
) => {
  if (value === undefined) return
  const id = parseInt(value, 10)
  if (!isNaN(id) && isValidIndex(id, maxLength)) {
    setter(id)
  }
}

/**
 * Composable for syncing app state with URL query parameters
 */
export function useUrlRouter() {
  const { settings, setSort, setShowLabels, setChartType, setScale } = useSettingsStore()
  const {
    dataSets,
    resultGroups,
    activeDataSetId,
    activeGroupId,
    selectDataSet,
    selectGroup,
  } = useDataPoint()

  /**
   * Apply parsed URL params to app state
   * Order matters: select benchmark/group first, then apply settings on top
   */
  const applyParams = (params: UrlParams) => {
    // 1. DataSet ID - apply first since it resets settings
    applyIndexParam(params.d, dataSets.value.length, selectDataSet)

    // 2. Group ID - depends on benchmark selection
    applyIndexParam(params.g, resultGroups.value.length, selectGroup)

    // 3. Now apply settings on top of benchmark defaults
    // Sort
    if (params.s && SORT_ORDERS.includes(params.s.toLowerCase() as SortOrder)) {
      setSort({ enabled: true, order: params.s })
    }

    // Labels
    switch (params.l?.toLowerCase()) {
      case 'true':
        setShowLabels(true)
        break
      case 'false':
        setShowLabels(false)
    }

    // Chart type
    if (params.c && DEFAULT_SETTINGS.charts.includes(params.c)) {
      setChartType(params.c)
    }

    // Scale type
    if (params.sc && SCALE_TYPES.includes(params.sc)) {
      setScale(params.sc)
    }
  }

  /**
   * Update URL with current state (without adding to history)
   */
  const syncUrlToState = () => {
    const params: Record<string, string | undefined> = {}

    // Sort - only include if enabled
    if (settings.sort.enabled) {
      params.s = settings.sort.order
    }

    // Labels - only include if true
    if (settings.showLabels) {
      params.l = 'true'
    }

    const chartType = settings.charts[settings.activeChartIndex]

    if (chartType && settings.activeChartIndex !== 0) {
      params.c = chartType
    }

    // Scale - only include if not linear (default)
    if (settings.scale !== 'linear') {
      params.sc = settings.scale
    }

    // DataSet ID - only include if not first
    if (activeDataSetId.value > 0) {
      params.d = activeDataSetId.value.toString()
    }

    // Group ID - only include if not first
    if (activeGroupId.value > 0) {
      params.g = activeGroupId.value.toString()
    }

    const queryString = buildQueryString(params)
    const newUrl = window.location.pathname + queryString

    if (newUrl !== window.location.pathname + window.location.search) {
      window.history.replaceState(null, '', newUrl)
    }
  }

  /**
   * Initialize app state from URL and set up watchers
   */
  const initFromUrl = () => {
    const params = parseUrlParams()
    applyParams(params)

    // Watch for state changes and update URL
    watch(
      () => ({
        sortEnabled: settings.sort.enabled,
        sortOrder: settings.sort.order,
        showLabels: settings.showLabels,
        chartIndex: settings.activeChartIndex,
        scale: settings.scale,
        benchmarkId: activeDataSetId.value,
        groupId: activeGroupId.value,
      }),
      () => syncUrlToState(),
      { deep: true }
    )
  }

  return {
    initFromUrl,
    syncUrlToState,
  }
}
