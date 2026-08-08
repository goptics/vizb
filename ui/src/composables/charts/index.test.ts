import { describe, it, expect } from 'vitest'
import * as charts from './index'

describe('composables/charts index', () => {
  it('re-exports chart option composables and shared surface', () => {
    expect(typeof charts.useBarChartOptions).toBe('function')
    expect(typeof charts.useLineChartOptions).toBe('function')
    expect(typeof charts.usePieChartOptions).toBe('function')
    expect(typeof charts.useHeatmapChartOptions).toBe('function')
    expect(typeof charts.useRadarChartOptions).toBe('function')
    expect(typeof charts.useSankeyChartOptions).toBe('function')
    expect(typeof charts.useBar3DChartOptions).toBe('function')
    expect(typeof charts.useLine3DChartOptions).toBe('function')
    expect(typeof charts.useScatterChartOptions).toBe('function')
    expect(typeof charts.useScatter3DChartOptions).toBe('function')
    expect(typeof charts.getBaseOptions).toBe('function')
    expect(typeof charts.buildValueAxes2DOptions).toBe('function')
    expect(typeof charts.resolveSeriesSymbol).toBe('function')
  })
})
