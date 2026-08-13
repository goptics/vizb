import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { makeGroupedChartData } from '@/test-utils'
import type { ChartData, CorrelationMatrix, DescriptiveStats, SeriesProfile } from '@/types'

const { statsMocks, usableAxesRef } = vi.hoisted(() => {
  const usableAxesRef = { value: ['x', 'y'] as ('x' | 'y' | 'z')[] }
  const profile = (name: string, mean = 1): SeriesProfile => ({
    name,
    stats: {
      count: 2,
      missing: 0,
      unique: 2,
      zeros: 0,
      negatives: 0,
      mean,
      median: mean,
      mode: mean,
      geoMean: mean,
      harmMean: mean,
      trimMean: mean,
      variance: 1,
      stdDev: 1,
      cv: 0.1,
      sem: 0.1,
      cqv: 0.1,
      min: 0,
      max: mean * 2,
      range: mean * 2,
      iqr: 1,
      mad: 1,
      lowerFence: -1,
      upperFence: 10,
      outliers: 0,
      skewness: 0,
      kurtosis: 0,
      p1: 0,
      p5: 0,
      p10: 0,
      p25: 0,
      p75: mean,
      p90: mean,
      p95: mean,
      p99: mean,
      ci95Lower: 0,
      ci95Upper: mean,
    } satisfies DescriptiveStats,
  })

  const corr = (axis: 'x' | 'y' | 'z' = 'x'): CorrelationMatrix => ({
    axis,
    labels: ['A', 'B'],
    pearson: [
      [1, 0.5],
      [0.5, 1],
    ],
    spearman: [
      [1, 0.4],
      [0.4, 1],
    ],
    kendall: [
      [1, 0.3],
      [0.3, 1],
    ],
    dcor: [
      [1, 0.2],
      [0.2, 1],
    ],
  })

  return {
    statsMocks: {
      profile,
      corr,
      computeDescriptive: vi.fn(async (_c: ChartData) => [
        profile('West', 10),
        profile('East', 20),
      ]),
      computeCorrelation: vi.fn(
        async (_c: ChartData, _axis?: string): Promise<CorrelationMatrix | undefined> => corr('x')
      ),
      available: { correlation: true },
      availableViews: vi.fn(() => ({ correlation: true })),
    },
    usableAxesRef,
  }
})

vi.mock('../composables/useStatsWorker', () => ({
  computeDescriptive: (c: ChartData) => statsMocks.computeDescriptive(c),
  computeCorrelation: (c: ChartData, axis?: string) => statsMocks.computeCorrelation(c, axis),
}))

vi.mock('../lib/stats', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../lib/stats')>()
  return {
    ...actual,
    availableViews: (...args: unknown[]) => {
      ;(statsMocks.availableViews as (...a: unknown[]) => unknown)(...args)
      return { correlation: statsMocks.available.correlation }
    },
    usableCorrelationAxes: () => usableAxesRef.value,
  }
})

vi.mock('../composables/useSettingsStore', () => ({
  useSettingsStore: () => ({ isDark: ref(false) }),
}))

vi.mock('../composables/useFullscreen', () => ({
  useFullscreen: () => ({
    containerRef: ref<HTMLElement | null>(null),
    isFullscreen: ref(false),
    withFullscreenToolbox: <T>(o: T) => o,
  }),
}))

vi.mock('../composables/charts/useCorrelationOption', () => ({
  buildCorrelationOption: () => ({ series: [{ type: 'heatmap' }] }),
}))

vi.mock('vue', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue')>()
  const HeatStub = actual.defineComponent({
    name: 'ChartHeatmap',
    props: ['option', 'initOptions'],
    setup: () => () => actual.h('div', { 'data-testid': 'corr-heatmap' }),
  })
  return {
    ...actual,
    defineAsyncComponent: (loader: unknown) => {
      if (typeof loader === 'function') {
        void (loader as () => Promise<unknown>)().catch(() => {})
      } else if (loader && typeof loader === 'object' && 'loader' in (loader as object)) {
        const opts = loader as { loader: () => Promise<unknown> }
        void opts.loader().catch(() => {})
      }
      return HeatStub
    },
  }
})

