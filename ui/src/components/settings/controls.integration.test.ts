import { describe, it, expect, vi, beforeEach } from 'vitest'
import { computed, defineComponent, h, ref } from 'vue'
import { mount } from '@vue/test-utils'
import type { Sort, ScaleType } from '@/types'

vi.mock('../ui', () => {
  const passthrough = (name: string) =>
    defineComponent({
      name,
      setup(_, { slots }) {
        return () => h('div', { 'data-stub': name }, slots.default?.())
      },
    })
  return {
    Separator: passthrough('Separator'),
    Label: defineComponent({
      name: 'Label',
      props: ['for'],
      setup(p, { slots }) {
        return () => h('label', { for: p.for }, slots.default?.())
      },
    }),
    Switch: defineComponent({
      name: 'Switch',
      props: ['id', 'checked'],
      emits: ['update:checked'],
      setup(p, { emit }) {
        return () =>
          h('button', {
            'data-testid': `switch-${p.id}`,
            onClick: () => emit('update:checked', !p.checked),
          })
      },
    }),
  }
})

vi.mock('../SelectionTabs.vue', () => ({
  default: defineComponent({
    name: 'SelectionTabs',
    props: ['modelValue', 'options', 'disabled'],
    emits: ['update:modelValue'],
    setup(p, { emit }) {
      return () =>
        h(
          'div',
          { 'data-testid': 'selection-tabs' },
          (p.options as { value: string; label: string }[]).map((o) =>
            h(
              'button',
              {
                'data-value': o.value,
                disabled: p.disabled,
                onClick: () => emit('update:modelValue', o.value),
              },
              o.label
            )
          )
        )
    },
  }),
}))

vi.mock('../Selector.vue', () => ({
  default: defineComponent({
    name: 'Selector',
    props: ['items', 'activeId'],
    emits: ['select'],
    setup(p, { emit }) {
      return () =>
        h(
          'button',
          {
            'data-testid': 'swap-selector',
            onClick: () => emit('select', 1),
          },
          String(p.activeId)
        )
    },
  }),
}))

vi.mock('../SettingHeader.vue', () => ({
  default: defineComponent({
    name: 'SettingHeader',
    props: ['label', 'description', 'id'],
    setup(p) {
      return () => h('div', { 'data-testid': 'setting-header' }, p.label)
    },
  }),
}))

vi.mock('../SettingsToggle.vue', async () => {
  return {
    default: defineComponent({
      name: 'SettingsToggle',
      props: ['id', 'label', 'description', 'checked'],
      emits: ['update:checked'],
      setup(p, { emit }) {
        return () =>
          h(
            'button',
            {
              'data-testid': p.id,
              onClick: () => emit('update:checked', !p.checked),
            },
            p.label
          )
      },
    }),
  }
})

const dpState = vi.hoisted(() => ({
  data: [] as { xAxis?: string; yAxis?: string; zAxis?: string }[],
  axes: undefined as { key: string; type?: string }[] | undefined,
  isValueMode: false,
  targetString: 'x/y',
}))

vi.mock('@/composables/useDataPoint', () => ({
  useDataPoint: () => ({
    activeDataset: computed(() => ({
      data: dpState.data,
      axes: dpState.axes,
    })),
    activeArrangement: computed(() => ({ targetString: dpState.targetString })),
    isValueMode: computed(() => dpState.isValueMode),
  }),
}))

vi.mock('@/composables/useSettingsStore', () => ({
  useSettingsStore: () => ({
    activeConfig: computed(() => ({ threeD: false })),
  }),
}))

vi.mock('@/lib/swap', () => ({
  swapOptionKeys: () => ['x/y', 'y/x', 'x/y/z'],
}))

vi.mock('@/lib/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/utils')>()
  return {
    ...actual,
    bundleHas3DChunk: () => false,
  }
})

import SortControl from './SortControl.vue'
import ScaleControl from './ScaleControl.vue'
import StackControl from './StackControl.vue'
import ShowLabelsControl from './ShowLabelsControl.vue'
import SmoothControl from './SmoothControl.vue'
import HorizontalControl from './HorizontalControl.vue'
import ThreeDControl from './ThreeDControl.vue'
import ThreeDRotateControl from './ThreeDRotateControl.vue'
import ThreeDVisualMapControl from './ThreeDVisualMapControl.vue'
import VisualMapControl from './VisualMapControl.vue'
import SwapControl from './SwapControl.vue'

