<script setup lang="ts">
import { ref } from 'vue'
import { Settings } from 'lucide-vue-next'
import { Popover, PopoverTrigger, PopoverContent } from './ui'
import ChartSettings from './ChartSettings.vue'
import ScaleSelector from './ScaleSelector.vue'
import IconButton from './IconButton.vue'

defineProps<{
  hasYAxis?: boolean
}>()

const isOpen = ref(false)
</script>

<template>
  <Popover v-model:open="isOpen">
    <PopoverTrigger as-child>
      <IconButton aria-label="Open chart settings">
        <Settings class="h-5 w-5" />
      </IconButton>
    </PopoverTrigger>

    <PopoverContent>
      <ChartSettings>
        <template #scale>
          <ScaleSelector v-if="hasYAxis" />
        </template>
      </ChartSettings>
    </PopoverContent>
  </Popover>
</template>
