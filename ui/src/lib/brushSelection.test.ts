import { describe, expect, it } from 'vitest'
import { brushSelectionStats } from './brushSelection'

describe('brushSelectionStats', () => {
  it('aggregates multiple regions across grouped bar series', () => {
    const stats = brushSelectionStats(
      {
        series: [
          { type: 'bar', data: [112, 242, 417] },
          { type: 'bar', data: [118, 233, 391] },
        ],
      },
      {
        batch: [
          {
            areas: [{}, {}],
            selected: [
              { seriesIndex: 0, dataIndex: [0, 2] },
              { seriesIndex: 1, dataIndex: [1, 2] },
            ],
          },
        ],
      }
    )

    expect(stats).toEqual({ regions: 2, total: 1153, average: 288.25, count: 4 })
  })
})
