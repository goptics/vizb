import type { Theme } from '../types'

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

/** @deprecated Prefer DEFAULT_THEME; kept for call sites that only need the color list. */
export const THEMES = {
  default: DEFAULT_THEME.colors,
} as const

export type ThemeName = 'default'
export const THEME_NAMES: ThemeName[] = ['default']
