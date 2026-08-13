import { describe, it, expect, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { Sigma } from 'lucide-vue-next'
import AccentLink from './AccentLink.vue'
import AppFooter from './AppFooter.vue'
import Badge from './Badge.vue'
import BadgeButton from './BadgeButton.vue'
import IconButton from './IconButton.vue'
import LoadingSkeleton from './LoadingSkeleton.vue'
import SettingHeader from './SettingHeader.vue'
import SettingsToggle from './SettingsToggle.vue'
import SelectionTabs from './SelectionTabs.vue'
import HistoryPopover from './HistoryPopover.vue'
import CpuBadge from './CpuBadge.vue'
import OsBadge from './OsBadge.vue'
import TimestampBadge from './TimestampBadge.vue'
import DatasetHeader from './DatasetHeader.vue'
import ChartSettingsPopover from './ChartSettingsPopover.vue'
import type { Dataset, HistoryEntry } from '@/types'

vi.mock('./ui', () => {
  const passthrough = (name: string) =>
    defineComponent({
      name,
      inheritAttrs: true,
      setup(_, { slots, attrs }) {
        return () => h('div', { 'data-stub': name, ...attrs }, slots.default?.())
      },
    })
  return {
    Label: defineComponent({
      name: 'Label',
      props: ['for'],
      setup(props, { slots }) {
        return () => h('label', { for: props.for }, slots.default?.())
      },
    }),
    Switch: defineComponent({
      name: 'Switch',
      props: ['id', 'checked'],
      emits: ['update:checked'],
      setup(props, { emit }) {
        return () =>
          h('button', {
            id: props.id,
            'data-checked': String(!!props.checked),
            onClick: () => emit('update:checked', !props.checked),
          })
      },
    }),
    Tabs: defineComponent({
      name: 'Tabs',
      props: ['modelValue'],
      emits: ['update:modelValue'],
      setup(props, { emit, slots }) {
        return () =>
          h(
            'div',
            {
              'data-tabs': String(props.modelValue),
              onClick: () => emit('update:modelValue', 'next'),
            },
            slots.default?.()
          )
      },
    }),
    TabsList: passthrough('TabsList'),
    TabsTrigger: defineComponent({
      name: 'TabsTrigger',
      props: ['value', 'disabled'],
      setup(props, { slots }) {
        return () =>
          h(
            'button',
            { 'data-value': String(props.value), disabled: props.disabled },
            slots.default?.()
          )
      },
    }),
    Popover: defineComponent({
      name: 'Popover',
      props: ['open'],
      emits: ['update:open'],
      setup(_, { slots }) {
        return () => h('div', { 'data-stub': 'Popover' }, slots.default?.())
      },
    }),
    PopoverTrigger: defineComponent({
      name: 'PopoverTrigger',
      setup(_, { slots }) {
        return () => h('div', { 'data-stub': 'PopoverTrigger' }, slots.default?.())
      },
    }),
    PopoverContent: defineComponent({
      name: 'PopoverContent',
      setup(_, { slots }) {
        return () => h('div', { 'data-stub': 'PopoverContent' }, slots.default?.())
      },
    }),
  }
})

vi.mock('./ui/Popover.vue', () => ({
  default: defineComponent({
    name: 'Popover',
    setup(_, { slots }) {
      return () => h('div', { 'data-stub': 'Popover' }, slots.default?.())
    },
  }),
}))
vi.mock('./ui/PopoverContent.vue', () => ({
  default: defineComponent({
    name: 'PopoverContent',
    setup(_, { slots, attrs }) {
      return () =>
        h('div', { 'data-stub': 'PopoverContent', class: attrs.class }, slots.default?.())
    },
  }),
}))
vi.mock('./ui/PopoverTrigger.vue', () => ({
  default: defineComponent({
    name: 'PopoverTrigger',
    setup(_, { slots }) {
      return () => h('div', { 'data-stub': 'PopoverTrigger' }, slots.default?.())
    },
  }),
}))

vi.mock('./SettingsPanel.vue', () => ({
  default: defineComponent({
    name: 'SettingsPanel',
    setup: () => () => h('div', { 'data-testid': 'settings-panel' }),
  }),
}))

vi.mock('./Selector.vue', () => ({
  default: defineComponent({
    name: 'Selector',
    props: ['items', 'activeId'],
    emits: ['select'],
    setup(props, { emit }) {
      return () =>
        h(
          'button',
          {
            'data-testid': 'selector-stub',
            onClick: () => emit('select', 1),
          },
          String(props.activeId)
        )
    },
  }),
}))

const history: HistoryEntry[] = [
  {
    tag: 'v1',
    timestamp: '2024-01-02T00:00:00.000Z',
    meta: { cpu: { name: 'Old', cores: 4 }, os: 'linux' },
  },
  {
    tag: 'v2',
    timestamp: '2024-06-01T12:00:00.000Z',
    meta: { cpu: { name: 'New', cores: 8 }, os: 'darwin' },
  },
]

describe('AccentLink', () => {
  it('renders external link slot', () => {
    const w = mount(AccentLink, {
      props: { href: 'https://example.com' },
      slots: { default: 'Go' },
    })
    const a = w.get('a')
    expect(a.attributes('href')).toBe('https://example.com')
    expect(a.attributes('target')).toBe('_blank')
    expect(a.text()).toBe('Go')
  })
})

describe('AppFooter', () => {
  it('renders version and year', () => {
    const w = mount(AppFooter, { props: { version: 'v9.9.9' } })
    expect(w.text()).toContain('Vizb')
    expect(w.text()).toContain('Goptics')
    expect(w.text()).toContain('v9.9.9')
    expect(w.text()).toContain(String(new Date().getFullYear()))
  })
})

describe('Badge / BadgeButton / IconButton', () => {
  it('Badge renders label, value, optional icon', () => {
    const withIcon = mount(Badge, {
      props: { label: 'X', value: '2', icon: Sigma },
    })
    expect(withIcon.text()).toContain('X:')
    expect(withIcon.text()).toContain('2')
    expect(withIcon.find('svg').exists()).toBe(true)

    const plain = mount(Badge, { props: { label: 'Y', value: '3' } })
    expect(plain.find('svg').exists()).toBe(false)
  })

  it('BadgeButton active and inactive + optional icon', async () => {
    const active = mount(BadgeButton, {
      props: { label: 'Stats', active: true, title: 't', icon: Sigma },
    })
    expect(active.attributes('aria-pressed')).toBe('true')
    expect(active.find('svg').exists()).toBe(true)

    const inactive = mount(BadgeButton, { props: { label: 'Stats', active: false } })
    expect(inactive.attributes('aria-pressed')).toBe('false')
    expect(inactive.find('svg').exists()).toBe(false)
  })

  it('IconButton is button without href and anchor with href', () => {
    const btn = mount(IconButton, { slots: { default: 'B' } })
    expect(btn.element.tagName.toLowerCase()).toBe('button')
    expect(btn.attributes('type')).toBe('button')

    const link = mount(IconButton, {
      props: { href: 'https://pkg.example' },
      slots: { default: 'L' },
    })
    expect(link.element.tagName.toLowerCase()).toBe('a')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toContain('noopener')
  })
})

describe('LoadingSkeleton', () => {
  it('full page vs contentOnly', () => {
    const full = mount(LoadingSkeleton)
    expect(full.find('header').exists()).toBe(true)
    expect(full.classes().join(' ')).toContain('min-h-screen')

    const content = mount(LoadingSkeleton, { props: { contentOnly: true } })
    expect(content.find('header').exists()).toBe(false)
  })
})

describe('SettingHeader / SettingsToggle / SelectionTabs', () => {
  it('SettingHeader optional description', () => {
    const withDesc = mount(SettingHeader, {
      props: { label: 'Sort', description: 'desc', id: 's1' },
    })
    expect(withDesc.text()).toContain('Sort')
    expect(withDesc.text()).toContain('desc')

    const bare = mount(SettingHeader, { props: { label: 'Only' } })
    expect(bare.text()).toBe('Only')
  })

  it('SettingsToggle emits checked updates', async () => {
    const onUpdateChecked = vi.fn()
    const w = mount(SettingsToggle, {
      props: {
        id: 't1',
        label: 'On',
        description: 'd',
        checked: false,
        'onUpdate:checked': onUpdateChecked,
      },
    })
    await w.get('button').trigger('click')
    expect(onUpdateChecked).toHaveBeenCalledWith(true)
  })

  it('SelectionTabs emits and respects disabled', async () => {
    const opts = [
      { value: 'a', label: 'A', icon: Sigma },
      { value: 'b', label: 'B' },
    ]
    const onUpdateModelValue = vi.fn()
    const w = mount(SelectionTabs, {
      props: { modelValue: 'a', options: opts, 'onUpdate:modelValue': onUpdateModelValue },
    })
    expect(w.text()).toContain('A')
    expect(w.text()).toContain('B')
    await w.get('[data-tabs]').trigger('click')
    expect(onUpdateModelValue).toHaveBeenCalledWith('next')

    const onDisabledUpdate = vi.fn()
    const disabled = mount(SelectionTabs, {
      props: {
        modelValue: 'a',
        options: opts,
        disabled: true,
        'onUpdate:modelValue': onDisabledUpdate,
      },
    })
    await disabled.get('[data-tabs]').trigger('click')
    expect(onDisabledUpdate).not.toHaveBeenCalled()
  })
})

describe('History + meta badges', () => {
  it('HistoryPopover renders entries when hasHistory', () => {
    const w = mount(HistoryPopover, {
      props: {
        icon: Sigma,
        label: 'CPU',
        value: 'x',
        historyTitle: 'CPU History',
        entries: history,
        hasHistory: true,
        contentWidth: 'w-80',
      },
      slots: {
        entry: `<span class="entry">{{ entry.tag }}</span>`,
      },
    })
    expect(w.text()).toContain('CPU History')
    expect(w.text()).toContain('v1')
  })

  it('HistoryPopover omits content without history', () => {
    const w = mount(HistoryPopover, {
      props: {
        icon: Sigma,
        label: 'CPU',
        value: 'x',
        historyTitle: 'H',
        entries: [],
        hasHistory: false,
      },
    })
    expect(w.find('[data-stub="PopoverContent"]').exists()).toBe(false)
  })

  it('CpuBadge / OsBadge / TimestampBadge render with history', () => {
    const cpu = mount(CpuBadge, {
      props: {
        cpu: { name: 'Ryzen', cores: 16 },
        history: [
          ...history,
          { tag: 'cores-only', timestamp: '2023-01-01T00:00:00.000Z', meta: { cpu: { cores: 2 } } },
          { tag: 'name-only', timestamp: '2023-02-01T00:00:00.000Z', meta: { cpu: { name: 'X' } } },
        ],
      },
    })
    expect(cpu.text()).toContain('CPU')
    expect(cpu.text()).toContain('Ryzen')

    const os = mount(OsBadge, { props: { os: 'linux', history } })
    expect(os.text()).toContain('OS')
    expect(os.text()).toContain('linux')

    const ts = mount(TimestampBadge, {
      props: { timestamp: '2024-06-01T12:30:00.000Z', history },
    })
    expect(ts.text()).toContain('Updated')

    const badTs = mount(TimestampBadge, {
      props: { timestamp: 'not-a-date' },
    })
    expect(badTs.text()).toContain('not-a-date')

    // empty timestamp string branch in formattedDate
    const emptyTs = mount(TimestampBadge, { props: { timestamp: '' } })
    expect(emptyTs.find('[data-stub="Popover"]').exists()).toBe(false)
  })

  it('badges hide without primary props', () => {
    expect(mount(CpuBadge, { props: {} }).find('[data-stub="Popover"]').exists()).toBe(false)
    expect(mount(OsBadge, { props: {} }).find('[data-stub="Popover"]').exists()).toBe(false)
    const bareTs = mount(TimestampBadge, { props: {} })
    expect(bareTs.find('[data-stub="Popover"]').exists()).toBe(false)
    const setup = (bareTs.vm as unknown as { $: { setupState: { formattedDate: string } } }).$
      .setupState
    expect(setup.formattedDate).toBe('')
  })
})

describe('DatasetHeader', () => {
  const baseDataset: Dataset = {
    name: 'Bench',
    description: 'desc',
    timestamp: '2024-01-01T00:00:00.000Z',
    meta: { cpu: { name: 'M1', cores: 8 }, os: 'darwin' },
    history,
    settings: [
      {
        type: 'bar',
        sort: { enabled: false, order: 'asc' },
        scale: 'linear',
        stack: false,
        showLabels: false,
        swap: 'x/y',
      },
    ],
    data: [],
  }

  it('single dataset title path', () => {
    const w = mount(DatasetHeader, {
      props: {
        dataset: baseDataset,
        datasets: [{ name: 'Bench' }],
        activeDatasetId: 0,
        resultGroups: [{ name: 'g0' }],
        activeGroupId: 0,
      },
    })
    expect(w.get('h1').text()).toBe('Bench')
    expect(w.text()).toContain('desc')
    expect(w.findComponent({ name: 'CpuBadge' }).exists()).toBe(true)
    expect(w.findComponent({ name: 'OsBadge' }).exists()).toBe(true)
  })

  it('multi dataset/group selectors emit', async () => {
    const onSelectDataset = vi.fn()
    const onSelectGroup = vi.fn()
    const w = mount(DatasetHeader, {
      props: {
        dataset: { ...baseDataset, meta: {}, description: undefined, timestamp: undefined },
        datasets: [{ name: 'A' }, { name: 'B' }],
        activeDatasetId: 0,
        resultGroups: [{ name: 'g0' }, { name: 'g1' }],
        activeGroupId: 0,
        onSelectDataset,
        onSelectGroup,
      },
    })
    const selectors = w.findAllComponents({ name: 'Selector' })
    expect(selectors.length).toBe(2)
    await selectors[0]!.trigger('click')
    expect(onSelectDataset).toHaveBeenCalledWith(1)
    await selectors[1]!.trigger('click')
    expect(onSelectGroup).toHaveBeenCalledWith(1)
  })

  it('falls back title when datasets empty', () => {
    const w = mount(DatasetHeader, {
      props: {
        dataset: { ...baseDataset, meta: undefined },
        datasets: [],
        activeDatasetId: 0,
        resultGroups: [],
        activeGroupId: 0,
      },
    })
    expect(w.get('h1').text()).toBe('Datasets')
  })
})

describe('ChartSettingsPopover', () => {
  it('mounts settings trigger and panel', async () => {
    const w = mount(ChartSettingsPopover)
    expect(w.find('[aria-label="Open chart settings"]').exists()).toBe(true)
    expect(w.find('[data-testid="settings-panel"]').exists()).toBe(true)
    const pop = w.findComponent({ name: 'Popover' })
    expect(pop.exists()).toBe(true)
    await pop.vm.$emit('update:open', true)
    await pop.vm.$emit('update:open', false)
  })
})
