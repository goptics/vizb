import type { ChartType, DataSet } from '../types'

/** Keep only settings whose chart type was bundled at HTML generation time. */
export function filterDataSetSettings(ds: DataSet, allowed?: ChartType[]): DataSet {
  if (!allowed?.length) return ds
  const settings = ds.settings?.filter((s) => allowed.includes(s.type)) ?? []
  return { ...ds, settings }
}