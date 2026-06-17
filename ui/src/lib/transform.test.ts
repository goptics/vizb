import { describe, it, expect } from 'vitest'
import {
  buildChartForSignature,
  buildChartData,
  listChartSignatures,
  build3DRender,
  projectAndGroup,
} from './transform'
import type { DataPoint, Sort, Point3D } from '../types'

const noSort: Sort = { enabled: false, order: 'asc' }
const descSort: Sort = { enabled: true, order: 'desc' }

function dp(xAxis: string, yAxis = '', zAxis = '', type = 'val', value = 1): DataPoint {
  return { xAxis, yAxis, zAxis, stats: [{ type, value }] }
}

function build(data: DataPoint[]) {
  const { signature, statTemplate } = listChartSignatures(data)[0]!
  return buildChartForSignature(data, signature, statTemplate, undefined, noSort)
}

const valuesFor = (data: DataPoint[], xAxis: string) =>
  build(data).series.find((s) => s.xAxis === xAxis)!.values

// ---------------------------------------------------------------------------
// listChartSignatures
// ---------------------------------------------------------------------------
describe('listChartSignatures', () => {
  it('returns empty for empty data', () => {
    expect(listChartSignatures([])).toEqual([])
  })

  it('preserves first-seen order across multiple rows', () => {
    const data: DataPoint[] = [
      { stats: [{ type: 'ns', unit: 'ms' }] },
      { stats: [{ type: 'mem', unit: 'B' }] },
    ]
    const sigs = listChartSignatures(data).map((s) => s.signature)
    expect(sigs[0]).toContain('ns')
    expect(sigs[1]).toContain('mem')
  })

  it('deduplicates same signature from multiple rows', () => {
    const data: DataPoint[] = [{ stats: [{ type: 'val' }] }, { stats: [{ type: 'val' }] }]
    expect(listChartSignatures(data)).toHaveLength(1)
  })

  it('skips rows with missing or empty stats', () => {
    const data: DataPoint[] = [
      { stats: [] },
      { stats: undefined as unknown as [] },
      { stats: [{ type: 'real' }] },
    ]
    const sigs = listChartSignatures(data)
    expect(sigs).toHaveLength(1)
    expect(sigs[0]!.signature).toContain('real')
  })
})

// ---------------------------------------------------------------------------
// buildChartForSignature — duplicate (xAxis,yAxis) handling (existing tests kept)
// ---------------------------------------------------------------------------
describe('buildChartForSignature — duplicate (xAxis,yAxis) handling', () => {
  it('averages duplicates rather than dropping (overwrite) them', () => {
    const data: DataPoint[] = [
      { xAxis: 'A', yAxis: '', stats: [{ type: 'val', value: 100 }] },
      { xAxis: 'A', yAxis: '', stats: [{ type: 'val', value: 200 }] },
    ]
    expect(valuesFor(data, 'A')).toEqual([150])
  })

  it('averages benchmark count=N repeats to their mean (not their sum)', () => {
    const data: DataPoint[] = [124, 126, 122, 128, 125].map((value) => ({
      xAxis: 'BenchFoo',
      yAxis: '',
      stats: [{ type: 'ns', value }],
    }))
    expect(valuesFor(data, 'BenchFoo')).toEqual([125])
  })

  it('keeps distinct keys separate', () => {
    const data: DataPoint[] = [
      { xAxis: 'A', yAxis: '', stats: [{ type: 'val', value: 100 }] },
      { xAxis: 'B', yAxis: '', stats: [{ type: 'val', value: 200 }] },
    ]
    expect(valuesFor(data, 'A')).toEqual([100])
    expect(valuesFor(data, 'B')).toEqual([200])
  })

  it('single point per key is unchanged (grouped-CSV pre-summed case)', () => {
    const data: DataPoint[] = [
      { xAxis: 'East', yAxis: '', stats: [{ type: 'val', value: 5000 }] },
      { xAxis: 'West', yAxis: '', stats: [{ type: 'val', value: 3000 }] },
    ]
    expect(valuesFor(data, 'East')).toEqual([5000])
    expect(valuesFor(data, 'West')).toEqual([3000])
  })
})