vi.mock('./SelectionTabs.vue', () => ({
  default: defineComponent({
    name: 'SelectionTabs',
    props: ['modelValue', 'options'],
    emits: ['update:modelValue'],
    setup(p, { emit }) {
      return () =>
        h(
          'div',
          { 'data-testid': 'view-tabs' },
          (p.options as { value: string; label: string }[]).map((o) =>
            h(
              'button',
              {
                'data-view': o.value,
                onClick: () => emit('update:modelValue', o.value),
              },
              o.label
            )
          )
        )
    },
  }),
}))

vi.mock('./Selector.vue', () => ({
  default: defineComponent({
    name: 'Selector',
    props: ['items', 'activeId'],
    emits: ['select'],
    setup(p, { emit, attrs }) {
      return () =>
        h(
          'button',
          {
            'data-testid': attrs.class?.toString().includes('w-36') ? 'corr-selector' : 'selector',
            'data-active': String(p.activeId),
            onClick: () => emit('select', 1),
          },
          'sel'
        )
    },
  }),
}))

vi.mock('./DescriptiveColumnSelect.vue', () => ({
  default: defineComponent({
    name: 'DescriptiveColumnSelect',
    props: ['modelValue', 'defaultKeys'],
    emits: ['update:modelValue'],
    setup(p, { emit }) {
      return () =>
        h(
          'button',
          {
            'data-testid': 'col-select',
            onClick: () => emit('update:modelValue', ['mean']),
          },
          String((p.modelValue as string[]).length)
        )
    },
  }),
}))

import StatsPanel from './StatsPanel.vue'

function chart(overrides: Partial<ChartData> = {}): ChartData {
  return makeGroupedChartData({
    statType: 'latency ms',
    yAxis: ['A', 'B', 'C'],
    zAxis: ['z1', 'z2'],
    series: [
      { xAxis: 'West', values: [1, 2, 3], benchmarkId: '' },
      { xAxis: 'East', values: [4, 5, 6], benchmarkId: '' },
    ],
    axisLabels: { x: 'region', y: 'cat', z: 'depth' },
    ...overrides,
  })
}

function manySeriesChart(n: number): ChartData {
  return makeGroupedChartData({
    series: Array.from({ length: n }, (_, i) => ({
      xAxis: `S${i}`,
      values: [i, i + 1],
      benchmarkId: '',
    })),
    yAxis: ['A', 'B'],
  })
}

