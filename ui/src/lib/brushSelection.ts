import type { EChartsOption } from 'echarts'

export type BrushSelectionStats = {
  regions: number
  total: number
  average: number
  count: number
}

type BrushSelectedEvent = {
  batch?: {
    areas?: unknown[]
    selected?: { seriesIndex: number; dataIndex: number[] }[]
  }[]
}

const selectedValue = (data: unknown): number | null => {
  if (data && typeof data === 'object' && !Array.isArray(data) && 'value' in data) {
    data = (data as { value: unknown }).value
  }
  const value = Array.isArray(data) ? data[1] : data
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

export function brushSelectionStats(
  option: EChartsOption,
  event: BrushSelectedEvent
): BrushSelectionStats {
  const series = Array.isArray(option.series) ? option.series : option.series ? [option.series] : []
  const seen = new Set<string>()
  let regions = 0
  let total = 0
  let count = 0

  for (const batch of event.batch ?? []) {
    regions += batch.areas?.length ?? 0
    for (const selected of batch.selected ?? []) {
      const data = (series[selected.seriesIndex] as { data?: unknown[] } | undefined)?.data
      for (const dataIndex of selected.dataIndex) {
        const key = `${selected.seriesIndex}:${dataIndex}`
        if (seen.has(key)) continue
        seen.add(key)
        const value = selectedValue(data?.[dataIndex])
        if (value === null) continue
        total += value
        count++
      }
    }
  }

  return { regions, total, average: count ? total / count : 0, count }
}
