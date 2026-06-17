<script setup lang="ts">
import { ref } from 'vue'
import { Settings } from 'lucide-vue-next'
import { Popover, PopoverTrigger, PopoverContent } from './ui'
import SettingsPanel from './SettingsPanel.vue'
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

    <PopoverContent>
      <SettingsPanel />
    </PopoverContent>
  </Popover>
</template>
