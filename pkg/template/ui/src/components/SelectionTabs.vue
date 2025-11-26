<template>
  <Tabs 
    :model-value="modelValue" 
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <TabsList class="flex w-full">
      <TabsTrigger 
        v-for="option in options" 
        :key="option.value" 
        :value="option.value"
        class="flex-1"
      >
        <component :is="option.icon" v-if="option.icon" class="h-4 w-4" />
        <span class="ml-2">{{ option.label }}</span>
      </TabsTrigger>
    </TabsList>
  </Tabs>
</template>

<script setup lang="ts">
import { type Component } from 'vue'
import { Tabs, TabsList, TabsTrigger } from './ui'

interface Option {
  value: string | number
  label: string
  icon?: Component
}

defineProps<{
  modelValue: string | number
  options: Option[]
}>()

defineEmits<{
  (e: 'update:modelValue', value: string | number): void
}>()
</script>
