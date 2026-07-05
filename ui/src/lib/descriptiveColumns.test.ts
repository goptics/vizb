import { describe, expect, it } from 'vitest'
import {
  CATEGORY_KEYS,
  DESCRIPTIVE_COLUMNS,
  columnsFromKeys,
  defaultSelectedKeys,
  keysForMath,
} from './descriptiveColumns'

describe('descriptiveColumns', () => {
  it('defaults to every descriptive column when math is absent or empty', () => {
    const allKeys = DESCRIPTIVE_COLUMNS.map((col) => col.key)

    expect(defaultSelectedKeys()).toEqual(allKeys)
    expect(defaultSelectedKeys([])).toEqual(allKeys)
  })

  it('maps stat math categories to ordered column keys', () => {
    expect(keysForMath(['counts', 'center'])).toEqual([
      ...CATEGORY_KEYS.counts,
      ...CATEGORY_KEYS.center,
    ])
  })

  it('ignores correlations when seeding descriptive columns', () => {
    expect(keysForMath(['correlations', 'spread'])).toEqual(CATEGORY_KEYS.spread)
  })

  it('keeps canonical column order when resolving selected keys', () => {
    expect(columnsFromKeys(['median', 'count', 'stdDev']).map((col) => col.key)).toEqual([
      'count',
      'median',
      'stdDev',
    ])
  })
})
