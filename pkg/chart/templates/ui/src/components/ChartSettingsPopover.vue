<template>
  <Popover v-model:open="isOpen">
    <PopoverTrigger as-child>
      <button
        class="inline-flex items-center justify-center w-12 h-12 rounded-lg border border-border bg-card text-card-foreground hover:bg-accent hover:text-accent-foreground transition-colors shadow-sm"
        aria-label="Open chart settings"
      >
        <Settings class="w-5 h-5" />
      </button>
    </PopoverTrigger>

    <!-- Popover content rendering only the card (no outer chrome) -->
    <PopoverContent class="w-[380px] p-0">
      <ChartSettings />
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { Settings } from "lucide-vue-next";
import { Popover, PopoverTrigger, PopoverContent } from "./ui";
import ChartSettings from "./ChartSettings.vue";

const isOpen = ref(false);
</script>
