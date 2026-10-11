import { describe, expect, it } from 'vitest'
import { affixForColumn } from './affix'
import type { LabelAffix } from '../types'

describe('affixForColumn', () => {
  it('returns empty string for undefined affix', () => {
    expect(affixForColumn(undefined, 'time')).toBe('')
  })

  it('returns global for any column when only global is set', () => {
    expect(affixForColumn({ global: '%' }, 'a')).toBe('%')
    expect(affixForColumn({ global: '%' }, 'b')).toBe('%')
  })

  it('matches columns exactly by name when only columns is set', () => {
    const affix: LabelAffix = { columns: { time: 'ms', ops: 'ops/s' } }
    expect(affixForColumn(affix, 'time')).toBe('ms')
    expect(affixForColumn(affix, 'ops')).toBe('ops/s')
    expect(affixForColumn(affix, 'unknown')).toBe('')
  })

  it('does not trim or case-fold column names', () => {
    const affix: LabelAffix = { columns: { Time: 'ms' } }
    expect(affixForColumn(affix, 'Time')).toBe('ms')
    expect(affixForColumn(affix, 'time')).toBe('')
    expect(affixForColumn(affix, ' Time ')).toBe('')
  })

  it('column entry wins over global', () => {
    const affix: LabelAffix = { global: '%', columns: { time: 'ms' } }
    expect(affixForColumn(affix, 'time')).toBe('ms')
    expect(affixForColumn(affix, 'other')).toBe('%')
  })

  it('treats empty-string values as set', () => {
    expect(affixForColumn({ global: '' }, 'a')).toBe('')
    expect(affixForColumn({ global: '%', columns: { a: '' } }, 'a')).toBe('')
  })
})
