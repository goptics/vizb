<script setup lang="ts">
import { computed } from 'vue'
import { Separator } from '../ui'

const props = defineProps<{
  modelValue: number | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number | undefined): void
}>()

const displayValue = computed(() => props.modelValue ?? 0)

function onInput(raw: string) {
  if (raw === '') {
    emit('update:modelValue', undefined)
    return
  }
  const n = Number(raw)
  if (!Number.isInteger(n) || n < 0) return
  emit('update:modelValue', n === 0 ? undefined : n)
}
</script>

<template>
  <div class="flex flex-col gap-2">
    <div class="flex items-center justify-between gap-3">
      <label for="border-radius-input" class="text-sm font-medium text-gray-700 dark:text-gray-300">
        Border radius
      </label>
      <div class="flex items-center gap-1">
        <input
          id="border-radius-input"
          type="number"
          min="0"
          step="1"
          :value="displayValue"
          class="w-16 rounded-md border px-2 py-1 text-sm dark:border-gray-700 dark:bg-gray-800 dark:text-white"
          @input="onInput(($event.target as HTMLInputElement).value)"
        />
        <span class="text-sm text-gray-500 dark:text-gray-400">px</span>
      </div>
    </div>
    <p class="text-xs text-gray-500 dark:text-gray-400">
      Round free outer corners. Stacked bars round only the top segment.
    </p>
  </div>
  <Separator />
</template>
