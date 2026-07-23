import type { ChartBuilder } from './types'
import { GroupedBuilder } from './grouped'
import { PreserveRowsBuilder } from './preserveRows'
import { ValueBuilder } from './value'
import { MixedBuilder } from './mixed'
import type { ChartData } from '@/types'

const grouped = new GroupedBuilder()
const preserveRows = new PreserveRowsBuilder()
const value = new ValueBuilder()
const mixed = new MixedBuilder()

/** Pick the builder for a chart from its rendered data shape. */
export function builderForChart(chart: ChartData): ChartBuilder {
  if (chart.series.length || chart.points.length) return grouped
  if (chart.valueTuples?.length || chart.valuePoints3D?.length) return value
  if (chart.mixedTuples?.length || chart.xCategories?.length) return mixed
  if (chart.render3D?.mode === 'continuous') return value
  if (chart.render3D?.mode === 'mixed') return mixed
  if (chart.statType === 'value') return value
  if (chart.statType === 'mixed') return mixed
  if (chart.statType === 'preserveRows') return preserveRows
  return grouped
}

/** Pick the builder for the build phase based on context flags. */
export function pickBuilder(ctx: {
  preserveRows?: boolean
  mixedMode?: boolean
  valueMode?: boolean
}): ChartBuilder {
  if (ctx.valueMode) return value
  if (ctx.mixedMode) return mixed
  if (ctx.preserveRows) return preserveRows
  return grouped
}

export { grouped, preserveRows, value, mixed }
