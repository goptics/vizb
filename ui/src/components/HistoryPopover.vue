<script setup lang="ts">
import type { Component } from 'vue'
import type { HistoryEntry } from '../types'
import Badge from './Badge.vue'
import Popover from './ui/Popover.vue'
import PopoverContent from './ui/PopoverContent.vue'
import PopoverTrigger from './ui/PopoverTrigger.vue'

defineProps<{
  icon: Component
  label: string
  value: string
  historyTitle: string
  entries: HistoryEntry[]
  hasHistory: boolean
  contentWidth?: string
}>()
</script>

<template>
  <Popover class="flex justify-center">
    <PopoverTrigger class="cursor-pointer">
      <Badge :icon="icon" :label="label" :value="value" />
    </PopoverTrigger>
    <PopoverContent v-if="hasHistory" :class="`${contentWidth ?? 'w-72'} p-3`" align="center">
      <p class="mb-2 text-sm font-medium">{{ historyTitle }}</p>
      <div class="space-y-1.5">
        <div
          v-for="entry in entries"
          :key="entry.tag"
          class="flex items-center justify-between gap-2 text-sm"
        >
          <span class="shrink-0 text-muted-foreground">{{ entry.tag }}</span>
          <slot name="entry" :entry="entry" />
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>
