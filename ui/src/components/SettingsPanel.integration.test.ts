import { describe, it, expect, vi } from 'vitest'
import { computed, ref } from 'vue'
import { mount } from '@vue/test-utils'
import type { BarConfig } from '@/types'
// Side-effect: top-level vi.mock for every settings control SFC.
import '@/test-utils/mockSettingsControls'

vi.mock('./Selector.vue', () => ({
  default: { name: 'Selector', template: '<div data-testid="selector-stub" />' },
}))

vi.mock('./SettingHeader.vue', () => ({
  default: {
    name: 'SettingHeader',
    props: ['label', 'description'],
    template: '<div data-testid="setting-header">{{ label }}</div>',
  },
}))

const barConfig = ref<BarConfig>({
  type: 'bar',
  sort: { enabled: false, order: 'asc' },
  scale: 'linear',
  stack: false,
  showLabels: false,
  swap: 'x/y',
})

vi.mock('@/composables/useSettingsStore', () => ({
  useSettingsStore: () => ({
    activeConfig: barConfig,
    chartType: ref('bar'),
    setChartType: vi.fn(),
    setSort: vi.fn(),
    setScale: vi.fn(),
    setStack: vi.fn(),
    setShowLabels: vi.fn(),
    setSmooth: vi.fn(),
    setHorizontal: vi.fn(),
    setThreeDRotate: vi.fn(),
    setSwap: vi.fn(),
    setThreeD: vi.fn(),
    setThreeDVisualMap: vi.fn(),
    setVisualMap: vi.fn(),
  }),
}))

vi.mock('@/composables/useDataPoint', () => ({
  useDataPoint: () => ({
    activeDataset: computed(() => ({
      settings: [barConfig.value],
      data: [],
      axes: [],
    })),
    activeDatasetId: ref('ds-1'),
    activeDataDimension: computed(() => '2D' as const),
    activeArrangement: computed(() => ({ targetString: 'x/y' })),
    getArrangement: vi.fn(() => undefined),
    setArrangement: vi.fn(),
    activeGroupId: ref(0),
    isValueMode: computed(() => false),
    chartMode: computed(() => 'grouped' as const),
  }),
}))

import SettingsPanel from './SettingsPanel.vue'

describe('SettingsPanel', () => {
  it('renders Settings card title', () => {
    const w = mount(SettingsPanel)
    expect(w.text()).toContain('Settings')
  })

  it('mounts getRenderableFields-driven bar 2D controls', () => {
    const w = mount(SettingsPanel)

    // Core bar 2D fields from the registry (not the full matrix).
    for (const name of [
      'SortControl',
      'ScaleControl',
      'StackControl',
      'ShowLabelsControl',
      'HorizontalControl',
      'SwapControl',
    ]) {
      expect(w.findComponent({ name }).exists()).toBe(true)
    }
  })
})
