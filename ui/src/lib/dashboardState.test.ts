import { describe, expect, it } from 'vitest'
import { keepDetailSkeletonVisible } from './dashboardState'

describe('lazy detail skeleton transition', () => {
  it('stays visible through network loading and worker slot initialization', () => {
    expect(
      keepDetailSkeletonVisible({
        lazy: true,
        detailLoading: true,
        detailError: null,
        hasDetailData: false,
        chartCount: 0,
      })
    ).toBe(true)

    expect(
      keepDetailSkeletonVisible({
        lazy: true,
        detailLoading: false,
        detailError: null,
        hasDetailData: true,
        chartCount: 0,
      })
    ).toBe(true)

    expect(
      keepDetailSkeletonVisible({
        lazy: true,
        detailLoading: false,
        detailError: null,
        hasDetailData: true,
        chartCount: 1,
      })
    ).toBe(false)
  })

  it('gives a detail error precedence over the skeleton', () => {
    expect(
      keepDetailSkeletonVisible({
        lazy: true,
        detailLoading: false,
        detailError: 'request failed',
        hasDetailData: false,
        chartCount: 0,
      })
    ).toBe(false)
  })
})
