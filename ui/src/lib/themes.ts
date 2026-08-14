import { computed, ref } from 'vue'
import type { Dataset, Theme } from '../types'

/** Sole UI built-in theme. Dataset-owned themes supply any additional palettes. */
export const DEFAULT_THEME = {
  name: 'default',
  colors: [
    '#5470C6',
    '#3BA272',
    '#FC8452',
    '#73C0DE',
    '#EE6666',
    '#FAC858',
    '#9A60B4',
    '#EA7CCC',
    '#91CC75',
    '#FF9F7F',
  ],
  visualMapColors: ['#91CC75', '#EE6666'],
} as const satisfies Theme

const HEX_COLOR = /^#[0-9a-fA-F]{3}(?:[0-9a-fA-F]{3})?$/

/** Extra themes from the active dataset, keyed by lowercased name (excludes default). */
const datasetThemesByName = ref(new Map<string, Theme>())

function cloneTheme(theme: Theme): Theme {
  return {
    name: theme.name,
    colors: [...theme.colors],
    visualMapColors: [...(theme.visualMapColors ?? [])],
  }
}

// Match CLI/Go visualMapFromColors: first + last of the palette.
function gradientEndpoints(palette: readonly string[]): [string, string] {
  return [palette[0]!, palette[palette.length - 1]!]
}

/**
 * Register dataset-owned themes so name-based resolution can find their colors.
 * UI always owns `default`; dataset entries named default are ignored.
 * A colors-only themes[0] with a blank name is registered as `custom` so it
 * matches resolveActiveTheme's fallback name.
 */
export function registerDatasetThemes(themes?: Theme[]) {
  const map = new Map<string, Theme>()
  for (let i = 0; i < (themes ?? []).length; i++) {
    const theme = themes![i]!
    let name = theme?.name?.trim()
    if (!name) {
      if (i === 0 && theme?.colors?.length) name = 'custom'
      else continue
    }
    const key = name.toLowerCase()
    if (key === DEFAULT_THEME.name) continue
    if (!theme.colors?.length) continue
    map.set(key, cloneTheme({ ...theme, name }))
  }
  datasetThemesByName.value = map
}

/** Count of embedded non-default dataset themes (author-provided). */
export function authorThemeCount(): number {
  return datasetThemesByName.value.size
}

/**
 * Themes the viewer may select / that localStorage may apply for this report.
 * - 0 author themes → only UI default
 * - 1 author theme → only that theme (no selector; localStorage cannot switch away)
 * - 2+ author themes → default + registered dataset themes
 */
export function listAvailableThemes(): Theme[] {
  const registered = [...datasetThemesByName.value.values()].map(cloneTheme)
  if (registered.length >= 2) {
    return [cloneTheme(DEFAULT_THEME), ...registered]
  }
  if (registered.length === 1) {
    return registered
  }
  return [cloneTheme(DEFAULT_THEME)]
}

export function listAvailableThemeNames(): string[] {
  return listAvailableThemes().map((theme) => theme.name)
}

/** True when the author embedded 2+ themes (selector is shown). */
export function shouldShowThemeSelector(): boolean {
  return authorThemeCount() >= 2
}

/** True when `value` is in the available set for the current dataset. */
export function isAvailableThemeName(value?: string): boolean {
  if (!value) return false
  const key = value.trim().toLowerCase()
  if (!key) return false
  return listAvailableThemeNames().some((name) => name.toLowerCase() === key)
}

export function findTheme(name?: string): Theme | undefined {
  if (!name) return undefined
  const key = name.trim().toLowerCase()
  if (!key) return undefined
  if (key === DEFAULT_THEME.name) return cloneTheme(DEFAULT_THEME)
  const fromDataset = datasetThemesByName.value.get(key)
  return fromDataset ? cloneTheme(fromDataset) : undefined
}

export function parseCustomPalette(value?: string): string[] | undefined {
  if (!value?.startsWith('#')) return undefined
  const colors = value.split(',').map((color) => color.trim())
  return colors.length >= 2 && colors.every((color) => HEX_COLOR.test(color)) ? colors : undefined
}

function themeFromCustomPalette(colors: string[]): Theme {
  return {
    name: 'custom',
    colors: [...colors],
    visualMapColors: gradientEndpoints(colors),
  }
}

/**
 * Resolve the chart-active theme for a dataset: themes[0] when present,
 * else soft-handle legacy theme hex string, else UI default.
 * Named legacy themes without a themes[] catalog fall back to default
 * (UI no longer ships the 13-theme catalog).
 */
export function resolveActiveTheme(dataset?: Pick<Dataset, 'themes' | 'theme'> | null): Theme {
  const first = dataset?.themes?.[0]
  if (first?.colors?.length) {
    return cloneTheme({
      name: first.name?.trim() || 'custom',
      colors: first.colors,
      visualMapColors: first.visualMapColors?.length
        ? first.visualMapColors
        : gradientEndpoints(first.colors),
    })
  }

  const legacy = dataset?.theme?.trim()
  if (legacy?.startsWith('#')) {
    const colors = parseCustomPalette(legacy)
    if (colors) return themeFromCustomPalette(colors)
  }

  return cloneTheme(DEFAULT_THEME)
}

export function isThemeName(value?: string): boolean {
  if (!value) return false
  const key = value.trim().toLowerCase()
  return key === DEFAULT_THEME.name || datasetThemesByName.value.has(key)
}

export function normalizeTheme(value?: string): string {
  const trimmed = value?.trim() || DEFAULT_THEME.name
  const name = trimmed.toLowerCase()
  if (name === DEFAULT_THEME.name) return name
  const registered = datasetThemesByName.value.get(name)
  if (registered) return registered.name
  const custom = parseCustomPalette(trimmed)
  return custom?.join(',') ?? DEFAULT_THEME.name
}

export function resolveTheme(theme?: string): Theme {
  const normalized = normalizeTheme(theme)
  if (normalized.startsWith('#')) {
    return themeFromCustomPalette(parseCustomPalette(normalized)!)
  }
  // normalizeTheme only returns default or a registered dataset name here.
  return findTheme(normalized)!
}

export const activeThemeName = ref<string>(DEFAULT_THEME.name)
export const activeTheme = computed(() => resolveTheme(activeThemeName.value))
export const activePalette = computed(() => activeTheme.value.colors)

export function applyTheme(theme?: string) {
  activeThemeName.value = normalizeTheme(theme)
}

export function palettePrimary(palette: readonly string[] = activePalette.value): string {
  return palette[0]!
}

export function resolveVisualMapColors(theme: string = activeThemeName.value): readonly string[] {
  const resolved = resolveTheme(theme)
  const visualMap = resolved.visualMapColors
  if (visualMap && visualMap.length >= 2) {
    return visualMap
  }
  return gradientEndpoints(resolved.colors)
}
