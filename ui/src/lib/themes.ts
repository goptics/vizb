import { computed, ref } from 'vue'
import { THEMES, THEME_VISUAL_MAP_COLORS, type ThemeName } from './themeCatalog'

export { THEMES, THEME_NAMES, THEME_VISUAL_MAP_COLORS, type ThemeName } from './themeCatalog'
const HEX_COLOR = /^#[0-9a-fA-F]{3}(?:[0-9a-fA-F]{3})?$/

export function isThemeName(value?: string): value is ThemeName {
  return !!value && Object.prototype.hasOwnProperty.call(THEMES, value.toLowerCase())
}

export function parseCustomPalette(value?: string): string[] | undefined {
  if (!value?.startsWith('#')) return undefined
  const colors = value.split(',').map((color) => color.trim())
  return colors.length >= 2 && colors.every((color) => HEX_COLOR.test(color)) ? colors : undefined
}

export function normalizeTheme(value?: string): string {
  const trimmed = value?.trim() || 'default'
  const name = trimmed.toLowerCase()
  if (isThemeName(name)) return name
  const custom = parseCustomPalette(trimmed)
  return custom?.join(',') ?? 'default'
}

export function resolvePalette(theme?: string): readonly string[] {
  const normalized = normalizeTheme(theme)
  return isThemeName(normalized) ? THEMES[normalized] : parseCustomPalette(normalized)!
}

export const activeThemeName = ref<string>('default')
export const activePalette = computed(() => resolvePalette(activeThemeName.value))

export function applyTheme(theme?: string) {
  activeThemeName.value = normalizeTheme(theme)
}

export function palettePrimary(palette: readonly string[] = activePalette.value): string {
  return palette[0]!
}

export function paletteGradientEndpoints(
  palette: readonly string[] = activePalette.value
): [string, string] {
  return [palette[0]!, palette[Math.min(4, palette.length - 1)]!]
}

export function resolveVisualMapColors(theme: string = activeThemeName.value): readonly string[] {
  const normalized = normalizeTheme(theme)
  return isThemeName(normalized)
    ? THEME_VISUAL_MAP_COLORS[normalized]
    : paletteGradientEndpoints(resolvePalette(normalized))
}
