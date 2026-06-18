import { describe, expect, it } from 'vitest'
import { chunkKeyOf, detectChartRoots, goStringMap, goStringSliceMap } from './go-codegen.ts'

describe('chunkKeyOf', () => {
  it('strips .js and prefixes with vizb:', () => {
    expect(chunkKeyOf('Chart3D-B_EOJPB8.js')).toBe('vizb:Chart3D-B_EOJPB8')
  })
})

describe('goStringMap', () => {
  it('emits an empty Go map body', () => {
    expect(goStringMap({})).toBe('{\n\n}')
  })

  it('emits Go double-quoted string literals', () => {
    const out = goStringMap({ foo: 'bar', baz: 'qux' })
    expect(out).toContain('\t"foo": "bar",')
    expect(out).toContain('\t"baz": "qux",')
  })
})

describe('goStringSliceMap', () => {
  it('emits an empty Go map body', () => {
    expect(goStringSliceMap({})).toBe('{\n\n}')
  })

  it('emits Go string slices', () => {
    const out = goStringSliceMap({ key: ['a', 'b'] })
    expect(out).toContain('\t"key": {"a", "b"},')
  })
})

describe('detectChartRoots', () => {
  it('maps hashed chart chunk filenames to logical chart keys', () => {
    const roots = detectChartRoots([
      'index-abc123.js',
      'ChartBar-DR8nN4wi.js',
      'Chart3D-B_EOJPB8.js',
    ])
    expect(roots).toEqual({
      bar: 'vizb:ChartBar-DR8nN4wi',
      '3d': 'vizb:Chart3D-B_EOJPB8',
    })
  })

  it('maps unhashed chart chunk filenames', () => {
    const roots = detectChartRoots(['ChartPie.js'], { ChartPie: 'pie' })
    expect(roots).toEqual({ pie: 'vizb:ChartPie' })
  })
})
