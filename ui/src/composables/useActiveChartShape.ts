import { computed } from 'vue'
import type {
  BarConfig,
  LabelMode,
  LineConfig,
  ScatterConfig,
  ScaleType,
  Sort,
  StatConfig,
} from '../types'
import { arrangementHasChartZ } from '../lib/swap'
import { canOfferValue3D, resolveLabelMode } from '../lib/utils'
import { useDataPoint } from './useDataPoint'
import { useSettingsStore } from './useSettingsStore'

// Pure read-only composable: derives everything from the store's `activeConfig`
// (the typed per-chart config at `dataset.value.settings[activeChartIndex]`).
// Every field uses `?? default` — no `cfg.type === 'bar' || ...` branching.
// For pie/heatmap/radar the optional `scale` / `threeDRotate` fields don't exist
// on the type at all, so `?? default` is the only mechanism that produces a
// value (TypeScript's optional-field semantics in place of a runtime guard).
export function useActiveChartShape() {
  const { activeConfig, chartType } = useSettingsStore()
  const { activeDataset, activeDatasetId, activeArrangement, getArrangement } = useDataPoint()

  const effectiveSwapTarget = computed(() => {
    const fromMap = getArrangement(activeDatasetId.value, chartType.value)
    if (fromMap) return fromMap
    const wire = (activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined)?.swap
    return wire || activeArrangement.value.targetString
  })

  const hasZOnChart = computed(() => arrangementHasChartZ(effectiveSwapTarget.value))

  const stack = computed<boolean>(
    () => (activeConfig.value as BarConfig | LineConfig | undefined)?.stack ?? false
  )

  const scale = computed<ScaleType>(() =>
    stack.value
      ? 'linear'
      : ((activeConfig.value as { scale?: ScaleType } | undefined)?.scale ?? 'linear')
  )

  const threeDRotate = computed<boolean>(
    () => (activeConfig.value as { threeDRotate?: boolean } | undefined)?.threeDRotate ?? false
  )

  const labelMode = computed<LabelMode>(() => resolveLabelMode(activeConfig.value))
  const showLabels = computed<boolean>(() => labelMode.value !== 'none')

  const sort = computed<Sort | undefined>(() => activeConfig.value?.sort)

  const threeD = computed<boolean>(
    () =>
      (activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined)?.threeD ?? false
  )

  const hasThreeDOption = computed<boolean>(() =>
    canOfferValue3D(
      chartType.value,
      activeDataset.value?.data,
      hasZOnChart.value,
      activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined,
      activeDataset.value?.axes
    )
  )

  const threeDVisualMap = computed<boolean>(
    () =>
      (activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined)?.threeDVisualMap ??
      false
  )

  const visualMap = computed<boolean>(
    () => (activeConfig.value as ScatterConfig | undefined)?.visualMap ?? false
  )

  const stat = computed<StatConfig | undefined>(() => activeConfig.value?.stat)

  const symbol = computed<string | undefined>(
    () => (activeConfig.value as LineConfig | ScatterConfig | undefined)?.symbol
  )

  const symbolSize = computed<number | undefined>(
    () => (activeConfig.value as LineConfig | ScatterConfig | undefined)?.symbolSize
  )

  const smooth = computed<boolean>(
    () => (activeConfig.value as LineConfig | undefined)?.smooth ?? false
  )

  const horizontal = computed<boolean>(
    () => (activeConfig.value as BarConfig | undefined)?.horizontal ?? false
  )

  return {
    scale,
    stack,
    threeDRotate,
    showLabels,
    labelMode,
    sort,
    threeD,
    hasThreeDOption,
    threeDVisualMap,
    visualMap,
    stat,
    symbol,
    symbolSize,
    smooth,
    horizontal,
  }
}
