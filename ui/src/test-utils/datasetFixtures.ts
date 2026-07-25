import type { ChartConfig, Dataset } from '@/types'

/** Minimal Dataset for store / shape / URL router composable tests. */
export function ds(settings: ChartConfig[], data: Dataset['data'] = []): Dataset {
  return {
    name: 'test',
    settings,
    data,
  }
}
