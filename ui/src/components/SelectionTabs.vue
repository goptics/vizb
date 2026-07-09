<script setup lang="ts">
import { type Component } from 'vue'
import { Tabs, TabsList, TabsTrigger } from './ui'

interface Option {
  value: string | number
  label: string
  icon?: Component
}

const props = defineProps<{
  modelValue: string | number
  options: Option[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string | number): void
}>()

const onUpdate = (value: string | number) => {
  if (props.disabled) return
  emit('update:modelValue', value)
}
</script>

<template>
  <Tabs :model-value="props.modelValue" @update:model-value="onUpdate">
    <TabsList class="flex w-full">
      <TabsTrigger
        v-for="option in props.options"
        :key="option.value"
        :value="option.value"
        :disabled="props.disabled"
        class="flex-1"
      >
        <component :is="option.icon" v-if="option.icon" class="h-4 w-4" />
        <span class="ml-2">{{ option.label }}</span>
      </TabsTrigger>
    </TabsList>
  </Tabs>
</template>
