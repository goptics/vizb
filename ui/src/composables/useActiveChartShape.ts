import { computed } from 'vue'
import type { BarConfig, LineConfig, ScatterConfig, ScaleType, Sort, StatConfig } from '../types'
import { arrangementHasChartZ } from '../lib/swap'
import { canOfferValue3D } from '../lib/utils'
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
  const { activeDataSet, activeDataSetId, activeArrangement, getArrangement } = useDataPoint()

  const effectiveSwapTarget = computed(() => {
    const fromMap = getArrangement(activeDataSetId.value, chartType.value)
    if (fromMap) return fromMap
    const wire = (activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined)?.swap
    return wire || activeArrangement.value.targetString
  })

  const hasZOnChart = computed(() => arrangementHasChartZ(effectiveSwapTarget.value))

  const scale = computed<ScaleType>(
    () => (activeConfig.value as { scale?: ScaleType } | undefined)?.scale ?? 'linear'
  )

  const threeDRotate = computed<boolean>(
    () => (activeConfig.value as { threeDRotate?: boolean } | undefined)?.threeDRotate ?? false
  )

  const showLabels = computed<boolean>(() => activeConfig.value?.showLabels ?? false)

  const sort = computed<Sort | undefined>(() => activeConfig.value?.sort)

  const threeD = computed<boolean>(
    () =>
      (activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined)?.threeD ?? false
  )

  const hasThreeDOption = computed<boolean>(() =>
    canOfferValue3D(
      chartType.value,
      activeDataSet.value?.data,
      hasZOnChart.value,
      activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined,
      activeDataSet.value?.axes
    )
  )

  const threeDVisualMap = computed<boolean>(
    () =>
      (activeConfig.value as BarConfig | LineConfig | ScatterConfig | undefined)?.threeDVisualMap ??
      false
  )

  const stat = computed<StatConfig | undefined>(() => activeConfig.value?.stat)

  return { scale, threeDRotate, showLabels, sort, threeD, hasThreeDOption, threeDVisualMap, stat }
}
