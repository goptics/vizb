import { computed } from 'vue'
import type { ScaleType, Sort } from '../types'
import { useSettingsStore } from './useSettingsStore'

// Pure read-only composable: derives everything from the store's `activeConfig`
// (the typed per-chart config at `dataset.value.settings[activeChartIndex]`).
// Every field uses `?? default` — no `cfg.type === 'bar' || ...` branching.
// For pie/heatmap/radar the optional `scale` / `autoRotate` fields don't exist
// on the type at all, so `?? default` is the only mechanism that produces a
// value (TypeScript's optional-field semantics in place of a runtime guard).
export function useActiveChartShape() {
  const { activeConfig } = useSettingsStore()

  const scale = computed<ScaleType>(
    () => (activeConfig.value as { scale?: ScaleType } | undefined)?.scale ?? 'linear'
  )

  const autoRotate = computed<boolean>(
    () => (activeConfig.value as { autoRotate?: boolean } | undefined)?.autoRotate ?? false
  )

  const showLabels = computed<boolean>(() => activeConfig.value?.showLabels ?? false)

  const sort = computed<Sort | undefined>(() => activeConfig.value?.sort)

  return { scale, autoRotate, showLabels, sort }
}
