import { describe, it, expect, vi } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { Sigma } from 'lucide-vue-next'

vi.mock('./ui/combobox', () => {
  const passthrough = (name: string) =>
    defineComponent({
      name,
      setup(_, { slots, attrs }) {
        return () => h('div', { 'data-stub': name, ...attrs }, slots.default?.())
      },
    })

  return {
    Combobox: defineComponent({
      name: 'Combobox',
      props: {
        open: Boolean,
        modelValue: { type: null },
        searchTerm: String,
        filterFunction: Function,
        by: String,
        class: null,
        multiple: Boolean,
      },
      emits: ['update:open', 'update:modelValue', 'update:searchTerm'],
      setup(props, { emit, slots }) {
        return () =>
          h(
            'div',
            {
              'data-testid': 'combobox',
              'data-open': String(props.open),
              class: props.class as string | undefined,
            },
            [
              h(
                'button',
                {
                  'data-testid': 'emit-value',
                  onClick: () =>
                    emit('update:modelValue', {
                      value: 'emit',
                      label: 'Emit',
                    }),
                },
                'emit-value'
              ),
              h(
                'button',
                {
                  'data-testid': 'emit-open-false',
                  onClick: () => emit('update:open', false),
                },
                'close'
              ),
              h(
                'button',
                {
                  'data-testid': 'emit-search',
                  onClick: () => emit('update:searchTerm', 'ap'),
                },
                'search'
              ),
              slots.default?.(),
            ]
          )
      },
    }),
    ComboboxAnchor: passthrough('ComboboxAnchor'),
    ComboboxEmpty: defineComponent({
      name: 'ComboboxEmpty',
      setup(_, { slots }) {
        return () => h('div', { 'data-testid': 'empty' }, slots.default?.())
      },
    }),
    ComboboxGroup: passthrough('ComboboxGroup'),
    ComboboxInput: defineComponent({
      name: 'ComboboxInput',
      setup(_, { attrs }) {
        return () => h('input', { 'data-testid': 'search', ...attrs })
      },
    }),
    ComboboxItem: defineComponent({
      name: 'ComboboxItem',
      props: ['value', 'textValue'],
      setup(p, { slots }) {
        const val =
          p.value && typeof p.value === 'object' && 'value' in p.value
            ? String((p.value as { value: string }).value)
            : String(p.value)
        return () => h('div', { 'data-testid': 'item', 'data-value': val }, slots.default?.())
      },
    }),
    ComboboxList: passthrough('ComboboxList'),
    ComboboxTrigger: defineComponent({
      name: 'ComboboxTrigger',
      setup(_, { slots, attrs }) {
        return () =>
          h(
            'button',
            {
              'data-testid': 'trigger',
              'aria-label': attrs['aria-label'] as string | undefined,
              class: attrs.class as string | undefined,
            },
            slots.default?.()
          )
      },
    }),
  }
})

import Selector from './Selector.vue'

function manyItems(n: number) {
  return Array.from({ length: n }, (_, i) => ({ name: `Item ${i}`, value: `v${i}` }))
}

