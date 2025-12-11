import { watch } from 'vue'
import type { SortOrder, ChartType } from '../types'
import { useSettingsStore } from './useSettingsStore'
import { useBenchmarkData } from './useBenchmarkData'
import { DEFAULT_SETTINGS } from './constants'

type UrlParams = {
  s?: SortOrder // sort order
  l?: 'true' | 'false' // show labels
  c?: ChartType // chart type
  b?: string // benchmark ID
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
    b: params.get('b') ?? undefined,
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
  if (!isNaN(id) && id >= 0 && id < maxLength) {
    setter(id)
  }
}

/**
 * Composable for syncing app state with URL query parameters
 */
export function useUrlRouter() {
  const { settings, setSort, setShowLabels, setChartType } = useSettingsStore()
  const {
    benchmarks,
    resultGroups,
    activeBenchmarkId,
    activeGroupId,
    selectBenchmark,
    selectGroup,
  } = useBenchmarkData()

  /**
   * Apply parsed URL params to app state
   * Order matters: select benchmark/group first, then apply settings on top
   */
  const applyParams = (params: UrlParams) => {
    // 1. Benchmark ID - apply first since it resets settings
    applyIndexParam(params.b, benchmarks.value.length, selectBenchmark)

    // 2. Group ID - depends on benchmark selection
    applyIndexParam(params.g, resultGroups.value.length, selectGroup)

    // 3. Now apply settings on top of benchmark defaults
    // Sort
    if (params.s && ['asc', 'desc'].includes(params.s.toLowerCase())) {
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

    // Benchmark ID - only include if not first
    if (activeBenchmarkId.value > 0) {
      params.b = activeBenchmarkId.value.toString()
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
        benchmarkId: activeBenchmarkId.value,
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
