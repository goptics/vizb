import type { Axis } from '@/types'
import { valueModeSwapEnabled } from '@/lib/utils'
import type { RenderableField } from './fieldRegistry'

/** Scatter value/hybrid datasets: hide sort; swap only for pure value mode. */
export const filterScatterDatasetFields = (
  fields: RenderableField[],
  axes: Axis[] | undefined
): RenderableField[] =>
  fields.filter((f) => {
    if (f.key === 'sort') return false
    if (f.key === 'swap') return valueModeSwapEnabled(axes)
    return true
  })
