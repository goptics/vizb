import type { DataSet } from './types'

declare global {
  interface Window {
    VIZB_DATA: DataSet[]
    VIZB_VERSION: string
    VIZB_DATA_URL?: string
  }
}

export {}
