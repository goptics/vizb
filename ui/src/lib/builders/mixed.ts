import type { ChartData } from '@/types'
import type { ChartBuilder } from './types'

// Mixed-axis scatter chart shape: category x + value y[,z] (solo --select mixed
// mode). Chart materialisation lives in buildMixedModeChart (lib/transform.ts).
export class MixedBuilder implements ChartBuilder {
  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number {
    if (chart.mixedTuples?.length && chart.xCategories?.length) {
      if (axis === 'x') return chart.xCategories.length
      if (axis === 'z') return 0
      return chart.mixedTuples.length
    }
    if (axis === 'x') return chart.xCategories!.length
    if (axis === 'z') return 0
    const pts = chart.render3D?.lineSeries[0]?.data ?? []
    return new Set(pts.map((p) => p.value[1])).size
  }

  grandTotal(chart: ChartData): number {
    if (chart.mixedTuples?.length) {
      return chart.mixedTuples.reduce((sum, [, y]) => sum + y, 0)
    }
    const pts = chart.render3D?.lineSeries[0]?.data ?? []
    return pts.reduce((sum, p) => sum + p.value[1]!, 0)
  }

  is3D(chart: ChartData): boolean {
    return chart.render3D?.mode === 'mixed'
  }

  canOfferValue3D(): boolean {
    return false
  }
}
