import type { ChartType, Dataset } from '../types'

/** Keep only settings whose chart type was bundled at HTML generation time. */
export function filterDatasetSettings(ds: Dataset, allowed?: ChartType[]): Dataset {
  if (!allowed?.length) return ds
  const settings = ds.settings?.filter((s) => allowed.includes(s.type)) ?? []
  return { ...ds, settings }
}
