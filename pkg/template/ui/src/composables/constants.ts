import type { Settings } from '../types'

export const DEFAULT_SETTINGS: Settings = {
  sort: {
    enabled: false,
    order: 'asc',
  },
  showLabels: false,
  charts: ['bar', 'line', 'pie'],
}