describe('settings controls', () => {
  beforeEach(() => {
    dpState.data = [{ xAxis: 'a', yAxis: 'b' }]
    dpState.axes = undefined
    dpState.isValueMode = false
    dpState.targetString = 'x/y'
  })

  it('SortControl toggles enabled and order', async () => {
    const model = ref<Sort | undefined>(undefined)
    const w = mount(SortControl, {
      props: {
        modelValue: model.value,
        'onUpdate:modelValue': (v: Sort) => {
          model.value = v
          w.setProps({ modelValue: v })
        },
      },
    })
    await w.get('[data-testid="sorting-switch"]').trigger('click')
    expect(model.value).toEqual({ enabled: true, order: 'asc' })
    await w.get('[data-value="desc"]').trigger('click')
    expect(model.value).toEqual({ enabled: true, order: 'desc' })
  })

  it('ScaleControl emits and ignores when disabled', async () => {
    const emitVals: ScaleType[] = []
    const w = mount(ScaleControl, {
      props: {
        modelValue: undefined,
        'onUpdate:modelValue': (v: ScaleType) => emitVals.push(v),
      },
    })
    await w.get('[data-value="log"]').trigger('click')
    expect(emitVals).toEqual(['log'])

    const disabled = mount(ScaleControl, {
      props: {
        modelValue: 'linear',
        disabled: true,
        'onUpdate:modelValue': (v: ScaleType) => emitVals.push(v),
      },
    })
    const before = emitVals.length
    // call onUpdate via component emit path with disabled true
    const tabs = disabled.findComponent({ name: 'SelectionTabs' })
    await tabs.vm.$emit('update:modelValue', 'log')
    expect(emitVals.length).toBe(before)
  })

  it.each([
    ['StackControl', StackControl, 'stack-switch'],
    ['ShowLabelsControl', ShowLabelsControl, 'labels-switch'],
    ['SmoothControl', SmoothControl, 'smooth-lines-switch'],
    ['HorizontalControl', HorizontalControl, 'horizontal-bars-switch'],
    ['ThreeDControl', ThreeDControl, 'three-d-switch'],
    ['ThreeDRotateControl', ThreeDRotateControl, 'three-d-rotate-switch'],
    ['ThreeDVisualMapControl', ThreeDVisualMapControl, 'three-d-visualmap-switch'],
    ['VisualMapControl', VisualMapControl, 'visualmap-switch'],
  ] as const)('%s toggles boolean model', async (_name, Comp, testId) => {
    const vals: boolean[] = []
    const w = mount(Comp, {
      props: {
        modelValue: undefined,
        'onUpdate:modelValue': (v: boolean) => vals.push(v),
      },
    })
    await w.get(`[data-testid="${testId}"]`).trigger('click')
    expect(vals).toEqual([true])
  })

  it('SwapControl selects arrangement option', async () => {
    const vals: string[] = []
    const w = mount(SwapControl, {
      props: {
        modelValue: undefined,
        'onUpdate:modelValue': (v: string) => vals.push(v),
      },
    })
    expect(w.text()).toContain('Swap axis')
    await w.get('[data-testid="swap-selector"]').trigger('click')
    expect(vals).toEqual(['y/x'])

    const vm = w.findComponent({ name: 'Selector' })
    vm.vm.$emit('select', 99)
    expect(vals).toEqual(['y/x'])
  })

  it('SwapControl uses modelValue index when set', () => {
    const w = mount(SwapControl, {
      props: { modelValue: 'x/y/z' },
    })
    expect(w.get('[data-testid="swap-selector"]').text()).toBe('2')
  })

  it('SwapControl falls back to index 0 for unknown target', () => {
    const w = mount(SwapControl, {
      props: { modelValue: 'not-an-option' },
    })
    expect(w.get('[data-testid="swap-selector"]').text()).toBe('0')
  })
})
