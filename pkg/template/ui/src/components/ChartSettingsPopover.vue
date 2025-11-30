<script setup lang="ts">
import { ref } from 'vue'
import { Settings } from 'lucide-vue-next'
import { Popover, PopoverTrigger, PopoverContent } from './ui'
import ChartSettings from './ChartSettings.vue'
import IconButton from './IconButton.vue'

const isOpen = ref(false)
</script>

<template>
  <Popover v-model:open="isOpen">
    <PopoverTrigger as-child>
      <IconButton aria-label="Open chart settings">
        <Settings class="h-5 w-5" />
      </IconButton>
    </PopoverTrigger>

    <!-- Popover content rendering only the card (no outer chrome) -->
    <PopoverContent>
      <ChartSettings />
    </PopoverContent>
  </Popover>
</template>
