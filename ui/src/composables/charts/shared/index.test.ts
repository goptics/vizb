import { describe, it, expect } from 'vitest'
import * as shared from './index'

describe('charts/shared index', () => {
  it('re-exports shared chart helpers', () => {
    expect(typeof shared.resolveSeriesSymbol).toBe('function')
    expect(typeof shared.buildValueAxes2DOptions).toBe('function')
    expect(typeof shared.create3DGridConfig).toBe('function')
    expect(typeof shared.getChartStyling).toBe('function')
  })
})