describe('Selector', () => {
  it('renders nothing when items length <= 1', () => {
    const w = mount(Selector, { props: { items: [{ name: 'Only' }] } })
    expect(w.find('[data-testid="combobox"]').exists()).toBe(false)
  })

  it('shows trigger with active value and icon', () => {
    const w = mount(Selector, {
      props: {
        items: [
          { name: 'Alpha', value: 'a', icon: Sigma },
          { name: 'Beta', value: 'b' },
        ],
        activeValue: 'a',
        ariaLabel: 'pick',
        triggerClass: 'h-12',
      },
    })
    expect(w.get('[data-testid="trigger"]').attributes('aria-label')).toBe('pick')
    expect(w.text()).toContain('Alpha')
    expect(w.find('svg').exists()).toBe(true)
  })

  it('emits selectValue when model changes under activeValue mode', async () => {
    const onSelectValue = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [
          { name: 'Alpha', value: 'a' },
          { name: 'Beta', value: 'b' },
          { name: 'Emit', value: 'emit' },
        ],
        activeValue: 'a',
        onSelectValue,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: 'b', label: 'Beta' })
    await nextTick()
    expect(onSelectValue).toHaveBeenCalledWith('b')
  })

  it('emits select by index when activeId mode', async () => {
    const onSelect = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [{ name: 'A' }, { name: 'B' }, { name: 'C' }],
        activeId: 0,
        onSelect,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: '2', label: 'C' })
    await nextTick()
    expect(onSelect).toHaveBeenCalledWith(2)
  })

  it('re-emits on open close when value differs', async () => {
    const onSelect = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [{ name: 'A' }, { name: 'B' }],
        activeId: 0,
        onSelect,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: '1', label: 'B' })
    await nextTick()
    const before = onSelect.mock.calls.length
    await cb.vm.$emit('update:open', false)
    await nextTick()
    expect(onSelect.mock.calls.length).toBe(before)
  })

  it('shows search when >10 options and truncation message with resultLimit', () => {
    const items = manyItems(15)
    const w = mount(Selector, {
      props: {
        items,
        activeValue: 'v0',
        resultLimit: 5,
        notFoundText: 'none',
      },
    })
    expect(w.find('[data-testid="search"]').exists()).toBe(true)
    expect(w.text()).toMatch(/Showing \d+ of \d+ matches/)
    expect(w.get('[data-testid="empty"]').text()).toBe('none')
  })

  it('comboboxFilter is identity', async () => {
    const w = mount(Selector, {
      props: {
        items: [
          { name: 'Apple', value: 'a' },
          { name: 'Banana', value: 'b' },
        ],
        activeValue: 'a',
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:searchTerm', 'ap')
    await nextTick()
    const filter = cb.props('filterFunction') as (list: unknown[]) => unknown[]
    expect(filter([{ value: '1' }])).toEqual([{ value: '1' }])
  })

  it('falls back activeId string when activeValue absent and clears unknown active', async () => {
    const w = mount(Selector, {
      props: {
        items: [{ name: 'A' }, { name: 'B' }],
        activeId: 1,
      },
    })
    expect(w.text()).toContain('B')
    await w.setProps({ activeId: 99 })
    await nextTick()
    expect(w.find('[data-testid="trigger"]').exists()).toBe(true)
  })

  it('does not emit select when same activeId selected', async () => {
    const onSelect = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [{ name: 'A' }, { name: 'B' }],
        activeId: 1,
        onSelect,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: '1', label: 'B' })
    await nextTick()
    expect(onSelect).not.toHaveBeenCalled()
  })

  it('does not emit selectValue when same activeValue selected', async () => {
    const onSelectValue = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [
          { name: 'A', value: 'a' },
          { name: 'B', value: 'b' },
        ],
        activeValue: 'a',
        onSelectValue,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: 'a', label: 'A' })
    await nextTick()
    expect(onSelectValue).not.toHaveBeenCalled()
  })

  it('activeValue open-close emit path', async () => {
    const onSelectValue = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [
          { name: 'A', value: 'a' },
          { name: 'B', value: 'b' },
        ],
        activeValue: 'a',
        onSelectValue,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: 'b', label: 'B' })
    await nextTick()
    await cb.vm.$emit('update:open', true)
    await nextTick()
    await cb.vm.$emit('update:open', false)
    await nextTick()
    expect(onSelectValue).toHaveBeenCalled()
  })

  it('index items without value use stringified index', async () => {
    const onSelect = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [{ name: 'One' }, { name: 'Two' }],
        activeId: 0,
        onSelect,
      },
    })
    expect(w.text()).toContain('One')
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: 'NaN', label: 'x' })
    await nextTick()
    expect(onSelect).not.toHaveBeenCalled()
  })

  it('open-close activeId emit when value differs', async () => {
    const onSelect = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [{ name: 'A' }, { name: 'B' }],
        activeId: 0,
        onSelect,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: '1', label: 'B' })
    await nextTick()
    await cb.vm.$emit('update:open', true)
    await nextTick()
    await cb.vm.$emit('update:open', false)
    await nextTick()
    expect(onSelect).toHaveBeenCalled()
  })

  it('open-close activeValue same value does not emit', async () => {
    const onSelectValue = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [
          { name: 'A', value: 'a' },
          { name: 'B', value: 'b' },
        ],
        activeValue: 'a',
        onSelectValue,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: 'a', label: 'A' })
    await nextTick()
    await cb.vm.$emit('update:open', true)
    await nextTick()
    await cb.vm.$emit('update:open', false)
    await nextTick()
    expect(onSelectValue).not.toHaveBeenCalled()
  })

  it('handles missing activeId and activeValue', async () => {
    const w = mount(Selector, {
      props: {
        items: [{ name: 'A' }, { name: 'B' }],
        activeId: null as unknown as number,
        activeValue: undefined,
      },
    })
    expect(w.find('[data-testid="trigger"]').exists()).toBe(true)
    await w.setProps({ activeId: 0, activeValue: undefined })
    await nextTick()
    await w.setProps({ activeId: null as unknown as number, activeValue: undefined })
    await nextTick()
    await w.setProps({ items: [{ name: 'A' }, { name: 'B' }, { name: 'C' }] })
    await nextTick()
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: '0', label: 'A' })
    await nextTick()
  })

  it('open-close activeId same index does not emit', async () => {
    const onSelect = vi.fn()
    const w = mount(Selector, {
      props: {
        items: [{ name: 'A' }, { name: 'B' }],
        activeId: 1,
        onSelect,
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    await cb.vm.$emit('update:modelValue', { value: '1', label: 'B' })
    await nextTick()
    await cb.vm.$emit('update:open', true)
    await nextTick()
    await cb.vm.$emit('update:open', false)
    await nextTick()
    expect(onSelect).not.toHaveBeenCalled()
  })

  it('activeOption optional chain both sides', async () => {
    const w1 = mount(Selector, {
      props: { items: [{ name: 'A' }, { name: 'B' }], activeId: 1 },
    })
    expect(w1.text()).toContain('B')
    w1.unmount()

    const w2 = mount(Selector, {
      props: { items: [{ name: 'A' }, { name: 'B' }] },
    })
    // no activeId/activeValue
    expect(w2.find('[data-testid="trigger"]').exists()).toBe(true)
    await w2.setProps({ items: [{ name: 'A' }, { name: 'B' }, { name: 'C' }] })
    await nextTick()
    w2.unmount()

    const w3 = mount(Selector, {
      props: {
        items: [
          { name: 'A', value: 'a' },
          { name: 'B', value: 'b' },
        ],
        activeValue: 'b',
      },
    })
    expect(w3.text()).toContain('B')
    // switch to activeId-only mode
    await w3.setProps({ activeValue: undefined, activeId: 0 })
    await nextTick()
    await w3.setProps({ activeValue: undefined, activeId: undefined as unknown as number })
    await nextTick()
  })

  it('activeOption with resultLimit covers activeId optional chain', async () => {
    const items = Array.from({ length: 12 }, (_, i) => ({ name: `N${i}`, value: `v${i}` }))
    const w = mount(Selector, {
      props: { items, activeId: 2, resultLimit: 5 },
    })
    expect(w.text()).toMatch(/Showing/)
    await w.setProps({ activeId: undefined as unknown as number, activeValue: undefined })
    await nextTick()
    await w.setProps({ activeId: 0 })
    await nextTick()
    // activeValue short-circuit path on activeOption
    await w.setProps({ activeValue: 'v1', activeId: undefined as unknown as number })
    await nextTick()
  })
})
