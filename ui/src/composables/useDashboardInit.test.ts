import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick, ref, type Ref } from 'vue'
import type { Dataset } from '../types'
import { ds } from '../test-utils'

const holder = vi.hoisted(() => ({
  datasets: undefined as Ref<Dataset[]> | undefined,
  activeDataset: undefined as Ref<Dataset | undefined> | undefined,
  initFromUrl: vi.fn(async () => {}),
}))

vi.mock('./useDataPoint', () => ({
  useDataPoint: () => ({
    get datasets() {
      if (!holder.datasets) throw new Error('forgot beforeEach datasets')
      return holder.datasets
    },
    get activeDataset() {
      if (!holder.activeDataset) throw new Error('forgot beforeEach activeDataset')
      return holder.activeDataset
    },
  }),
}))

vi.mock('./useUrlRouter', () => ({
  useUrlRouter: () => ({
    initFromUrl: holder.initFromUrl,
  }),
}))

describe('useDashboardInit', () => {
  beforeEach(() => {
    vi.resetModules()
    holder.datasets = ref<Dataset[]>([])
    holder.activeDataset = ref<Dataset | undefined>(undefined)
    holder.initFromUrl.mockReset()
    holder.initFromUrl.mockResolvedValue(undefined)
    // Unit project is node — stub the only DOM surface this composable touches.
    vi.stubGlobal('document', { title: 'Vizb' })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('calls initFromUrl once when datasets become non-empty', async () => {
    const { useDashboardInit } = await import('./useDashboardInit')
    useDashboardInit()

    expect(holder.initFromUrl).not.toHaveBeenCalled()

    holder.datasets!.value = [ds([{ type: 'bar' }])]
    await nextTick()

    expect(holder.initFromUrl).toHaveBeenCalledTimes(1)
  })

  it('sets document.title from activeDataset name', async () => {
    holder.activeDataset = ref(ds([{ type: 'bar' }]))
    holder.activeDataset.value = { ...holder.activeDataset.value!, name: 'Sales' }

    const { useDashboardInit } = await import('./useDashboardInit')
    useDashboardInit()
    await nextTick()

    expect(document.title).toBe('Vizb | Sales')
  })

  it('does not call initFromUrl again on a second datasets update', async () => {
    holder.datasets = ref([
      ds([{ type: 'bar' }], [{ name: 'a', stats: [{ type: 'ns', value: 1 }] }]),
    ])

    const { useDashboardInit } = await import('./useDashboardInit')
    useDashboardInit()
    await nextTick()

    expect(holder.initFromUrl).toHaveBeenCalledTimes(1)

    holder.datasets!.value = [
      ds([{ type: 'bar' }], [{ name: 'a', stats: [{ type: 'ns', value: 1 }] }]),
      { name: 'Second', settings: [{ type: 'line' }], data: [] },
    ]
    await nextTick()

    expect(holder.initFromUrl).toHaveBeenCalledTimes(1)
  })
})
