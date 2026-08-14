import type { ChartData } from '@/types'
import type { ChartBuilder } from './types'

// Value-mode chart shape: continuous numeric axes (--axes x,y[,z]).
// Chart materialisation lives in buildValueModeChart (lib/transform.ts).
export class ValueBuilder implements ChartBuilder {
  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number {
    if (chart.valuePoints3D?.length) {
      const idx = axis === 'x' ? 0 : axis === 'y' ? 1 : 2
      return new Set(chart.valuePoints3D.map((p) => p[idx])).size
    }
    if (chart.valueTuples?.length) {
      if (axis === 'z') return 0
      const idx = axis === 'x' ? 0 : 1
      return new Set(chart.valueTuples.map((p) => p[idx])).size
    }
    return 0
  }

  grandTotal(chart: ChartData): number {
    if (chart.valuePoints3D?.length) {
      return chart.valuePoints3D.reduce((sum, [, , z]) => sum + z, 0)
    }
    if (chart.valueTuples?.length) {
      return chart.valueTuples.reduce((sum, [, y]) => sum + y, 0)
    }
    return 0
  }

  is3D(chart: ChartData): boolean {
    return chart.render3D?.mode === 'continuous'
  }

  canOfferValue3D(): boolean {
    return false
  }
}
