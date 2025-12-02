<script setup lang="ts">
import { computed } from 'vue'
import SettingHeader from './SettingHeader.vue'
import SwapSelector from './Selector.vue'
import type { BenchmarkData } from '../types/benchmark'
import { useBenchmarkData } from '../composables/useBenchmarkData'
import { resetColor } from '../lib/utils'
import { useSettingsStore } from '../composables/useSettingsStore'

const { activeBenchmarkDimension, activeBenchmark, activeGroupId, activeBenchmarkId } =
  useBenchmarkData()
const { setSelectedSwapIndex, getSelectedSwapIndex } = useSettingsStore()

const dimensionMap = {
  1: ['x', 'y'],
  2: ['nx', 'ny', 'xy', 'yx', 'yn', 'xn'],
  3: ['nxy', 'nyx', 'xny', 'xyn', 'ynx', 'yxn'],
}

type AxisKey = 'name' | 'xAxis' | 'yAxis'

const translateAxisKey = (key: string): AxisKey[] => {
  const keyMap = {
    x: 'xAxis',
    y: 'yAxis',
    n: 'name',
  }
  return key.split('').map((k) => keyMap[k as keyof typeof keyMap]) as AxisKey[]
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
  const dimensions = dimensionMap[activeBenchmarkDimension.value as keyof typeof dimensionMap] || []
  return dimensions.map((key) => ({ name: key }))
})

const getInitialSwapIndex = () => {
  const data = activeBenchmark.value?.data
  if (!data || data.length === 0) return 0

  const hasName = data.some((d) => d.name)
  const hasX = data.some((d) => d.xAxis)
  const hasY = data.some((d) => d.yAxis)

  const presentKeys = new Set<string>()
  if (hasName) presentKeys.add('n')
  if (hasX) presentKeys.add('x')
  if (hasY) presentKeys.add('y')

  const index = swapOptions.value.findIndex((option) => {
    if (option.name.length !== presentKeys.size) return false
    const optionKeys = option.name.split('')
    return optionKeys.every((k) => presentKeys.has(k))
  })

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
