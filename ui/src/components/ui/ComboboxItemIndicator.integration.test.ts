import { describe, it, expect, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'

// Replace only the radix indicator with a slot-forwarding stub so our wrapper's
// <slot /> path is instrumented. Other radix primitives stay real.
vi.mock('radix-vue', async (importOriginal) => {
  const actual = await importOriginal<typeof import('radix-vue')>()
  const Stub = defineComponent({
    name: 'RadixComboboxItemIndicatorStub',
    setup(_, { slots, attrs }) {
      return () => h('span', { ...attrs, 'data-testid': 'radix-ind' }, slots.default?.())
    },
  })
  return {
    ...actual,
    ComboboxItemIndicator: Stub,
  }
})

import ComboboxItemIndicator from './ComboboxItemIndicator.vue'

describe('ComboboxItemIndicator', () => {
  it('forwards class and renders default slot via radix child', () => {
    const w = mount(ComboboxItemIndicator, {
      props: { class: 'ind-class' },
      slots: { default: 'MARK' },
    })
    expect(w.get('[data-testid="radix-ind"]').text()).toBe('MARK')
    expect(w.get('[data-testid="radix-ind"]').classes()).toContain('ind-class')
  })

  it('renders without optional class', () => {
    const w = mount(ComboboxItemIndicator, {
      slots: { default: 'X' },
    })
    expect(w.text()).toContain('X')
  })
})
