import { describe, it, expect } from 'vitest'
import { toCsvCell, descriptiveCsv, correlationCsv } from './csv'
import { columnsFromKeys } from './descriptiveColumns'
import type { DescriptiveStats, SeriesProfile } from '../types'

// Full DescriptiveStats with every key NaN, then apply overrides — the CSV
// builders only read the columns they're given, but the type wants all keys.
function stats(overrides: Partial<DescriptiveStats>): DescriptiveStats {
  const keys: (keyof DescriptiveStats)[] = [
    'count',
    'missing',
    'unique',
    'mean',
    'median',
    'mode',
    'variance',
    'stdDev',
    'min',
    'max',
    'range',
    'iqr',
    'mad',
    'cv',
    'skewness',
    'kurtosis',
    'p5',
    'p25',
    'p75',
    'p95',
  ]
  const base = Object.fromEntries(keys.map((k) => [k, NaN])) as DescriptiveStats
  return { ...base, ...overrides }
}

describe('toCsvCell', () => {
  it('plain string passes through', () => expect(toCsvCell('abc')).toBe('abc'))
  it('quotes commas', () => expect(toCsvCell('a,b')).toBe('"a,b"'))
  it('escapes quotes by doubling', () => expect(toCsvCell('a"b')).toBe('"a""b"'))
  it('quotes newlines', () => expect(toCsvCell('a\nb')).toBe('"a\nb"'))
  it('quotes edge whitespace', () => expect(toCsvCell(' x')).toBe('" x"'))
  it('finite number full precision', () => expect(toCsvCell(1.5)).toBe('1.5'))
  it('NaN → empty', () => expect(toCsvCell(NaN)).toBe(''))
  it('Infinity → empty', () => expect(toCsvCell(Infinity)).toBe(''))
})

describe('descriptiveCsv', () => {
  const columns: { key: keyof DescriptiveStats; label: string }[] = [
    { key: 'count', label: 'Count' },
    { key: 'mean', label: 'Mean' },
  ]
  it('header + rows, full precision, NaN blank', () => {
    const profiles: SeriesProfile[] = [
      { name: 'A', stats: stats({ count: 2, mean: 1.5 }) },
      { name: 'B', stats: stats({ count: 3, mean: NaN }) },
    ]
    expect(descriptiveCsv(profiles, columns)).toBe('Series,Count,Mean\nA,2,1.5\nB,3,')
  })
  it('quotes a series name with a comma', () => {
    const profiles: SeriesProfile[] = [{ name: 'a,b', stats: stats({ count: 1, mean: 0 }) }]
    expect(descriptiveCsv(profiles, columns)).toBe('Series,Count,Mean\n"a,b",1,0')
  })

  it('exports only the currently visible descriptive columns', () => {
    const profiles: SeriesProfile[] = [
      { name: 'A', stats: stats({ count: 2, mean: 1.5, median: 1.25 }) },
    ]
    expect(descriptiveCsv(profiles, columnsFromKeys(['median']))).toBe('Series,Median\nA,1.25')
  })
})

describe('correlationCsv', () => {
  it('corner cell, labels header, NaN blank', () => {
    const csv = correlationCsv(
      ['x', 'y'],
      [
        [1, NaN],
        [NaN, 1],
      ]
    )
    expect(csv).toBe(',x,y\nx,1,\ny,,1')
  })
})
