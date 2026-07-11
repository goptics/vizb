import { describe, expect, it } from 'vitest'
import {
  normalizeTheme,
  paletteGradientEndpoints,
  parseCustomPalette,
  resolvePalette,
  resolveVisualMapColors,
  THEMES,
  THEME_NAMES,
} from './themes'

describe('themes', () => {
  it('ships the complete catalog with ten unique colors per built-in', () => {
    expect(THEME_NAMES).toEqual([
      'default',
      'vintage',
      'meadow',
      'westeros',
      'essos',
      'wonderland',
      'walden',
      'chalk',
      'infographic',
      'macarons',
      'roma',
      'shine',
      'purple-passion',
    ])
    for (const palette of Object.values(THEMES)) {
      expect(palette).toHaveLength(10)
      expect(new Set(palette.map((color) => color.toLowerCase())).size).toBe(10)
    }
  })

  it('resolves names case-insensitively and defaults invalid values', () => {
    expect(resolvePalette('VINTAGE')).toBe(THEMES.vintage)
    expect(resolvePalette('missing')).toBe(THEMES.default)
    expect(normalizeTheme()).toBe('default')
  })

  it('accepts flexible custom #rgb and #rrggbb palettes', () => {
    expect(parseCustomPalette('#f00, #00ff00,#00f')).toEqual(['#f00', '#00ff00', '#00f'])
    expect(resolvePalette('#f00,#0f0')).toEqual(['#f00', '#0f0'])
    expect(parseCustomPalette('#f00')).toBeUndefined()
    expect(parseCustomPalette('#ggg,#000')).toBeUndefined()
  })

  it('uses the last available color for short-palette gradients', () => {
    expect(paletteGradientEndpoints(['#111', '#222'])).toEqual(['#111', '#222'])
    expect(paletteGradientEndpoints(['#1', '#2', '#3', '#4', '#5', '#6'])).toEqual(['#1', '#5'])
  })

  it("uses each built-in theme's dedicated visual-map gradient", () => {
    expect(resolveVisualMapColors('default')).toEqual(['#91CC75', '#EE6666'])
    expect(resolveVisualMapColors('macarons')).toEqual(['#5ab1ef', '#d87a80'])
    expect(resolveVisualMapColors('#111,#222,#333')).toEqual(['#111', '#333'])
  })
})
