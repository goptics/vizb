import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest'
import {
  baseConfig,
  makeHeatmapChartData,
  emptyChartData,
  groupedRender3D,
  installDevicePixelRatio,
} from '@/test-utils'
import { LARGE_X_THRESHOLD } from './shared'
import { useHeatmapChartOptions } from './useHeatmapChartOptions'

let restoreDpr: () => void
beforeAll(() => {
  restoreDpr = installDevicePixelRatio()
})
afterAll(() => restoreDpr())

const longLabel = 'abcdefghij' // 10 chars

const grouped3DHeatmap = (overrides: Record<string, unknown> = {}) =>
  emptyChartData({
    title: 'matrix3d',
    statType: 'ns',
    series: [
      { xAxis: 'x1', values: [1], benchmarkId: '' },
      { xAxis: 'x2', values: [2], benchmarkId: '' },
    ],
    yAxis: ['y1', 'y2'],
    zAxis: ['zA', 'zB'],
    points: [
      { xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 },
      { xAxis: 'x1', yAxis: 'y1', zAxis: 'zB', value: 7 },
      { xAxis: 'x1', yAxis: 'y2', zAxis: 'zA', value: 3 },
      { xAxis: 'x2', yAxis: 'y1', zAxis: 'zA', value: 10 },
      { xAxis: 'x2', yAxis: 'y2', zAxis: 'zB', value: 2 },
    ],
    axisLabels: { x: 'X', y: 'Y', z: 'Z' },
    render3D: {
      ...groupedRender3D,
      xValues: ['x1', 'x2'],
      yValues: ['y1', 'y2'],
      zValues: ['zA', 'zB'],
    },
    ...overrides,
  })