// ---------------------------------------------------------------------------
// buildChartForSignature — additional branches
// ---------------------------------------------------------------------------
describe('buildChartForSignature — additional branches', () => {
  it('returns empty series and points when signature is absent from data', () => {
    const data: DataPoint[] = [dp('A', '', '', 'val', 10)]
    const { signature, statTemplate } = listChartSignatures(data)[0]!
    const chart = buildChartForSignature(
      data,
      'nonexistent-sig',
      { type: 'nonexistent' },
      undefined,
      noSort
    )
    expect(chart.series).toEqual([])
    expect(chart.points).toEqual([])
    expect(chart.render3D).toBeUndefined()
    // suppress unused-var warning on purpose — we read from listChartSignatures to get a valid sig
    void signature
    void statTemplate
  })

  it('attaches render3D only when x, y, and z are all populated', () => {
    const with3D: DataPoint[] = [dp('X1', 'Y1', 'Z1', 'v', 5)]
    const without3D: DataPoint[] = [dp('X1', 'Y1', '', 'v', 5)]

    const { signature, statTemplate } = listChartSignatures(with3D)[0]!
    const chart3D = buildChartForSignature(with3D, signature, statTemplate, undefined, noSort)
    const chart2D = buildChartForSignature(without3D, signature, statTemplate, undefined, noSort)

    expect(chart3D.render3D).toBeDefined()
    expect(chart2D.render3D).toBeUndefined()
  })

  it('desc sort places highest-total xAxis series first', () => {
    const data: DataPoint[] = [
      dp('Low', '', '', 'v', 1),
      dp('High', '', '', 'v', 9),
      dp('Mid', '', '', 'v', 5),
    ]
    const { signature, statTemplate } = listChartSignatures(data)[0]!
    const chart = buildChartForSignature(data, signature, statTemplate, undefined, descSort)
    expect(chart.series[0]!.xAxis).toBe('High')
    expect(chart.series[2]!.xAxis).toBe('Low')
  })
})

// ---------------------------------------------------------------------------
// build3DRender
// ---------------------------------------------------------------------------
describe('build3DRender', () => {
  const p3 = (xAxis: string, yAxis: string, zAxis: string, value: number): Point3D => ({
    xAxis,
    yAxis,
    zAxis,
    value,
  })

  it('log scale drops non-positive cells from both bar and line series', () => {
    const points = [p3('A', '1', 'Z1', -5), p3('B', '1', 'Z1', 10)]
    const render = build3DRender(points, ['Z1'], noSort, false, 'log')
    const allValues = [...render.barSeries[0]!.data, ...render.lineSeries[0]!.data].map(
      (d) => d.value[2]
    )
    expect(allValues).not.toContain(-5)
    expect(allValues).not.toContain(0)
    expect(allValues.some((v) => v === 10)).toBe(true)
  })

  it('linear: barSeries is a full grid (fills missing cells), lineSeries is sparse (only real data)', () => {
    // Two x values, two y values, only one data point → barSeries has 4 entries (full
    // xi×yi grid), lineSeries has 1 entry (only the cell that actually has data).
    const points = [p3('A', '1', 'Z1', 5)]
    const render = build3DRender(
      [...points, p3('B', '2', 'Z1', 0)],
      ['Z1'],
      noSort,
      false,
      'linear'
    )
    // bar: full grid = 2x×2y = 4 cells
    expect(render.barSeries[0]!.data).toHaveLength(4)
    // line: sparse = only 2 actual cells (A/1 and B/2)
    expect(render.lineSeries[0]!.data).toHaveLength(2)
  })

  it('showLabels=true aggregates cellTotals across z-series', () => {
    const points = [p3('A', '1', 'Z1', 10), p3('A', '1', 'Z2', 5)]
    const render = build3DRender(points, ['Z1', 'Z2'], noSort, true, 'linear')
    expect(render.cellTotals['0,0']).toBe(15)
  })

  it('showLabels=false leaves cellTotals empty', () => {
    const points = [p3('A', '1', 'Z1', 10)]
    const render = build3DRender(points, ['Z1'], noSort, false, 'linear')
    expect(Object.keys(render.cellTotals)).toHaveLength(0)
  })

  it('duplicates at same (xAxis,yAxis,zAxis) are averaged not summed', () => {
    const points = [p3('A', '1', 'Z1', 10), p3('A', '1', 'Z1', 30)]
    const render = build3DRender(points, ['Z1'], noSort, false, 'linear')
    const cell = render.barSeries[0]!.data.find((d) => d.value[0] === 0 && d.value[1] === 0)
    expect(cell!.value[2]).toBe(20)
  })

  it('filters empty-string z values from zAxisAll', () => {
    const points = [p3('A', '1', 'Z1', 5)]
    const render = build3DRender(points, ['Z1', ''], noSort, false, 'linear')
    expect(render.zValues).toEqual(['Z1'])
    expect(render.barSeries).toHaveLength(1)
  })
})

