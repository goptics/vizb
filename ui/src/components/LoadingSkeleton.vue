<template>
  <div class="mx-auto min-h-screen max-w-7xl animate-pulse px-4 py-8 sm:px-6 lg:px-8">
    <header class="space-y-3 py-5 text-center">
      <div class="mx-auto h-9 w-64 rounded-md bg-muted"></div>
      <div class="flex justify-center">
        <div class="h-6 w-28 rounded-full bg-muted"></div>
      </div>
      <div class="flex justify-center">
        <div class="h-6 w-24 rounded-full bg-muted"></div>
      </div>
      <div class="mx-auto h-4 w-48 rounded bg-muted"></div>
    </header>
    <div class="space-y-5">
      <div
        v-for="i in 3"
        :key="i"
        class="rounded-lg border border-border bg-card p-6 shadow-sm"
      >
        <div class="mb-4 h-5 w-40 rounded bg-muted"></div>
        <div class="h-[500px] rounded bg-muted"></div>
      </div>
    </div>
  </div>
</template>
