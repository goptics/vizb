<script setup lang="ts">
import { computed } from 'vue'
import { Cpu } from 'lucide-vue-next'
import type { HistoryEntry } from '../types'
import Badge from './Badge.vue'
import Popover from './ui/Popover.vue'
import PopoverContent from './ui/PopoverContent.vue'
import PopoverTrigger from './ui/PopoverTrigger.vue'
import { CPUtoString } from '../lib/utils'

const props = defineProps<{
  cpu?: { name?: string; cores?: number }
  history?: HistoryEntry[]
}>()

const cpuString = computed(() => CPUtoString(props.cpu))

const sortedHistory = computed(() => {
  if (!props.history?.length) return []
  return [...props.history]
    .filter((e) => e.cpu?.name || e.cpu?.cores)
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
})

const hasHistory = computed(() => sortedHistory.value.length > 0)
</script>

<template>
  <Popover v-if="cpu" class="flex justify-center">
    <PopoverTrigger class="cursor-pointer">
      <Badge :icon="Cpu" label="CPU" :value="cpuString" />
    </PopoverTrigger>
    <PopoverContent v-if="hasHistory" class="w-80 p-3" align="center">
      <p class="mb-2 text-sm font-medium">CPU History</p>
      <div class="space-y-1.5">
        <div
          v-for="entry in sortedHistory"
          :key="entry.tag"
          class="flex items-center justify-between gap-2 text-sm"
        >
          <span class="shrink-0 text-muted-foreground">{{ entry.tag }}</span>
          <span class="min-w-0 truncate text-right tabular-nums">{{ CPUtoString(entry.cpu) }}</span>
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>
