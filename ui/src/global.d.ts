import type { Benchmark } from './types'

declare global {
  interface Window {
    VIZB_DATA: Benchmark[]
    VIZB_VERSION: string
    VIZB_DATA_URL?: string
  }
}

export {}
