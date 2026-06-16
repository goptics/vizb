<script setup lang="ts">
import { computed } from 'vue'
import SettingHeader from '../SettingHeader.vue'
import Selector from '../Selector.vue'
import { useDataPoint } from '../../composables/useDataPoint'
import type { DataPoint } from '../../types'

const { activeDataSet, activeArrangement } = useDataPoint()

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

const presentKeys = (data: DataPoint[] | undefined): string[] => {
  if (!data?.length) return []
  const fieldFor = { n: 'name', x: 'xAxis', y: 'yAxis', z: 'zAxis' } as const
  return AXIS_ORDER.filter((k) => data.some((d) => d[fieldFor[k]]))
}

const swapOptions = computed(() => {
  const data = activeDataSet.value?.data
  if (!data?.length) return []
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

const props = defineProps<{
  modelValue: string | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

// Index of the current arrangement in swapOptions. Falls back to the active
// arrangement's targetString (identity when nothing is set) when the wire-format
// swap doesn't match any option (e.g., stale value or 1D data with bare 'n').
const selectedSwapIndex = computed(() => {
  const target = props.modelValue ?? activeArrangement.value.targetString
  const idx = swapOptions.value.findIndex((opt) => opt.name === target)
  return idx !== -1 ? idx : 0
})

const handleSelect = (index: number) => {
  const opt = swapOptions.value[index]
  if (!opt) return
  emit('update:modelValue', opt.name)
}
</script>

<template>
  <div class="flex items-center justify-between">
    <SettingHeader label="Swap axis" description="Swap the axis of your data." />
    <Selector
      :items="swapOptions"
      :activeId="selectedSwapIndex"
      @select="handleSelect"
      class="w-28"
    />
  </div>
</template>
