import { describe, it, expect, vi } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import type { DescriptiveStats } from '@/types'
import { allDescriptiveColumnKeys } from '@/lib/descriptiveColumns'

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
        modelValue: { type: null },
        open: Boolean,
        searchTerm: String,
        filterFunction: Function,
        multiple: Boolean,
        resetSearchTermOnSelect: Boolean,
        class: null,
      },
      emits: ['update:modelValue', 'update:open', 'update:searchTerm'],
      setup(props, { emit, slots, expose }) {
        expose({
          setModel(v: unknown) {
            emit('update:modelValue', v)
          },
          setSearch(v: string) {
            emit('update:searchTerm', v)
          },
          runFilter(keys: string[], q: string) {
            return (props.filterFunction as (k: string[], q: string) => string[])(keys, q)
          },
        })
        return () =>
          h(
            'div',
            { 'data-testid': 'col-combobox', class: props.class as string },
            slots.default?.()
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
      setup() {
        return () => h('input', { 'data-testid': 'col-search' })
      },
    }),
    ComboboxItem: defineComponent({
      name: 'ComboboxItem',
      props: ['value', 'textValue', 'disabled'],
      setup(p, { slots }) {
        return () =>
          h(
            'div',
            {
              'data-testid': `col-item-${p.value}`,
              'data-disabled': String(!!p.disabled),
            },
            slots.default?.()
          )
      },
    }),
    ComboboxItemIndicator: passthrough('ComboboxItemIndicator'),
    ComboboxLabel: passthrough('ComboboxLabel'),
    ComboboxList: passthrough('ComboboxList'),
    ComboboxSeparator: passthrough('ComboboxSeparator'),
    ComboboxTrigger: defineComponent({
      name: 'ComboboxTrigger',
      setup(_, { slots }) {
        return () => h('button', { 'data-testid': 'col-trigger' }, slots.default?.())
      },
    }),
  }
})

vi.mock('../lib/descriptiveColumns', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../lib/descriptiveColumns')>()
  return {
    ...actual,
    isDescriptiveColumnKey: (key: string) => key === 'ghost' || actual.isDescriptiveColumnKey(key),
  }
})
import DescriptiveColumnSelect from './DescriptiveColumnSelect.vue'

describe('DescriptiveColumnSelect', () => {
  it('renders groups and trigger summary', () => {
    const model = ref<(keyof DescriptiveStats)[]>(['count', 'mean'])
    const w = mount(DescriptiveColumnSelect, {
      props: {
        modelValue: model.value,
        defaultKeys: ['count', 'mean'],
        'onUpdate:modelValue': (v: (keyof DescriptiveStats)[]) => {
          model.value = v
          w.setProps({ modelValue: v })
        },
      },
    })
    expect(w.get('[data-testid="col-trigger"]').text()).toMatch(/Count|Mean|columns/)
    expect(w.text()).toContain('Counts')
    expect(w.text()).toContain('Select all')
    expect(w.text()).toContain('Reset defaults')
  })

  it('select all and reset defaults', async () => {
    const model = ref<(keyof DescriptiveStats)[]>(['count'])
    const w = mount(DescriptiveColumnSelect, {
      props: {
        modelValue: model.value,
        defaultKeys: ['count', 'mean'],
        'onUpdate:modelValue': (v: (keyof DescriptiveStats)[]) => {
          model.value = v
          w.setProps({ modelValue: v })
        },
      },
    })
    const buttons = w.findAll('button')
    const selectAll = buttons.find((b) => b.text() === 'Select all')!
    const reset = buttons.find((b) => b.text() === 'Reset defaults')!
    await selectAll.trigger('click')
    expect(model.value.length).toBe(allDescriptiveColumnKeys().length)
    expect(w.get('[data-testid="col-trigger"]').text()).toBe('All columns')
    await reset.trigger('click')
    expect(model.value).toEqual(['count', 'mean'])
  })

  it('disables last selected item', async () => {
    const model = ref<(keyof DescriptiveStats)[]>(['count'])
    const w = mount(DescriptiveColumnSelect, {
      props: {
        modelValue: model.value,
        defaultKeys: ['count'],
        'onUpdate:modelValue': (v: (keyof DescriptiveStats)[]) => {
          model.value = v
          w.setProps({ modelValue: v })
        },
      },
    })
    expect(w.get('[data-testid="col-item-count"]').attributes('data-disabled')).toBe('true')
    expect(w.get('[data-testid="col-item-mean"]').attributes('data-disabled')).toBe('false')
  })

  it('filterColumns covers empty query, match, and non-keys', async () => {
    const model = ref<(keyof DescriptiveStats)[]>(['count', 'mean'])
    const w = mount(DescriptiveColumnSelect, {
      props: {
        modelValue: model.value,
        defaultKeys: ['count'],
        'onUpdate:modelValue': (v: (keyof DescriptiveStats)[]) => {
          model.value = v
        },
      },
    })
    const cb = w.findComponent({ name: 'Combobox' })
    const filter = cb.props('filterFunction') as (keys: string[], q: string) => string[]
    const keys = ['count', 'mean', 'not-a-key']
    expect(filter(keys, '')).toEqual(keys)
    expect(filter(keys, 'mean')).toEqual(['mean'])
    expect(filter(keys, 'counts')).toContain('count')
    expect(filter(keys, 'zzz')).toEqual([])
    // force unknown key path inside filter (isDescriptiveColumnKey true but find fails is hard;
    // hit via bogus key already returns false early)
    expect(filter(['nope'], 'nope')).toEqual([])
    expect(filter(['ghost'], 'ghost')).toEqual([])
    await cb.vm.$emit('update:modelValue', ['mean', 'stdDev'])
    await nextTick()
    expect(model.value).toEqual(['mean', 'stdDev'])
    // open/search term v-models
    await cb.vm.$emit('update:open', true)
    await cb.vm.$emit('update:searchTerm', 'mean')
    await nextTick()
    expect(w.find('[data-testid="col-combobox"]').exists()).toBe(true)
  })
})
