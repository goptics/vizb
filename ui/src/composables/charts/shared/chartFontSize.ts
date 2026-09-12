import { shallowRef } from 'vue'
import { resolveFontSize, type FontSizeSpec, type ResolvedFontSize } from '@/lib/fontSize'

/** Baked appearance.fontSize for the active dataset (set by useDataPoint). */
export const appearanceFontSize = shallowRef<number | FontSizeSpec | undefined>()

/** Resolved series/legend/label sizes from the active dataset appearance. */
export function chartFontSize(): ResolvedFontSize {
  return resolveFontSize(appearanceFontSize.value)
}
