import type { Benchmark } from './types/benchmark'

declare global {
  interface Window {
    VIZB_DATA: Benchmark[]
    VIZB_VERSION: string
  }
}

export {}
