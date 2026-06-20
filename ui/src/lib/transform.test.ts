import { describe, it, expect } from 'vitest'
import {
  buildChartForSignature,
  buildChartData,
  listChartSignatures,
  build3DRender,
  projectAndGroup,
  canonicalAxisOrdersFromStrings,
} from './transform'
import { translateAxisKey } from './swap'
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

  it('attaches grouped render3D only when x, y, and z are all populated', () => {
    const with3D: DataPoint[] = [dp('X1', 'Y1', 'Z1', 'v', 5)]
    const without3D: DataPoint[] = [dp('X1', 'Y1', '', 'v', 5)]

    const { signature, statTemplate } = listChartSignatures(with3D)[0]!
    const chart3D = buildChartForSignature(with3D, signature, statTemplate, undefined, noSort)
    const chart2D = buildChartForSignature(without3D, signature, statTemplate, undefined, noSort)

    expect(chart3D.render3D).toBeDefined()
    expect(chart3D.render3D!.mode).toBe('grouped')
    expect(chart2D.render3D).toBeUndefined()
  })

  it('attaches value render3D when threeD is on for x+y data', () => {
    const data: DataPoint[] = [
      dp('X1', 'Y1', '', 'v', 5),
      dp('X2', 'Y1', '', 'v', 3),
      dp('X1', 'Y2', '', 'v', 7),
    ]
    const { signature, statTemplate } = listChartSignatures(data)[0]!
    const chart = buildChartForSignature(
      data,
      signature,
      statTemplate,
      undefined,
      noSort,
      false,
      'linear',
      undefined,
      true
    )

    expect(chart.render3D).toBeDefined()
    expect(chart.render3D!.mode).toBe('value')
    expect(chart.render3D!.zValues).toEqual([])
    expect(chart.render3D!.barSeries).toHaveLength(1)
    expect(chart.render3D!.barSeries[0]!.data[0]!.value).toEqual([0, 0, 5])
  })

  it('skips value render3D when threeD toggle is off', () => {
    const data: DataPoint[] = [dp('X1', 'Y1', '', 'v', 5)]
    const { signature, statTemplate } = listChartSignatures(data)[0]!
    const chart = buildChartForSignature(
      data,
      signature,
      statTemplate,
      undefined,
      noSort,
      false,
      'linear',
      undefined,
      false
    )
    expect(chart.render3D).toBeUndefined()
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
// 3-axis swap: grouped 3D (xyz) vs value 3D (xyn/nxy) when threeD is baked
// ---------------------------------------------------------------------------
describe('3-axis swap with threeD', () => {
  const identity: Array<'xAxis' | 'yAxis' | 'zAxis'> = ['xAxis', 'yAxis', 'zAxis']
  const raw: DataPoint[] = [
    dp('X1', 'Y1', 'Z1', 'v', 1),
    dp('X2', 'Y1', 'Z1', 'v', 2),
    dp('X1', 'Y2', 'Z2', 'v', 3),
  ]

  const renderModeForSwap = (swap: string, threeD: boolean) => {
    const { grouped } = projectAndGroup(raw, identity, translateAxisKey(swap))
    const rows = grouped.values().next().value ?? []
    const { signature, statTemplate } = listChartSignatures(raw)[0]!
    const chart = buildChartForSignature(
      rows,
      signature,
      statTemplate,
      undefined,
      noSort,
      false,
      'linear',
      undefined,
      threeD
    )
    return chart.render3D?.mode
  }

  it('xyz + threeD renders grouped 3D', () => {
    expect(renderModeForSwap('xyz', true)).toBe('grouped')
  })

  it('xyn + threeD renders value 3D', () => {
    expect(renderModeForSwap('xyn', true)).toBe('value')
  })

  it('nxy + threeD renders value 3D', () => {
    expect(renderModeForSwap('nxy', true)).toBe('value')
  })

  it('swap back to xyz + threeD renders grouped 3D again', () => {
    expect(renderModeForSwap('xyn', true)).toBe('value')
    expect(renderModeForSwap('xyz', true)).toBe('grouped')
  })
})

// ---------------------------------------------------------------------------
// canonical axis order (3D z stability on arrangement change)
// ---------------------------------------------------------------------------
describe('canonical axis order', () => {
  const identityString = 'nxyz'

  it('orders zValues by raw first-seen when sort is off', () => {
    const raw: DataPoint[] = [
      dp('X1', 'Y1', 'Z3', 'v', 1),
      dp('X1', 'Y1', 'Z1', 'v', 2),
      dp('X1', 'Y1', 'Z2', 'v', 3),
    ]
    const canonical = canonicalAxisOrdersFromStrings(raw, 'xyz', 'xyz')
    const { signature, statTemplate } = listChartSignatures(raw)[0]!
    const chart = buildChartForSignature(
      raw,
      signature,
      statTemplate,
      undefined,
      noSort,
      false,
      'linear',
      canonical
    )
    expect(chart.render3D!.zValues).toEqual(['Z3', 'Z1', 'Z2'])
  })

  it('keeps zValues order stable when chart axes permute but z source is unchanged (sort off)', () => {
    const raw: DataPoint[] = [
      { name: 'G1', xAxis: 'X1', yAxis: 'Y1', zAxis: 'Z2', stats: [{ type: 'v', value: 1 }] },
      { name: 'G1', xAxis: 'X1', yAxis: 'Y1', zAxis: 'Z1', stats: [{ type: 'v', value: 2 }] },
      { name: 'G2', xAxis: 'X2', yAxis: 'Y1', zAxis: 'Z3', stats: [{ type: 'v', value: 3 }] },
    ]
    const { signature, statTemplate } = listChartSignatures(raw)[0]!

    const identity = translateAxisKey(identityString)

    const zForTarget = (targetString: string) => {
      const target = translateAxisKey(targetString)
      const canonical = canonicalAxisOrdersFromStrings(raw, identityString, targetString)
      expect(canonical.z).toEqual(['Z2', 'Z1', 'Z3'])
      const { grouped } = projectAndGroup(raw, identity, target)
      expect(grouped.has('G1'), `missing G1 group for ${targetString}`).toBe(true)
      const rows = grouped.get('G1')!
      const chart = buildChartForSignature(
        rows,
        signature,
        statTemplate,
        undefined,
        noSort,
        false,
        'linear',
        canonical
      )
      expect(chart.render3D, `expected 3D chart for ${targetString}`).toBeDefined()
      return chart.render3D!.zValues
    }

    expect(zForTarget('nxyz')).toEqual(['Z2', 'Z1'])
    expect(zForTarget('nyxz')).toEqual(['Z2', 'Z1'])
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

// ---------------------------------------------------------------------------
// buildValueModeChart
// ---------------------------------------------------------------------------
import { buildValueModeChart, buildHybridScatterChart } from './transform'
import type { Axis } from '../types'

describe('buildValueModeChart', () => {
  const valueAxes: Axis[] = [
    { key: 'x', label: 'price', type: 'value' },
    { key: 'y', label: 'latency', type: 'value' },
  ]

  function vdp(xAxis: string, yAxis: string): DataPoint {
    return { xAxis, yAxis, stats: [] }
  }

  it('parses string coordinates into [number, number] tuples', () => {
    const data = [vdp('100', '12'), vdp('200', '8')]

    const chart = buildValueModeChart(data, valueAxes)

    expect(chart.valueTuples).toEqual([
      [100, 12],
      [200, 8],
    ])
  })

  it('drops rows with non-finite x or y', () => {
    const data = [vdp('1', '2'), vdp('bad', '3'), vdp('4', 'NaN')]

    const chart = buildValueModeChart(data, valueAxes)

    expect(chart.valueTuples).toEqual([[1, 2]])
  })

  it('title combines x and y labels', () => {
    const chart = buildValueModeChart([vdp('1', '2')], valueAxes)
    expect(chart.title).toBe('price vs latency')
  })

  it('falls back to axis key when label is absent', () => {
    const axes: Axis[] = [
      { key: 'x', type: 'value' },
      { key: 'y', type: 'value' },
    ]
    const chart = buildValueModeChart([vdp('1', '2')], axes)
    expect(chart.title).toBe('x vs y')
  })

  it('sets axisLabels from axes', () => {
    const chart = buildValueModeChart([vdp('1', '2')], valueAxes)
    expect(chart.axisLabels).toEqual({ x: 'price', y: 'latency' })
  })

  it('emits empty series, points, yAxis, zAxis', () => {
    const chart = buildValueModeChart([vdp('1', '2')], valueAxes)
    expect(chart.series).toEqual([])
    expect(chart.points).toEqual([])
    expect(chart.yAxis).toEqual([])
    expect(chart.zAxis).toEqual([])
  })

  it('returns empty valueTuples for empty data', () => {
    const chart = buildValueModeChart([], valueAxes)
    expect(chart.valueTuples).toEqual([])
  })
})

describe('buildValueModeChart — 3-col swap-driven 3D', () => {
  const valueAxes3: Axis[] = [
    { key: 'x', label: 'x', type: 'value' },
    { key: 'y', label: 'y', type: 'value' },
    { key: 'z', label: 'z', type: 'value' },
  ]

  function vdp3(x: string, y: string, z: string): DataPoint {
    return { xAxis: x, yAxis: y, zAxis: z, stats: [] }
  }

  it('xyz swap emits continuous render3D', () => {
    const chart = buildValueModeChart([vdp3('1', '2', '3')], valueAxes3, 'xyz', 'xyz')
    expect(chart.render3D?.mode).toBe('continuous')
    expect(chart.valuePoints3D).toEqual([[1, 2, 3]])
    expect(chart.valueTuples).toBeUndefined()
  })

  it('nxy swap emits 2D tuples projecting y and z onto chart axes', () => {
    const chart = buildValueModeChart([vdp3('1', '2', '3')], valueAxes3, 'xyz', 'nxy')
    expect(chart.render3D).toBeUndefined()
    expect(chart.valueTuples).toEqual([[2, 3]])
  })

  it('respects log scale on 3D path', () => {
    const chart = buildValueModeChart([vdp3('1', '2', '0')], valueAxes3, 'xyz', 'xyz', {
      scale: 'log',
    })
    expect(chart.valuePoints3D).toBeUndefined()
    expect(chart.render3D).toBeUndefined()
  })
})

describe('buildHybridScatterChart', () => {
  const hybridAxes: Axis[] = [
    { key: 'x', label: 'Region' },
    { key: 'y', label: 'Category' },
    { key: 'z', label: 'Latency (ms)', type: 'value' },
  ]

  function hdp(xAxis: string, yAxis: string, value: number): DataPoint {
    return {
      xAxis,
      yAxis,
      stats: [{ type: 'Latency (ms)', value }],
    }
  }

  it('reads z from stats and builds grouped series', () => {
    const data = [hdp('US', 'Widget', 12), hdp('EU', 'Gadget', 8), hdp('US', 'Gadget', 5)]
    const chart = buildHybridScatterChart(data, hybridAxes)

    expect(chart.statType).toBe('Latency (ms)')
    expect(chart.yAxis).toEqual(expect.arrayContaining(['Widget', 'Gadget']))
    expect(chart.series).toHaveLength(2)
    const us = chart.series.find((s) => s.xAxis === 'US')
    expect(us?.values).toEqual(expect.arrayContaining([12, 5]))
  })

  it('attaches value render3D when threeD is on', () => {
    const data = [hdp('US', 'Widget', 12), hdp('EU', 'Gadget', 8)]
    const chart = buildHybridScatterChart(data, hybridAxes, 'xy', 'xy', { threeD: true })

    expect(chart.render3D?.mode).toBe('value')
    expect(chart.render3D?.zValues).toEqual([])
  })

  it('skips rows without a matching z stat', () => {
    const data: DataPoint[] = [
      hdp('US', 'Widget', 12),
      { xAxis: 'EU', yAxis: 'Gadget', stats: [{ type: 'other', value: 99 }] },
    ]
    const chart = buildHybridScatterChart(data, hybridAxes)
    expect(chart.points).toHaveLength(1)
  })
})
