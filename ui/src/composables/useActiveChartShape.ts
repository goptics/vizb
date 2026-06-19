import { computed } from 'vue'
import type { BarConfig, LineConfig, ScaleType, Sort, StatConfig } from '../types'
import { useSettingsStore } from './useSettingsStore'

// Pure read-only composable: derives everything from the store's `activeConfig`
// (the typed per-chart config at `dataset.value.settings[activeChartIndex]`).
// Every field uses `?? default` — no `cfg.type === 'bar' || ...` branching.
// For pie/heatmap/radar the optional `scale` / `threeDRotate` fields don't exist
// on the type at all, so `?? default` is the only mechanism that produces a
// value (TypeScript's optional-field semantics in place of a runtime guard).
export function useActiveChartShape() {
  const { activeConfig } = useSettingsStore()

  const scale = computed<ScaleType>(
    () => (activeConfig.value as { scale?: ScaleType } | undefined)?.scale ?? 'linear'
  )

  const threeDRotate = computed<boolean>(
    () => (activeConfig.value as { threeDRotate?: boolean } | undefined)?.threeDRotate ?? false
  )

  const showLabels = computed<boolean>(() => activeConfig.value?.showLabels ?? false)

  const sort = computed<Sort | undefined>(() => activeConfig.value?.sort)

  const threeD = computed<boolean>(
    () => (activeConfig.value as BarConfig | LineConfig | undefined)?.threeD ?? false
  )

  const hasThreeDOption = computed<boolean>(() => {
    const cfg = activeConfig.value as BarConfig | LineConfig | undefined
    return cfg?.threeD !== undefined
  })

  const threeDVisualMap = computed<boolean>(
    () => (activeConfig.value as BarConfig | LineConfig | undefined)?.threeDVisualMap ?? false
  )

  const stat = computed<StatConfig | undefined>(() => activeConfig.value?.stat)

  return { scale, threeDRotate, showLabels, sort, threeD, hasThreeDOption, threeDVisualMap, stat }
}
