<script setup lang="ts">
import { computed } from 'vue'
import SettingHeader from './SettingHeader.vue'
import SwapSelector from './Selector.vue'
import type { BenchmarkData } from '../types'
import { useBenchmarkData } from '../composables/useBenchmarkData'
import { resetColor } from '../lib/utils'
import { useSettingsStore } from '../composables/useSettingsStore'

const { activeBenchmark, activeGroupId, activeBenchmarkId } = useBenchmarkData()
const { setSelectedSwapIndex, getSelectedSwapIndex } = useSettingsStore()

// Canonical axis order; swap options are permutations of whichever are present.
const AXIS_ORDER = ['n', 'x', 'y', 'z'] as const

type AxisKey = 'name' | 'xAxis' | 'yAxis' | 'zAxis'

const translateAxisKey = (key: string): AxisKey[] => {
  const keyMap = {
    x: 'xAxis',
    y: 'yAxis',
    n: 'name',
    z: 'zAxis',
  }
  return key.split('').map((k) => keyMap[k as keyof typeof keyMap]) as AxisKey[]
}

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

const presentKeys = (data: BenchmarkData[]): string[] => {
  const fieldFor = { n: 'name', x: 'xAxis', y: 'yAxis', z: 'zAxis' } as const
  return AXIS_ORDER.filter((k) => data.some((d) => d[fieldFor[k]]))
}

const swapAxis = (currentKey: string, targetKey: string, data: BenchmarkData[]) => {
  const currentKeys = translateAxisKey(currentKey)
  const targetKeys = translateAxisKey(targetKey)

  if (currentKeys.length !== targetKeys.length) return

  for (const bench of data) {
    const values = currentKeys.map((k) => bench[k])

    for (const k of currentKeys) {
      delete bench[k]
    }

    for (const k of targetKeys) {
      bench[k] = values.shift()
    }
  }
}

const swapOptions = computed(() => {
  const data = activeBenchmark.value?.data
  if (!data || data.length === 0) return []
  // k = number of values; pool = full axis set. Selecting an arrangement that
  // omits z (e.g. nxy) renders 2D, while one using z (xyz) renders 3D.
  // z is only valid alongside both x and y (3D needs an x/y floor).
  return kPermutations(AXIS_ORDER, presentKeys(data).length)
    .filter((key) => !key.includes('z') || (key.includes('x') && key.includes('y')))
    .map((key) => ({ name: key }))
})

const getInitialSwapIndex = () => {
  const data = activeBenchmark.value?.data
  if (!data || data.length === 0) return 0

  // Identity ordering (present keys in canonical order) = current data layout.
  const identity = presentKeys(data).join('')
  const index = swapOptions.value.findIndex((option) => option.name === identity)

  return index !== -1 ? index : 0
}

const selectedSwapIndex = computed(() => {
  const benchmarkId = activeBenchmarkId.value
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

  if (currentOption && targetOption && activeBenchmark.value) {
    swapAxis(currentOption.name, targetOption.name, activeBenchmark.value.data)
    resetColor()
    activeGroupId.value = 0
  }

  setSelectedSwapIndex(activeBenchmarkId.value, index)
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
