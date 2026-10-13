import { describe, expect, it } from 'vitest'
import {
  affixesOf,
  affixForColumn,
  applyAffixes,
  escapeHtml,
  NO_WRAP,
  tooltipAffixes,
  wrapFor,
} from './affix'
import type { LabelAffix } from '../types'

describe('applyAffixes', () => {
  it('wraps the value with prefix and suffix', () => {
    expect(applyAffixes('42', { prefix: '$', suffix: '%' })).toBe('$42%')
  })

  it('passes the value through when both affixes are empty', () => {
    expect(applyAffixes('42', NO_WRAP)).toBe('42')
  })

  it('applies a lone prefix or a lone suffix unchanged', () => {
    expect(applyAffixes('42', { prefix: '$', suffix: '' })).toBe('$42')
    expect(applyAffixes('42', { prefix: '', suffix: '%' })).toBe('42%')
  })
})

describe('wrapFor', () => {
  it('resolves both sides for a column, empty without config', () => {
    const a = { prefix: { global: '$' }, suffix: { columns: { time: 'ms' } } }
    expect(wrapFor(a, 'time')).toEqual({ prefix: '$', suffix: 'ms' })
    expect(wrapFor(a, 'other')).toEqual({ prefix: '$', suffix: '' })
    expect(wrapFor(undefined, 'time')).toEqual(NO_WRAP)
  })

  it('applies the fallback per side only when that side resolves empty', () => {
    const fallback = { prefix: '~', suffix: '%' }
    const a = { suffix: { columns: { time: 'ms' } } }
    expect(wrapFor(a, 'time', fallback)).toEqual({ prefix: '~', suffix: 'ms' })
    expect(wrapFor(a, 'other', fallback)).toEqual(fallback)
  })
})

describe('affixesOf', () => {
  it('reads labelPrefix / labelSuffix ref values', () => {
    const prefix = { global: '$' }
    const suffix = { global: '%' }
    expect(affixesOf({ labelPrefix: { value: prefix }, labelSuffix: { value: suffix } })).toEqual({
      prefix,
      suffix,
    })
    expect(affixesOf({})).toEqual({ prefix: undefined, suffix: undefined })
  })
})

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

describe('escapeHtml', () => {
  it('escapes HTML-significant characters', () => {
    expect(escapeHtml(`<img src=x onerror="a('&')">`)).toBe(
      '&lt;img src=x onerror=&quot;a(&#39;&amp;&#39;)&quot;&gt;'
    )
  })

  it('leaves plain affixes untouched', () => {
    expect(escapeHtml('ms')).toBe('ms')
    expect(escapeHtml('')).toBe('')
  })
})

describe('tooltipAffixes', () => {
  it('escapes both affixes around the value', () => {
    expect(tooltipAffixes('5', { prefix: '<b>', suffix: '&' })).toBe('&lt;b&gt;5&amp;')
  })
})
