import { describe, it, expect } from 'vitest'
import { BASE_2D } from './base'

describe('BASE_2D', () => {
  it('exports the shared 2D echarts module list', () => {
    expect(Array.isArray(BASE_2D)).toBe(true)
    expect(BASE_2D).toHaveLength(6)
    for (const mod of BASE_2D) {
      expect(mod).toBeTruthy()
    }
  })
})
