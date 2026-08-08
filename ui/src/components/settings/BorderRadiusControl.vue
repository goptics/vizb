<template>
  <div class="flex flex-col gap-2">
    <div class="flex items-center justify-between">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        Border Radius
      </label>
      <span class="text-sm text-gray-500 dark:text-gray-400">
        {{ displayValue }}px
      </span>
    </div>
    <div class="flex items-center gap-3">
      <input
        type="range"
        :min="0"
        :max="50"
        :step="1"
        :value="modelValue ?? 0"
        @input="updateValue(Number(($event.target as HTMLInputElement).value))"
        class="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer dark:bg-gray-700"
      />
      <input
        type="number"
        :min="0"
        :max="50"
        :step="1"
        :value="modelValue ?? 0"
        @input="updateValue(Number(($event.target as HTMLInputElement).value))"
        class="w-16 px-2 py-1 text-sm border rounded-md dark:bg-gray-800 dark:border-gray-700 dark:text-white"
      />
    </div>
    <p class="text-xs text-gray-500 dark:text-gray-400">
      Round bar corners (0 = square, max 50px). For stacked bars, only the top segment is rounded.
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  modelValue: number | undefined
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number | undefined): void
}>()

const displayValue = computed(() => props.modelValue ?? 0)

/**
 * Updates the borderRadius value. Rejects non-integer and negative values.
 * Clamps values to 50px max.
 */

function updateValue(value: number) {
  // Reject non-integer values (CLI expects integer)
  if (!Number.isInteger(value) || value < 0) {
    emit('update:modelValue', undefined)
    return
  }
  emit('update:modelValue', Math.min(value, 50))
}
</script>