// ---------------------------------------------------------------------------
// projectAndGroup
// ---------------------------------------------------------------------------
describe('projectAndGroup', () => {
  it('length mismatch → single Default group, rows cloned as-is', () => {
    const raw: DataPoint[] = [dp('A', 'Y1')]
    const { grouped, groupNames } = projectAndGroup(raw, ['name', 'xAxis'], ['xAxis'])
    expect(groupNames).toEqual(['Default'])
    expect(grouped.get('Default')).toHaveLength(1)
  })

  it('"name" in targetKeys becomes group discriminator and is excluded from output row', () => {
    const raw: DataPoint[] = [
      { name: 'Alpha', xAxis: 'A', stats: [] },
      { name: 'Beta', xAxis: 'B', stats: [] },
    ]
    const { grouped, groupNames } = projectAndGroup(raw, ['name', 'xAxis'], ['name', 'xAxis'])
    expect(groupNames).toEqual(['Alpha', 'Beta'])
    expect(grouped.get('Alpha')![0]).not.toHaveProperty('name')
    expect(grouped.get('Alpha')![0]!.xAxis).toBe('A')
  })

  it('no "name" in targetKeys → all rows go to Default', () => {
    const raw: DataPoint[] = [dp('A', 'Y1'), dp('B', 'Y2')]
    const { grouped, groupNames } = projectAndGroup(raw, ['xAxis', 'yAxis'], ['xAxis', 'yAxis'])
    expect(groupNames).toEqual(['Default'])
    expect(grouped.get('Default')).toHaveLength(2)
  })

  it('empty name value falls back to Default group', () => {
    const raw: DataPoint[] = [{ name: '', xAxis: 'A', stats: [] }]
    const { groupNames } = projectAndGroup(raw, ['name', 'xAxis'], ['name', 'xAxis'])
    expect(groupNames).toEqual(['Default'])
  })

  it('preserves stats through projection unchanged', () => {
    const stat = { type: 'ns', value: 42 }
    const raw: DataPoint[] = [{ name: 'G', xAxis: 'X', stats: [stat] }]
    const { grouped } = projectAndGroup(raw, ['name', 'xAxis'], ['name', 'xAxis'])
    expect(grouped.get('G')![0]!.stats).toEqual([stat])
  })
})

// ---------------------------------------------------------------------------
// buildChartData (bulk entry point)
// ---------------------------------------------------------------------------
describe('buildChartData', () => {
  it('returns one ChartData per unique signature', () => {
    const data: DataPoint[] = [
      {
        xAxis: 'A',
        stats: [
          { type: 'val', unit: 'ms' },
          { type: 'mem', unit: 'B' },
        ],
      },
    ]
    const charts = buildChartData(data, undefined, noSort)
    expect(charts).toHaveLength(2)
    expect(charts.map((c) => c.statType)).toEqual(expect.arrayContaining(['val', 'mem']))
  })

  it('returns empty for empty data', () => {
    expect(buildChartData([], undefined, noSort)).toEqual([])
  })
})