describe('useHeatmapChartOptions — 2D', () => {
  it('emits heatmap series type', () => {
    const { options } = useHeatmapChartOptions(baseConfig({ chartData: makeHeatmapChartData() }))
    const series = options.value.series as { type: string }[]
    expect(series[0]!.type).toBe('heatmap')
  })

  it('omits dataZoom for small axes', () => {
    const { options } = useHeatmapChartOptions(baseConfig({ chartData: makeHeatmapChartData() }))
    expect(options.value.dataZoom).toBeUndefined()
  })

  it('attaches dataZoom when x categories exceed LARGE_X_THRESHOLD', () => {
    const manyX = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `x${i}`)
    const chartData = makeHeatmapChartData({
      series: manyX.map((x) => ({ xAxis: x, values: [1, 2], benchmarkId: '' })),
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    expect(options.value.dataZoom).toBeDefined()
    expect(
      Array.isArray(options.value.dataZoom) ? options.value.dataZoom.length : 1
    ).toBeGreaterThan(0)
  })

  it('attaches dataZoom when y categories exceed LARGE_X_THRESHOLD', () => {
    const manyY = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `y${i}`)
    const chartData = makeHeatmapChartData({
      yAxis: manyY,
      series: [
        {
          xAxis: 'x1',
          values: manyY.map((_, i) => i),
          benchmarkId: '',
        },
      ],
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    expect(options.value.dataZoom).toBeDefined()
  })

  it('expands equal min/max visualMap range', () => {
    const chartData = makeHeatmapChartData({
      series: [
        { xAxis: 'a', values: [5, 5], benchmarkId: '' },
        { xAxis: 'b', values: [5, 5], benchmarkId: '' },
      ],
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const vm = options.value.visualMap as { min: number; max: number }
    expect(vm.min).toBe(4)
    expect(vm.max).toBe(6)
  })

  it('rotates x labels when total label length exceeds 100 without large X', () => {
    const chartData = makeHeatmapChartData({
      series: Array.from({ length: 12 }, (_, i) => ({
        xAxis: `${longLabel}${i}`,
        values: [1, 2],
        benchmarkId: '',
      })),
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const xAxis = options.value.xAxis as { axisLabel?: { rotate?: number } }
    expect(xAxis.axisLabel?.rotate).toBe(30)
  })

  it('omits axis names when axisLabels are absent', () => {
    const chartData = makeHeatmapChartData({ axisLabels: {} })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    expect((options.value.xAxis as { name?: string }).name).toBeUndefined()
    expect((options.value.yAxis as { name?: string }).name).toBeUndefined()
  })

  it('formats 2D tooltip with row/column percentages', () => {
    const { options } = useHeatmapChartOptions(baseConfig({ chartData: makeHeatmapChartData() }))
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    const html = tooltip.formatter({ data: [0, 0, 1] })
    expect(html).toContain('<b>x1 / y1</b>')
    expect(html).toContain('Value: <b>1</b>')
    expect(html).toContain('Σ x1:')
    expect(html).toContain('Σ y1:')
    expect(html).toContain('%')
  })

  it('handles zero totals in 2D tooltip percentages', () => {
    const chartData = makeHeatmapChartData({
      series: [
        { xAxis: 'a', values: [0, 0], benchmarkId: '' },
        { xAxis: 'b', values: [0, 0], benchmarkId: '' },
      ],
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    const html = tooltip.formatter({ data: [0, 0, 0] })
    expect(html).toContain('(0.0%)')
  })

  it('formats cell labels with K/M suffixes when showLabels is on', () => {
    const chartData = makeHeatmapChartData({
      series: [
        { xAxis: 'a', values: [1500, 2_500_000], benchmarkId: '' },
        { xAxis: 'b', values: [0.5, 42], benchmarkId: '' },
      ],
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData, showLabels: true }))
    const series = options.value.series as {
      label: { formatter: (p: { data: number[] }) => string }
    }[]
    const fmt = series[0]!.label.formatter
    expect(fmt({ data: [0, 0, 1500] })).toBe('1.5K')
    expect(fmt({ data: [0, 1, 2_500_000] })).toBe('2.5M')
    expect(fmt({ data: [1, 0, 0.5] })).toBe('0.5')
    expect(fmt({ data: [1, 1, 42] })).toBe('42')
  })

  it('skips holes in sparse series arrays', () => {
    const chartData = makeHeatmapChartData()
    const sparse: typeof chartData.series = []
    sparse[0] = chartData.series[0]!
    sparse[2] = chartData.series[1]!
    chartData.series = sparse
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const series = options.value.series as { data: number[][] }[]
    // only indices 0 and 2 contribute cells
    expect(series[0]!.data.some((d) => d[0] === 0)).toBe(true)
    expect(series[0]!.data.some((d) => d[0] === 2)).toBe(true)
  })
})

describe('useHeatmapChartOptions — grouped 3D', () => {
  it('builds heatmap plus silent scatter legend series for z groups', () => {
    const { options } = useHeatmapChartOptions(
      baseConfig({ chartData: grouped3DHeatmap(), showLabels: true })
    )
    const series = options.value.series as { type: string; name?: string }[]
    expect(series[0]!.type).toBe('heatmap')
    expect(series.slice(1).map((s) => s.name)).toEqual(['zA', 'zB'])
    expect(series.slice(1).every((s) => s.type === 'scatter')).toBe(true)
    expect((options.value.legend as { show?: boolean }).show).toBe(true)
  })

  it('hides legend when only one z value', () => {
    const { options } = useHeatmapChartOptions(
      baseConfig({
        chartData: grouped3DHeatmap({
          zAxis: ['zA'],
          points: [{ xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 }],
          render3D: {
            ...groupedRender3D,
            xValues: ['x1', 'x2'],
            yValues: ['y1', 'y2'],
            zValues: ['zA'],
          },
        }),
      })
    )
    expect((options.value.legend as { show?: boolean }).show).toBe(false)
  })

  it('filters points by visibleZ selection', () => {
    const { options } = useHeatmapChartOptions(
      baseConfig({
        chartData: grouped3DHeatmap(),
        visibleZ: { zA: true, zB: false },
      })
    )
    const series = options.value.series as { data: number[][] }[]
    // x1,y1 was 5+7; with zB hidden only 5 remains
    const cell = series[0]!.data.find((d) => d[0] === 0 && d[1] === 0)
    expect(cell?.[2]).toBe(5)
  })

  it('expands equal min/max when all visible cells share one total', () => {
    const { options } = useHeatmapChartOptions(
      baseConfig({
        chartData: grouped3DHeatmap({
          points: [
            { xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 4 },
            { xAxis: 'x2', yAxis: 'y2', zAxis: 'zB', value: 4 },
          ],
        }),
      })
    )
    const vm = options.value.visualMap as { min: number; max: number }
    expect(vm.max).toBeGreaterThan(vm.min)
  })

  it('uses default axis labels when axisLabels omitted', () => {
    const { options } = useHeatmapChartOptions(
      baseConfig({
        chartData: grouped3DHeatmap({ axisLabels: {} }),
      })
    )
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    const html = tooltip.formatter({ data: [0, 0, 12] })
    expect(html).toContain('x: x1')
    expect(html).toContain('y: y1')
  })

  it('formats 3D tooltip with z breakdown, donut, and margins', () => {
    const { options } = useHeatmapChartOptions(
      baseConfig({ chartData: grouped3DHeatmap(), isDark: true })
    )
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    const html = tooltip.formatter({ value: [0, 0, 12] })
    expect(html).toContain('X: x1 / Y: y1')
    expect(html).toContain('zA:')
    expect(html).toContain('zB:')
    expect(html).toContain('Σ Z:')
    expect(html).toContain('Σ X(x1):')
    expect(html).toContain('Σ Y(y1):')
    expect(html).toContain('<svg')
  })

  it('returns empty tooltip for unknown cell coords', () => {
    const { options } = useHeatmapChartOptions(baseConfig({ chartData: grouped3DHeatmap() }))
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    expect(tooltip.formatter({ data: [99, 99, 0] })).toBe('')
  })

  it('formats 3D cell labels via value or data payloads', () => {
    const { options } = useHeatmapChartOptions(
      baseConfig({ chartData: grouped3DHeatmap(), showLabels: true })
    )
    const series = options.value.series as {
      label: { formatter: (p: any) => string }
    }[]
    const fmt = series[0]!.label.formatter
    expect(fmt({ value: [0, 0, 1500] })).toBe('1.5K')
    expect(fmt({ data: [0, 0, 0] })).toBe('')
    expect(fmt({ data: [0, 0, undefined] })).toBe('')
    expect(fmt({ value: [0, 0, 2_000_000] })).toBe('2.0M')
  })

  it('attaches dataZoom for large grouped axes and rotates long x labels', () => {
    const manyX = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `x${i}`)
    const manyY = Array.from({ length: 3 }, (_, i) => `y${i}`)
    const longX = Array.from({ length: 12 }, (_, i) => `${longLabel}${i}`)
    const { options: largeOpts } = useHeatmapChartOptions(
      baseConfig({
        chartData: grouped3DHeatmap({
          series: manyX.map((x) => ({ xAxis: x, values: [1], benchmarkId: '' })),
          render3D: {
            ...groupedRender3D,
            xValues: manyX,
            yValues: manyY,
            zValues: ['zA', 'zB'],
          },
          points: [{ xAxis: manyX[0]!, yAxis: manyY[0]!, zAxis: 'zA', value: 1 }],
        }),
      })
    )
    expect(largeOpts.value.dataZoom).toBeDefined()

    const { options: rotateOpts } = useHeatmapChartOptions(
      baseConfig({
        chartData: grouped3DHeatmap({
          series: longX.map((x) => ({ xAxis: x, values: [1], benchmarkId: '' })),
          render3D: {
            ...groupedRender3D,
            xValues: longX,
            yValues: ['y1'],
            zValues: ['zA'],
          },
          points: [{ xAxis: longX[0]!, yAxis: 'y1', zAxis: 'zA', value: 1 }],
        }),
      })
    )
    expect((rotateOpts.value.xAxis as { axisLabel?: { rotate?: number } }).axisLabel?.rotate).toBe(
      30
    )
  })

  it('falls back when render3D is missing on grouped 3D data', () => {
    const chartData = grouped3DHeatmap()
    delete (chartData as { render3D?: unknown }).render3D
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const series = options.value.series as { type: string; data: number[][] }[]
    expect(series[0]!.type).toBe('heatmap')
    expect(series[0]!.data).toEqual([])
  })

  it('works without visibleZ ref', () => {
    const cfg = baseConfig({ chartData: grouped3DHeatmap() })
    delete cfg.visibleZ
    const { options } = useHeatmapChartOptions(cfg)
    expect((options.value.series as unknown[]).length).toBeGreaterThan(0)
  })
})

describe('useHeatmapChartOptions — branch edges', () => {
  it('handles null cell values and oob tooltip indices in 2D', () => {
    const chartData = makeHeatmapChartData({
      series: [
        { xAxis: 'a', values: [null, 1], benchmarkId: '' },
        { xAxis: 'b', values: [2, null], benchmarkId: '' },
      ],
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    expect(tooltip.formatter({ data: [99, 99, 1] })).toContain('/')
  })

  it('covers single-z 3D tooltip without sum/donut and missing z rows', () => {
    const chartData = grouped3DHeatmap({
      zAxis: ['zA'],
      points: [{ xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 }],
      render3D: {
        ...groupedRender3D,
        xValues: ['x1', 'x2'],
        yValues: ['y1', 'y2'],
        zValues: ['zA', 'zMissing'],
      },
    })
    delete (chartData as { points?: unknown }).points
    chartData.points = [{ xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 }]
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    const html = tooltip.formatter({ data: [0, 0, 5] })
    expect(html).toContain('zA')
    expect(html).not.toContain('Σ Z:')
    expect(html).not.toContain('<svg')
  })

  it('uses largeY auto interval and missing points default', () => {
    const manyY = Array.from({ length: LARGE_X_THRESHOLD + 1 }, (_, i) => `y${i}`)
    const chartData = grouped3DHeatmap({
      yAxis: manyY,
      render3D: {
        ...groupedRender3D,
        xValues: ['x1'],
        yValues: manyY,
        zValues: ['zA'],
      },
      points: [{ xAxis: 'x1', yAxis: manyY[0]!, zAxis: 'zA', value: 1 }],
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    expect(
      (options.value.yAxis as { axisLabel?: { interval?: string | number } }).axisLabel?.interval
    ).toBe('auto')
  })
})

describe('useHeatmapChartOptions — mocked color fallbacks', () => {
  it('uses default theme color when getNextColorFor returns undefined', async () => {
    const utils = await import('@/lib/utils')
    const spy = vi.spyOn(utils, 'getNextColorFor').mockReturnValue(undefined as unknown as string)
    const { options } = useHeatmapChartOptions(
      baseConfig({ chartData: grouped3DHeatmap(), showLabels: true })
    )
    const legend = options.value.legend as { data: { itemStyle: { color: string } }[] }
    expect(legend.data[0]!.itemStyle.color).toBeTruthy()
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    expect(tooltip.formatter({ data: [0, 0, 12] })).toContain('zA')
    spy.mockRestore()
  })

  it('handles undefined points on 3D heatmap', async () => {
    const utils = await import('@/lib/utils')
    const spy = vi.spyOn(utils, 'is3D').mockReturnValue(false)
    const chartData = grouped3DHeatmap()
    ;(chartData as { points?: unknown }).points = undefined
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    expect((options.value.series as { type: string }[])[0]!.type).toBe('heatmap')
    spy.mockRestore()
  })

  it('tooltip marginal fallbacks for unknown names', () => {
    const { options } = useHeatmapChartOptions(baseConfig({ chartData: grouped3DHeatmap() }))
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    // empty cell still has map entry with total 0
    const html = tooltip.formatter({ data: [1, 1, 0] })
    expect(html).toContain('X: x2')
  })
})

describe('useHeatmapChartOptions — marginal fallbacks', () => {
  it('uses 0 marginals for x/y categories with no points', () => {
    const chartData = grouped3DHeatmap({
      points: [{ xAxis: 'x1', yAxis: 'y1', zAxis: 'zA', value: 5 }],
      render3D: {
        ...groupedRender3D,
        xValues: ['x1', 'ghostX'],
        yValues: ['y1', 'ghostY'],
        zValues: ['zA'],
      },
    })
    const { options } = useHeatmapChartOptions(baseConfig({ chartData }))
    const tooltip = options.value.tooltip as { formatter: (p: unknown) => string }
    // ghost cell has no points → marginal fallbacks
    const html = tooltip.formatter({ data: [1, 1, 0] })
    expect(html).toContain('ghostX')
    expect(html).toContain('ghostY')
  })
})