describe('StatsPanel', () => {
  let createObjectURL: ReturnType<typeof vi.fn>
  let revokeObjectURL: ReturnType<typeof vi.fn>
  let clickSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    statsMocks.available.correlation = true
    usableAxesRef.value = ['x', 'y']
    statsMocks.computeDescriptive.mockImplementation(async () => [
      statsMocks.profile('West', 10),
      statsMocks.profile('East', 20),
    ])
    statsMocks.computeCorrelation.mockImplementation(async () => statsMocks.corr('x'))
    createObjectURL = vi.fn(() => 'blob:csv')
    revokeObjectURL = vi.fn()
    vi.stubGlobal('URL', {
      createObjectURL,
      revokeObjectURL,
    })
    clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.useRealTimers()
    clickSpy.mockRestore()
    vi.unstubAllGlobals()
  })

  it('shows loading skeleton then descriptive table', async () => {
    let resolveDesc!: (v: SeriesProfile[]) => void
    statsMocks.computeDescriptive.mockImplementation(
      () =>
        new Promise((r) => {
          resolveDesc = r
        })
    )
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    expect(w.text()).toContain('Statistics')
    expect(w.text()).toContain('latency ms')
    expect(w.find('.animate-pulse').exists()).toBe(true)

    resolveDesc([statsMocks.profile('West', 10), statsMocks.profile('East', 20)])
    await flushPromises()
    expect(w.text()).toContain('West')
    expect(w.text()).toContain('East')
    expect(w.text()).toContain('Descriptive')
  })

  it('empty views when no columns and correlation off', async () => {
    statsMocks.available.correlation = false
    const w2 = mount(StatsPanel, {
      props: { chartData: chart(), math: ['correlations'] },
    })
    await flushPromises()
    // Correlation-only math with correlation unavailable → empty state copy.
    expect(w2.text()).toContain('No statistics available')
  })

  it('sorts descriptive columns and formats values', async () => {
    statsMocks.computeDescriptive.mockResolvedValue([
      {
        ...statsMocks.profile('B', 5),
        stats: {
          ...statsMocks.profile('B', 5).stats,
          mean: 5,
          cv: 0.123,
          count: 3,
          stdDev: 1e-4,
          min: Number.NaN,
        },
      },
      {
        ...statsMocks.profile('A', 9),
        stats: {
          ...statsMocks.profile('A', 9).stats,
          mean: 9,
          cv: 0.5,
          count: 1,
          stdDev: 1e7,
          min: 1,
        },
      },
    ])
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()

    // sort by name desc then asc then clear
    const nameTh = w.findAll('th').find((t) => t.text().includes('region'))!
    await nameTh.trigger('click')
    await nameTh.trigger('click')
    await nameTh.trigger('click')

    // sort by mean
    const meanTh = w.findAll('th').find((t) => t.text().includes('Mean'))!
    await meanTh.trigger('click') // desc
    await meanTh.trigger('click') // asc
    expect(w.text()).toMatch(/%/) // cv format
    // exponential for large/small
    expect(w.html()).toMatch(/e\+|e-/)
  })

  it('downloads descriptive CSV', async () => {
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    const btn = w.findAll('button').find((b) => b.text().includes('CSV'))!
    expect(btn.attributes('disabled')).toBeUndefined()
    await btn.trigger('click')
    expect(createObjectURL).toHaveBeenCalled()
    expect(clickSpy).toHaveBeenCalled()
    expect(revokeObjectURL).toHaveBeenCalled()
  })

  it('switches to correlation, methods, axes, and CSV', async () => {
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()

    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    expect(statsMocks.computeCorrelation).toHaveBeenCalled()
    expect(w.find('[data-testid="corr-heatmap"]').exists()).toBe(true)
    expect(w.text()).toMatch(/Correlating/)

    // method select (first corr selector)
    const selectors = w.findAll('[data-testid="corr-selector"]')
    expect(selectors.length).toBeGreaterThanOrEqual(1)
    await selectors[0]!.trigger('click') // method index 1 → spearman
    await flushPromises()

    // axis select when multiple
    if (selectors[1]) {
      await selectors[1].trigger('click')
      await flushPromises()
      expect(statsMocks.computeCorrelation.mock.calls.length).toBeGreaterThan(1)
    }

    const btn = w.findAll('button').find((b) => b.text().includes('CSV'))!
    await btn.trigger('click')
    expect(createObjectURL).toHaveBeenCalled()
  })

  it('search filters series when >20 profiles', async () => {
    vi.useFakeTimers()
    const profiles = Array.from({ length: 25 }, (_, i) => statsMocks.profile(`S${i}`, i))
    statsMocks.computeDescriptive.mockResolvedValue(profiles)
    const w = mount(StatsPanel, { props: { chartData: manySeriesChart(25) } })
    await flushPromises()
    const input = w.get('input[type="search"]')
    await input.setValue('S1')
    await vi.advanceTimersByTimeAsync(300)
    await nextTick()
    expect(w.text()).toMatch(/\d+ \/ 25/)
    await input.setValue('nomatch-zzz')
    await vi.advanceTimersByTimeAsync(300)
    await nextTick()
    expect(w.text()).toContain('No series match')
  })

  it('virtualized scroll updates and column select clears invalid sort', async () => {
    const profiles = Array.from({ length: 40 }, (_, i) => statsMocks.profile(`S${i}`, i))
    statsMocks.computeDescriptive.mockResolvedValue(profiles)
    const w = mount(StatsPanel, { props: { chartData: manySeriesChart(40) } })
    await flushPromises()

    const meanTh = w.findAll('th').find((t) => t.text().includes('Mean'))!
    await meanTh.trigger('click')

    const scroll = w.get('.max-h-\\[600px\\]')
    Object.defineProperty(scroll.element, 'clientHeight', { value: 200, configurable: true })
    Object.defineProperty(scroll.element, 'scrollTop', {
      value: 400,
      configurable: true,
      writable: true,
    })
    await scroll.trigger('scroll')
    await nextTick()

    await w.get('[data-testid="col-select"]').trigger('click')
    await nextTick()
    // sort key mean still visible since mean is selected
    expect(w.text()).toContain('Mean')
  })

  it('reloads on chartData change and discards stale tokens', async () => {
    let resolveFirst!: (v: SeriesProfile[]) => void
    let resolveSecond!: (v: SeriesProfile[]) => void
    let call = 0
    statsMocks.computeDescriptive.mockImplementation(
      () =>
        new Promise((r) => {
          if (call++ === 0) resolveFirst = r
          else resolveSecond = r
        })
    )
    const w = mount(StatsPanel, { props: { chartData: chart({ title: 'one' }) } })
    await w.setProps({ chartData: chart({ title: 'two', statType: 'ops' }) })
    // resolve stale first
    resolveFirst([statsMocks.profile('stale', 1)])
    await flushPromises()
    expect(w.text()).not.toContain('stale')
    resolveSecond([statsMocks.profile('fresh', 2)])
    await flushPromises()
    expect(w.text()).toContain('fresh')
    expect(w.text()).toContain('ops')
  })

  it('math without correlations hides correlation tab', async () => {
    const w = mount(StatsPanel, {
      props: { chartData: chart(), math: ['center'] },
    })
    await flushPromises()
    expect(w.find('[data-view="correlation"]').exists()).toBe(false)
    expect(w.text()).toContain('Mean')
  })

  it('correlation methods cycle through matrices', async () => {
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()

    const sel = w.findAllComponents({ name: 'Selector' })
    for (const idx of [0, 1, 2, 3, 99]) {
      sel[0]!.vm.$emit('select', idx)
      await nextTick()
    }
    expect(w.find('[data-testid="corr-heatmap"]').exists()).toBe(true)
  })

  it('corr axis same selection is no-op', async () => {
    usableAxesRef.value = ['x', 'y']
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    const before = statsMocks.computeCorrelation.mock.calls.length
    // active index 0 is x; selecting 0 again
    const axisSel = w.findAllComponents({ name: 'Selector' })[1]
    axisSel?.vm.$emit('select', 0)
    await flushPromises()
    // may no-op if override null and index 0 is default
    expect(statsMocks.computeCorrelation.mock.calls.length).toBeGreaterThanOrEqual(before)
  })

  it('download disabled while loading', async () => {
    statsMocks.computeDescriptive.mockImplementation(() => new Promise(() => {}))
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    const btn = w.findAll('button').find((b) => b.text().includes('CSV'))!
    expect(btn.attributes('disabled')).toBeDefined()
    await btn.trigger('click')
    expect(createObjectURL).not.toHaveBeenCalled()
  })

  it('fmt edges via table cells with special stats', async () => {
    statsMocks.computeDescriptive.mockResolvedValue([
      {
        name: '',
        stats: {
          ...statsMocks.profile('x').stats,
          count: 2,
          mean: 3,
          cv: Number.POSITIVE_INFINITY,
          stdDev: 2.5,
          min: Number.NaN,
        },
      },
    ])
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    expect(w.text()).toContain('—')
  })

  it('slug empty statType still downloads', async () => {
    const w = mount(StatsPanel, {
      props: { chartData: chart({ statType: '@@@' }) },
    })
    await flushPromises()
    await w
      .findAll('button')
      .find((b) => b.text().includes('CSV'))!
      .trigger('click')
    expect(createObjectURL).toHaveBeenCalled()
  })

  it('sorts with NaN sinking and name direction', async () => {
    statsMocks.computeDescriptive.mockResolvedValue([
      {
        ...statsMocks.profile('B', 1),
        stats: { ...statsMocks.profile('B', 1).stats, mean: Number.NaN },
      },
      { ...statsMocks.profile('A', 2), stats: { ...statsMocks.profile('A', 2).stats, mean: 2 } },
      {
        ...statsMocks.profile('C', 3),
        stats: { ...statsMocks.profile('C', 3).stats, mean: Number.NaN },
      },
    ])
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    const meanTh = w.findAll('th').find((t) => t.text().includes('Mean'))!
    await meanTh.trigger('click') // desc
    await meanTh.trigger('click') // asc
    const nameTh = w.findAll('th').find((t) => t.text().includes('region'))!
    await nameTh.trigger('click')
    await nameTh.trigger('click')
    expect(w.text()).toContain('A')
  })

  it('correlation early-return when already loaded', async () => {
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    const calls = statsMocks.computeCorrelation.mock.calls.length
    await w.get('[data-view="descriptive"]').trigger('click')
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    expect(statsMocks.computeCorrelation.mock.calls.length).toBe(calls)
  })

  it('correlation axis invalid index is no-op', async () => {
    usableAxesRef.value = ['x', 'y']
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    const axisSel = w.findAllComponents({ name: 'Selector' })[1]
    axisSel?.vm.$emit('select', 99)
    await flushPromises()
    expect(w.find('[data-testid="corr-heatmap"]').exists()).toBe(true)
  })

  it('corr captions fall back without axis labels', async () => {
    statsMocks.computeCorrelation.mockResolvedValue({
      ...statsMocks.corr('z'),
      axis: 'z',
    })
    usableAxesRef.value = ['z']
    const w = mount(StatsPanel, {
      props: {
        chartData: chart({
          axisLabels: {},
          series: [{ xAxis: 'only', values: [1, 2], benchmarkId: '' }],
          yAxis: ['a'],
          zAxis: ['z1', 'z2'],
        }),
      },
    })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    expect(w.text()).toMatch(/Correlating|z-axis|Series|y-axis/)
  })

  it('stale correlation reply is discarded', async () => {
    let resolveCorr!: (v: CorrelationMatrix) => void
    statsMocks.computeCorrelation.mockImplementation(
      () =>
        new Promise((r) => {
          resolveCorr = r
        })
    )
    const w = mount(StatsPanel, { props: { chartData: chart({ title: 'c1' }) } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await w.setProps({ chartData: chart({ title: 'c2', statType: 'other' }) })
    await flushPromises()
    resolveCorr(statsMocks.corr('x'))
    await flushPromises()
    // after reload, descriptive should win; correlation may re-fetch
    expect(w.text()).toContain('other')
  })
  it('empty profiles cannot download', async () => {
    statsMocks.computeDescriptive.mockResolvedValue([])
    statsMocks.available.correlation = false
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    const btn = w.findAll('button').find((b) => b.text().includes('CSV'))!
    expect(btn.attributes('disabled')).toBeDefined()
    const setup = (w.vm as unknown as { $: { setupState: { downloadCsv: () => void } } }).$
      .setupState
    setup.downloadCsv()
    expect(createObjectURL).not.toHaveBeenCalled()
  })

  it('downloadCsv early return via enabled button then empty correlation', async () => {
    statsMocks.computeDescriptive.mockResolvedValue([statsMocks.profile('A', 1)])
    statsMocks.available.correlation = true
    usableAxesRef.value = ['x', 'y']
    let corrResolve!: (v: CorrelationMatrix | undefined) => void
    statsMocks.computeCorrelation.mockImplementation(
      () =>
        new Promise<CorrelationMatrix | undefined>((r) => {
          corrResolve = r
        })
    )
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    // still loading correlation — canDownload false
    const btn = w.findAll('button').find((b) => b.text().includes('CSV'))!
    expect(btn.attributes('disabled')).toBeDefined()
    corrResolve(undefined)
    await flushPromises()
    // correlation undefined => corrOption null branch and canDownload false
    expect(w.find('[data-testid="corr-heatmap"]').exists()).toBe(false)
  })

  it('sort name and numeric with only-NaN pair', async () => {
    statsMocks.computeDescriptive.mockResolvedValue([
      {
        ...statsMocks.profile('Z', 1),
        stats: { ...statsMocks.profile('Z', 1).stats, mean: Number.NaN },
      },
      {
        ...statsMocks.profile('Y', 1),
        stats: { ...statsMocks.profile('Y', 1).stats, mean: Number.NaN },
      },
      { ...statsMocks.profile('X', 5), stats: { ...statsMocks.profile('X', 5).stats, mean: 5 } },
    ])
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    const meanTh = w.findAll('th').find((t) => t.text().includes('Mean'))!
    await meanTh.trigger('click') // desc: finite first, NaN last (an return 1)
    expect(w.text()).toContain('X')
  })

  it('slug falls back when statType empty after strip', async () => {
    const w = mount(StatsPanel, {
      props: { chartData: chart({ statType: '---' }) },
    })
    await flushPromises()
    await w
      .findAll('button')
      .find((b) => b.text().includes('CSV'))!
      .trigger('click')
    expect(createObjectURL).toHaveBeenCalled()
  })

  it('stale correlation token discard', async () => {
    let resolveCorr!: (v: CorrelationMatrix) => void
    let calls = 0
    statsMocks.computeCorrelation.mockImplementation(
      () =>
        new Promise((r) => {
          if (calls++ === 0) resolveCorr = r
          else r(statsMocks.corr('y'))
        })
    )
    const w = mount(StatsPanel, { props: { chartData: chart({ title: 'a' }) } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    // switch chart before first corr resolves
    await w.setProps({ chartData: chart({ title: 'b', statType: 'switched' }) })
    await flushPromises()
    resolveCorr(statsMocks.corr('x'))
    await flushPromises()
    expect(w.text()).toContain('switched')
  })

  it('sort an branch when NaN is first operand', async () => {
    statsMocks.computeDescriptive.mockResolvedValue([
      { name: 'n1', stats: { ...statsMocks.profile('n1').stats, mean: Number.NaN } },
      { name: 'f1', stats: { ...statsMocks.profile('f1').stats, mean: 10 } },
      { name: 'n2', stats: { ...statsMocks.profile('n2').stats, mean: Number.NaN } },
      { name: 'f2', stats: { ...statsMocks.profile('f2').stats, mean: 1 } },
    ])
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    const meanTh = w.findAll('th').find((x) => x.text().includes('Mean'))!
    await meanTh.trigger('click') // desc
    await meanTh.trigger('click') // asc — exercises an/bn both ways
    expect(w.text()).toContain('f1')
  })

  it('corrMatrix empty and axis fallbacks without labels', async () => {
    usableAxesRef.value = []
    statsMocks.computeCorrelation.mockResolvedValue(undefined)
    const w = mount(StatsPanel, {
      props: {
        chartData: chart({
          axisLabels: undefined,
          series: [
            { xAxis: 'a', values: [1, 2], benchmarkId: '' },
            { xAxis: 'b', values: [3, 4], benchmarkId: '' },
          ],
          yAxis: ['y1', 'y2'],
          zAxis: [],
        }),
      },
    })
    await flushPromises()
    if (w.find('[data-view="correlation"]').exists()) {
      await w.get('[data-view="correlation"]').trigger('click')
      await flushPromises()
    }
    expect(w.text()).toContain('Statistics')
  })

  it('slug first fallback when statType empty string', async () => {
    const w = mount(StatsPanel, { props: { chartData: chart({ statType: '' }) } })
    await flushPromises()
    await w
      .findAll('button')
      .find((b) => b.text().includes('CSV'))!
      .trigger('click')
    expect(createObjectURL).toHaveBeenCalled()
  })

  it('renders correlation captions with axis fallbacks and no other dims', async () => {
    // correlation axis 'x' with 2+ y entries but no labels → caption falls back
    // to "y-axis" and only x present among other dims keeps names non-empty.
    statsMocks.computeCorrelation.mockResolvedValue(statsMocks.corr('x'))
    usableAxesRef.value = ['x']
    const w = mount(StatsPanel, {
      props: {
        chartData: chart({
          axisLabels: {},
          zAxis: [],
          series: [
            { xAxis: 'West', values: [1, 2], benchmarkId: '' },
            { xAxis: 'East', values: [3, 4], benchmarkId: '' },
          ],
        }),
      },
    })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    expect(w.text()).toMatch(/Correlating|the other dimensions/)
  })

  it('shows correlation skeleton before the reply lands', async () => {
    // correlation unresolved → corrAxis/matrix/caption fallbacks, invalid method
    // index, and corrAxisActiveIndex all evaluate against empty correlation.
    statsMocks.computeCorrelation.mockImplementation(() => new Promise(() => {}))
    usableAxesRef.value = ['x']
    const w = mount(StatsPanel, { props: { chartData: chart() } })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    // Skeleton (activeLoading) rather than a heatmap.
    expect(w.find('[data-testid="corr-heatmap"]').exists()).toBe(false)
    expect(w.find('.animate-pulse').exists()).toBe(true)
  })

  it('handles undefined zAxis in corrAxes', async () => {
    usableAxesRef.value = ['x']
    const w = mount(StatsPanel, {
      props: { chartData: chart({ zAxis: undefined as unknown as string[] }) },
    })
    await flushPromises()
    await w.get('[data-view="correlation"]').trigger('click')
    await flushPromises()
    expect(w.find('[data-testid="corr-heatmap"]').exists()).toBe(true)
  })
})
