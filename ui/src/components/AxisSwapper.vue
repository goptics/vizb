<script setup lang="ts">
import { computed, inject } from 'vue'
import SettingHeader from './SettingHeader.vue'
import SwapSelector from './Selector.vue'
import type { DataPoint } from '../types'
import { useDataPoint } from '../composables/useDataPoint'
import { resetColor } from '../lib/utils'
import { useSettingsStore } from '../composables/useSettingsStore'
import { translateAxisKey, swapAxisFields, swapAxisLabels } from '../lib/swap'
import type { TriggerSwap } from '../composables/useChartPipeline'

const { activeDataSet, activeGroupId, activeDataSetId } = useDataPoint()
const { setSelectedSwapIndex, getSelectedSwapIndex } = useSettingsStore()
const triggerSwap = inject<TriggerSwap>('triggerSwap', () => {})

// Canonical axis order; swap options are permutations of whichever are present.
const AXIS_ORDER = ['n', 'x', 'y', 'z'] as const

// All ordered length-`k` arrangements drawn from `pool`, deterministic.
// k = number of present values; pool = full axis set, so axes can be rotated
// in/out of `name` (e.g. a 3-axis dataset still offers nxy, nxz, ...).
const kPermutations = (pool: readonly string[], k: number): string[] => {
  if (k <= 0) return ['']
  const result: string[] = []
  pool.forEach((key, i) => {
    const rest = [...pool.slice(0, i), ...pool.slice(i + 1)]
    for (const perm of kPermutations(rest, k - 1)) result.push(key + perm)
  })
  return result
}

const presentKeys = (data: DataPoint[]): string[] => {
  const fieldFor = { n: 'name', x: 'xAxis', y: 'yAxis', z: 'zAxis' } as const
  return AXIS_ORDER.filter((k) => data.some((d) => d[fieldFor[k]]))
}


const swapOptions = computed(() => {
  const data = activeDataSet.value?.data
  if (!data || data.length === 0) return []
  // k = number of values; pool = full axis set. Selecting an arrangement that
  // omits z (e.g. nxy) renders 2D, while one using z (xyz) renders 3D.
  // z is only valid alongside both x and y (3D needs an x/y floor).
  return kPermutations(AXIS_ORDER, presentKeys(data).length)
    .filter((key) => !key.includes('z') || (key.includes('x') && key.includes('y')))
    // 1D data: drop the bare `n` arrangement — putting the lone value on `name`
    // gives one chart per point, which is useless. Offer only x / y placement.
    // (Multi-axis arrangements are never exactly 'n', so this is 1D-only.)
    .filter((key) => key !== 'n')
    .map((key) => ({ name: key }))
})

const getInitialSwapIndex = () => {
  const data = activeDataSet.value?.data
  if (!data || data.length === 0) return 0

  // Identity ordering (present keys in canonical order) = current data layout.
  const identity = presentKeys(data).join('')
  const index = swapOptions.value.findIndex((option) => option.name === identity)

  return index !== -1 ? index : 0
}

const selectedSwapIndex = computed(() => {
  const benchmarkId = activeDataSetId.value
  const stored = getSelectedSwapIndex(benchmarkId)

  if (stored !== undefined) {
    return stored
  }

  // Calculate and store if not found
  const index = getInitialSwapIndex()
  setSelectedSwapIndex(benchmarkId, index)
  return index
})

const handleSwapSelect = (index: number) => {
  const currentIndex = selectedSwapIndex.value
  if (index === currentIndex) return

  const currentOption = swapOptions.value[currentIndex]
  const targetOption = swapOptions.value[index]

  if (currentOption && targetOption && activeDataSet.value) {
    const newLabels = swapAxisLabels(currentOption.name, targetOption.name, activeDataSet.value.axisLabels)
    // Signal the pipeline BEFORE mutations so it can suppress the data-watcher
    // reinit that would otherwise fire from the axisLabels replacement below.
    triggerSwap(currentOption.name, targetOption.name, newLabels)
    // Keep main-thread store in sync (needed for future reinits after dataset/group
    // changes). swapAxisFields is the same O(n) rename the worker will also apply.
    swapAxisFields(activeDataSet.value.data, translateAxisKey(currentOption.name), translateAxisKey(targetOption.name))
    activeDataSet.value.axisLabels = newLabels
    resetColor()
    activeGroupId.value = 0
  }

  setSelectedSwapIndex(activeDataSetId.value, index)
}
</script>

<template>
  <div class="flex items-center justify-between">
    <SettingHeader label="Swap axis" description="Swap the axis of your data." />
    <SwapSelector
      :items="swapOptions"
      :activeId="selectedSwapIndex"
      @select="handleSwapSelect"
      class="w-28"
    />
  </div>
</template>
