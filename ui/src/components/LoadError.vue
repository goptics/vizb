<script setup lang="ts">
withDefaults(defineProps<{ message: string; inline?: boolean; retry?: () => void }>(), {
  inline: false,
  retry: undefined,
})
</script>

<template>
  <div
    class="flex items-center justify-center px-4 text-center"
    :class="inline ? 'min-h-[32rem]' : 'min-h-screen'"
  >
    <div class="max-w-md space-y-2">
      <p class="font-medium text-destructive">Failed to load Data set</p>
      <p class="break-all text-sm text-muted-foreground">{{ message }}</p>
      <p class="text-xs text-muted-foreground/60">
        Ensure the data URL is reachable and the server sends
        <code>Access-Control-Allow-Origin: *</code>.
      </p>
      <button
        v-if="retry"
        type="button"
        class="mt-3 inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        @click="retry"
      >
        Retry
      </button>
    </div>
  </div>
</template>
