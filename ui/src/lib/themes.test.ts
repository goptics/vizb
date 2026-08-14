import { describe, expect, it, beforeEach } from 'vitest'
import type { Theme } from '../types'
import {
  activeThemeName,
  applyTheme,
  authorThemeCount,
  DEFAULT_THEME,
  findTheme,
  isAvailableThemeName,
  isThemeName,
  listAvailableThemes,
  listAvailableThemeNames,
  normalizeTheme,
  parseCustomPalette,
  registerDatasetThemes,
  resolveActiveTheme,
  resolveTheme,
  resolveVisualMapColors,
  shouldShowThemeSelector,
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
    expect(DEFAULT_THEME.name).toBe('default')
    expect(DEFAULT_THEME.colors).toHaveLength(10)
    expect(new Set(DEFAULT_THEME.colors.map((c) => c.toLowerCase())).size).toBe(10)
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
    expect(authorThemeCount()).toBe(2)
    expect(shouldShowThemeSelector()).toBe(true)
    expect(listAvailableThemeNames()).toEqual(['default', 'westeros', 'vintage'])
    expect(isAvailableThemeName('default')).toBe(true)
    expect(isAvailableThemeName('WESTEROS')).toBe(true)
    expect(isAvailableThemeName('missing')).toBe(false)
    expect(resolveTheme('WESTEROS').colors).toEqual(westeros.colors)
    expect(resolveTheme('vintage').colors).toEqual(vintage.colors)
    expect(resolveVisualMapColors('westeros')).toEqual(westeros.visualMapColors)
    expect(findTheme('missing')).toBeUndefined()
  })

  it('scopes available set to the single author theme when only one is embedded', () => {
    registerDatasetThemes([westeros])
    expect(authorThemeCount()).toBe(1)
    expect(shouldShowThemeSelector()).toBe(false)
    expect(listAvailableThemeNames()).toEqual(['westeros'])
    expect(isAvailableThemeName('westeros')).toBe(true)
    expect(isAvailableThemeName('default')).toBe(false)
    // Palette resolution still finds UI default by name; availability is separate.
    expect(resolveTheme('default').colors).toEqual([...DEFAULT_THEME.colors])
  })

  it('with zero author themes available set is only default and selector is hidden', () => {
    registerDatasetThemes(undefined)
    expect(authorThemeCount()).toBe(0)
    expect(shouldShowThemeSelector()).toBe(false)
    expect(listAvailableThemeNames()).toEqual(['default'])
    expect(isAvailableThemeName('default')).toBe(true)
    expect(isAvailableThemeName('westeros')).toBe(false)
  })

  it('ignores dataset entries named default and entries without colors', () => {
    registerDatasetThemes([
      { name: 'default', colors: ['#000', '#111'], visualMapColors: ['#000', '#111'] },
      { name: 'empty', colors: [], visualMapColors: [] },
      westeros,
    ])
    // Only one real author theme → available set is that theme alone.
    expect(authorThemeCount()).toBe(1)
    expect(listAvailableThemeNames()).toEqual(['westeros'])
    expect(resolveTheme('default').colors).toEqual([...DEFAULT_THEME.colors])
  })

  it('defaults invalid or unknown names to the UI default palette', () => {
    expect(resolveTheme('missing').colors).toEqual([...DEFAULT_THEME.colors])
    expect(resolveTheme('VINTAGE').colors).toEqual([...DEFAULT_THEME.colors]) // not registered
    expect(normalizeTheme()).toBe('default')
    expect(normalizeTheme('roma')).toBe('default')
  })

  it('accepts flexible custom #rgb and #rrggbb palettes', () => {
    expect(parseCustomPalette('#f00, #00ff00,#00f')).toEqual(['#f00', '#00ff00', '#00f'])
    expect(resolveTheme('#f00,#0f0').colors).toEqual(['#f00', '#0f0'])
    expect(parseCustomPalette('#f00')).toBeUndefined()
    expect(parseCustomPalette('#ggg,#000')).toBeUndefined()
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

  it('isThemeName recognizes default and registered dataset names only', () => {
    expect(isThemeName(undefined)).toBe(false)
    expect(isThemeName('')).toBe(false)
    expect(isThemeName('   ')).toBe(false)
    expect(isThemeName('default')).toBe(true)
    expect(isThemeName('DEFAULT')).toBe(true)
    expect(isThemeName('westeros')).toBe(false)
    registerDatasetThemes([westeros])
    expect(isThemeName('WESTEROS')).toBe(true)
    expect(isThemeName('missing')).toBe(false)
  })

  it('isAvailableThemeName and findTheme handle empty and whitespace names', () => {
    expect(isAvailableThemeName(undefined)).toBe(false)
    expect(isAvailableThemeName('')).toBe(false)
    expect(isAvailableThemeName('   ')).toBe(false)
    expect(findTheme(undefined)).toBeUndefined()
    expect(findTheme('')).toBeUndefined()
    expect(findTheme('   ')).toBeUndefined()
    expect(findTheme('default')?.name).toBe('default')
  })

  it('normalizeTheme returns the registered author spelling', () => {
    registerDatasetThemes([{ ...westeros, name: 'ROMA' }])
    expect(normalizeTheme('roma')).toBe('ROMA')
    expect(normalizeTheme('ROMA')).toBe('ROMA')
    applyTheme('roma')
    expect(activeThemeName.value).toBe('ROMA')
  })

  it('resolveActiveTheme fills missing visualMapColors and skips empty first theme', () => {
    expect(
      resolveActiveTheme({
        themes: [{ name: 'brand', colors: ['#111', '#222', '#333'], visualMapColors: [] }],
      })
    ).toMatchObject({
      name: 'brand',
      colors: ['#111', '#222', '#333'],
      visualMapColors: ['#111', '#333'],
    })
    expect(
      resolveActiveTheme({
        themes: [{ name: '', colors: ['#aaa', '#bbb'], visualMapColors: ['#aaa', '#bbb'] }],
      }).name
    ).toBe('custom')
    // First entry without colors falls through to default / legacy.
    expect(
      resolveActiveTheme({
        themes: [{ name: 'empty', colors: [], visualMapColors: [] }],
      })
    ).toEqual(DEFAULT_THEME)
    // Legacy # string that is not a valid multi-hex palette → default.
    expect(resolveActiveTheme({ theme: '#f00' })).toEqual(DEFAULT_THEME)
    expect(resolveActiveTheme({ theme: '#ggg,#000' })).toEqual(DEFAULT_THEME)
  })

  it('registerDatasetThemes names a blank colors-only themes[0] as custom', () => {
    registerDatasetThemes([
      { name: '  ', colors: ['#111', '#222'], visualMapColors: ['#111', '#222'] },
      { name: undefined as unknown as string, colors: ['#a', '#b'], visualMapColors: ['#a', '#b'] },
      westeros,
    ])
    // First blank-name entry → custom; later blank names still skipped.
    expect(authorThemeCount()).toBe(2)
    expect(listAvailableThemeNames()).toEqual(['default', 'custom', 'westeros'])
    expect(findTheme('custom')?.colors).toEqual(['#111', '#222'])
    expect(normalizeTheme('custom')).toBe('custom')
  })

  it('clones themes that omit visualMapColors as an empty list', () => {
    registerDatasetThemes([{ name: 'no-vm-field', colors: ['#111', '#222'] }])
    expect(findTheme('no-vm-field')).toMatchObject({
      name: 'no-vm-field',
      colors: ['#111', '#222'],
      visualMapColors: [],
    })
    expect(listAvailableThemes()[0]?.visualMapColors).toEqual([])
  })
})
