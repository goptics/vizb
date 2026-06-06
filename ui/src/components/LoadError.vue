<script setup lang="ts">
defineProps<{ message: string }>()
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-4 text-center">
    <div class="max-w-md space-y-2">
      <p class="font-medium text-destructive">Failed to load benchmark data</p>
      <p class="break-all text-sm text-muted-foreground">{{ message }}</p>
      <p class="text-xs text-muted-foreground/60">
        Ensure the data URL is reachable and the server sends
        <code>Access-Control-Allow-Origin: *</code>.
      </p>
    </div>
  </div>
</template>
