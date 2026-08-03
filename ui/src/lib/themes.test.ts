import { describe, expect, it, beforeEach } from 'vitest'
import type { Theme } from '../types'
import {
  DEFAULT_THEME,
  findTheme,
  listAvailableThemeNames,
  normalizeTheme,
  paletteGradientEndpoints,
  parseCustomPalette,
  registerDatasetThemes,
  resolveActiveTheme,
  resolvePalette,
  resolveTheme,
  resolveVisualMapColors,
  THEME_NAMES,
} from './themes'

const westeros: Theme = {
  name: 'westeros',
  colors: [
    '#516b91',
    '#59c4e6',
    '#edafda',
    '#93b7e3',
    '#a5e7f0',
    '#cbb0e3',
    '#3f5575',
    '#41a7cb',
    '#d58fc4',
    '#789bc7',
  ],
  visualMapColors: ['#59c4e6', '#d58fc4'],
}

const vintage: Theme = {
  name: 'vintage',
  colors: [
    '#d87c7c',
    '#919e8b',
    '#d7ab82',
    '#6e7074',
    '#61a0a8',
    '#efa18d',
    '#787464',
    '#cc7e63',
    '#724e58',
    '#4b565b',
  ],
  visualMapColors: ['#919e8b', '#d87c7c'],
}

describe('themes', () => {
  beforeEach(() => {
    registerDatasetThemes(undefined)
  })

  it('ships only the UI default theme in the bundle catalog', () => {
    expect(THEME_NAMES).toEqual(['default'])
    expect(DEFAULT_THEME.name).toBe('default')
    expect(DEFAULT_THEME.colors).toHaveLength(10)
    expect(new Set(DEFAULT_THEME.colors.map((c) => c.toLowerCase())).size).toBe(10)
    expect(DEFAULT_THEME.visualMapColors).toEqual(['#91CC75', '#EE6666'])
  })

  it('resolveActiveTheme uses themes[0] when present, else default', () => {
    expect(resolveActiveTheme({ themes: [westeros, vintage] })).toMatchObject({
      name: 'westeros',
      colors: westeros.colors,
      visualMapColors: westeros.visualMapColors,
    })
    expect(resolveActiveTheme({ themes: [] })).toEqual(DEFAULT_THEME)
    expect(resolveActiveTheme({})).toEqual(DEFAULT_THEME)
    expect(resolveActiveTheme(null)).toEqual(DEFAULT_THEME)
  })

  it('soft-handles legacy hex theme string when themes is empty', () => {
    expect(resolveActiveTheme({ theme: '#f00,#0f0,#00f' })).toMatchObject({
      name: 'custom',
      colors: ['#f00', '#0f0', '#00f'],
    })
  })

  it('falls back to default for legacy named themes without themes[]', () => {
    expect(resolveActiveTheme({ theme: 'westeros' })).toEqual(DEFAULT_THEME)
    expect(resolveActiveTheme({ theme: 'macarons' })).toEqual(DEFAULT_THEME)
  })

  it('registers dataset themes for name-based palette resolution', () => {
    registerDatasetThemes([westeros, vintage])
    expect(listAvailableThemeNames()).toEqual(['default', 'westeros', 'vintage'])
    expect(resolvePalette('WESTEROS')).toEqual(westeros.colors)
    expect(resolvePalette('vintage')).toEqual(vintage.colors)
    expect(resolveVisualMapColors('westeros')).toEqual(westeros.visualMapColors)
    expect(findTheme('missing')).toBeUndefined()
  })

  it('ignores dataset entries named default and entries without colors', () => {
    registerDatasetThemes([
      { name: 'default', colors: ['#000', '#111'], visualMapColors: ['#000', '#111'] },
      { name: 'empty', colors: [], visualMapColors: [] },
      westeros,
    ])
    expect(listAvailableThemeNames()).toEqual(['default', 'westeros'])
    expect(resolvePalette('default')).toEqual(DEFAULT_THEME.colors)
  })

  it('defaults invalid or unknown names to the UI default palette', () => {
    expect(resolvePalette('missing')).toEqual(DEFAULT_THEME.colors)
    expect(resolvePalette('VINTAGE')).toEqual(DEFAULT_THEME.colors) // not registered
    expect(normalizeTheme()).toBe('default')
    expect(normalizeTheme('roma')).toBe('default')
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

  it("uses the active theme's visualMapColors, with gradient fallback", () => {
    expect(resolveVisualMapColors('default')).toEqual(['#91CC75', '#EE6666'])
    registerDatasetThemes([westeros])
    expect(resolveVisualMapColors('westeros')).toEqual(['#59c4e6', '#d58fc4'])
    expect(resolveVisualMapColors('#111,#222,#333')).toEqual(['#111', '#333'])

    registerDatasetThemes([
      {
        name: 'no-vm',
        colors: ['#aaa', '#bbb', '#ccc'],
        visualMapColors: [],
      },
    ])
    expect(resolveVisualMapColors('no-vm')).toEqual(['#aaa', '#ccc'])
  })

  it('resolveTheme returns a full theme object for names and custom palettes', () => {
    registerDatasetThemes([vintage])
    expect(resolveTheme('vintage').name).toBe('vintage')
    expect(resolveTheme('#111,#222').visualMapColors).toEqual(['#111', '#222'])
  })
})
