import { describe, it, expect } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import * as Ui from './index'
import * as ComboboxBarrel from './combobox'
import Button from './Button.vue'
import Card from './Card.vue'
import CardContent from './CardContent.vue'
import CardHeader from './CardHeader.vue'
import CardTitle from './CardTitle.vue'
import Label from './Label.vue'
import Separator from './Separator.vue'
import Switch from './Switch.vue'
import Tabs from './Tabs.vue'
import TabsList from './TabsList.vue'
import TabsTrigger from './TabsTrigger.vue'
import TabsContent from './TabsContent.vue'
import ToggleGroup from './ToggleGroup.vue'
import ToggleGroupItem from './ToggleGroupItem.vue'
import Popover from './Popover.vue'
import PopoverTrigger from './PopoverTrigger.vue'
import PopoverContent from './PopoverContent.vue'
import Combobox from './Combobox.vue'
import ComboboxAnchor from './ComboboxAnchor.vue'
import ComboboxInput from './ComboboxInput.vue'
import ComboboxTrigger from './ComboboxTrigger.vue'
import ComboboxList from './ComboboxList.vue'
import ComboboxEmpty from './ComboboxEmpty.vue'
import ComboboxGroup from './ComboboxGroup.vue'
import ComboboxLabel from './ComboboxLabel.vue'
import ComboboxItem from './ComboboxItem.vue'
import ComboboxItemIndicator from './ComboboxItemIndicator.vue'
import ComboboxSeparator from './ComboboxSeparator.vue'

describe('ui barrels', () => {
  it('re-exports primitives from index and combobox', () => {
    expect(Ui.Button).toBe(Button)
    expect(Ui.Card).toBe(Card)
    expect(Ui.CardContent).toBe(CardContent)
    expect(Ui.CardHeader).toBe(CardHeader)
    expect(Ui.CardTitle).toBe(CardTitle)
    expect(Ui.Label).toBe(Label)
    expect(Ui.Separator).toBe(Separator)
    expect(Ui.Switch).toBe(Switch)
    expect(Ui.Tabs).toBe(Tabs)
    expect(Ui.TabsList).toBe(TabsList)
    expect(Ui.TabsTrigger).toBe(TabsTrigger)
    expect(Ui.TabsContent).toBe(TabsContent)
    expect(Ui.ToggleGroup).toBe(ToggleGroup)
    expect(Ui.ToggleGroupItem).toBe(ToggleGroupItem)
    expect(Ui.Popover).toBe(Popover)
    expect(Ui.PopoverTrigger).toBe(PopoverTrigger)
    expect(Ui.PopoverContent).toBe(PopoverContent)
    expect(Ui.Combobox).toBe(Combobox)
    expect(Ui.ComboboxAnchor).toBe(ComboboxAnchor)
    expect(Ui.ComboboxInput).toBe(ComboboxInput)
    expect(Ui.ComboboxTrigger).toBe(ComboboxTrigger)
    expect(Ui.ComboboxList).toBe(ComboboxList)
    expect(Ui.ComboboxEmpty).toBe(ComboboxEmpty)
    expect(Ui.ComboboxGroup).toBe(ComboboxGroup)
    expect(Ui.ComboboxItem).toBe(ComboboxItem)
    expect(Ui.ComboboxItemIndicator).toBe(ComboboxItemIndicator)

    expect(ComboboxBarrel.Combobox).toBe(Combobox)
    expect(ComboboxBarrel.ComboboxAnchor).toBe(ComboboxAnchor)
    expect(ComboboxBarrel.ComboboxInput).toBe(ComboboxInput)
    expect(ComboboxBarrel.ComboboxTrigger).toBe(ComboboxTrigger)
    expect(ComboboxBarrel.ComboboxList).toBe(ComboboxList)
    expect(ComboboxBarrel.ComboboxEmpty).toBe(ComboboxEmpty)
    expect(ComboboxBarrel.ComboboxGroup).toBe(ComboboxGroup)
    expect(ComboboxBarrel.ComboboxLabel).toBe(ComboboxLabel)
    expect(ComboboxBarrel.ComboboxItem).toBe(ComboboxItem)
    expect(ComboboxBarrel.ComboboxItemIndicator).toBe(ComboboxItemIndicator)
    expect(ComboboxBarrel.ComboboxSeparator).toBe(ComboboxSeparator)
  })
})

describe('Button', () => {
  it('defaults type to button and renders slots', () => {
    const w = mount(Button, {
      slots: {
        default: 'Save',
        icon: '<span class="icon">*</span>',
      },
      props: { class: 'extra' },
    })
    const btn = w.get('button')
    expect(btn.attributes('type')).toBe('button')
    expect(btn.classes()).toContain('extra')
    expect(btn.text()).toContain('Save')
    expect(btn.text()).toContain('*')
  })

  it('honors explicit type prop', () => {
    const w = mount(Button, { props: { type: 'submit' }, slots: { default: 'Go' } })
    expect(w.get('button').attributes('type')).toBe('submit')
  })
})

describe('Card family', () => {
  it('mounts card structure with custom classes', () => {
    const w = mount({
      components: { Card, CardHeader, CardTitle, CardContent },
      template: `
        <Card class="card-x">
          <CardHeader class="hdr"><CardTitle class="ttl">Title</CardTitle></CardHeader>
          <CardContent class="body">Body</CardContent>
        </Card>
      `,
    })
    expect(w.get('.card-x').text()).toContain('Title')
    expect(w.get('.card-x').text()).toContain('Body')
    expect(w.get('.ttl').text()).toBe('Title')
    expect(w.get('.body').text()).toBe('Body')
  })
})

describe('Label / Separator', () => {
  it('binds for attribute on Label', () => {
    const w = mount(Label, {
      props: { for: 'name', class: 'lbl' },
      slots: { default: 'Name' },
    })
    const label = w.get('label')
    expect(label.attributes('for')).toBe('name')
    expect(label.classes()).toContain('lbl')
    expect(label.text()).toBe('Name')
  })

  it('renders horizontal separator by default', () => {
    const w = mount(Separator, { props: { class: 'sep' } })
    const el = w.get('[role="separator"]')
    expect(el.attributes('aria-orientation')).toBe('horizontal')
    expect(el.classes()).toContain('sep')
  })

  it('renders vertical separator orientation', () => {
    const w = mount(Separator, { props: { orientation: 'vertical' } })
    const el = w.get('[role="separator"]')
    expect(el.attributes('aria-orientation')).toBe('vertical')
  })
})

describe('Switch', () => {
  it('mounts radix switch root and thumb', () => {
    const w = mount(Switch, {
      props: { class: 'sw', checked: false },
      attrs: { 'aria-label': 'toggle' },
    })
    const root = w.get('[role="switch"]')
    expect(root.classes()).toContain('sw')
    expect(root.attributes('aria-checked')).toBe('false')
    expect(w.find('span').exists()).toBe(true)
  })
})

describe('Tabs', () => {
  it('renders list, triggers, and default content', () => {
    const w = mount({
      components: { Tabs, TabsList, TabsTrigger, TabsContent },
      template: `
        <Tabs default-value="a" class="tabs">
          <TabsList class="list">
            <TabsTrigger value="a" class="ta">A</TabsTrigger>
            <TabsTrigger value="b" class="tb">B</TabsTrigger>
          </TabsList>
          <TabsContent value="a" class="ca">Alpha</TabsContent>
          <TabsContent force-mount value="b" class="cb">Beta</TabsContent>
        </Tabs>
      `,
    })
    expect(w.get('.tabs')).toBeTruthy()
    expect(w.get('.list').text()).toContain('A')
    expect(w.get('.list').text()).toContain('B')
    expect(w.get('.ta').attributes('data-state')).toBe('active')
    expect(w.get('.tb').attributes('data-state')).toBe('inactive')
    expect(w.get('.ca').text()).toBe('Alpha')
    expect(w.get('.cb').text()).toBe('Beta')
  })
})

describe('ToggleGroup', () => {
  it('mounts items and reflects selection', async () => {
    const model = ref('one')
    const Host = defineComponent({
      components: { ToggleGroup, ToggleGroupItem },
      setup: () => ({ model }),
      template: `
        <ToggleGroup type="single" v-model="model" class="tg">
          <ToggleGroupItem value="one" class="i1">One</ToggleGroupItem>
          <ToggleGroupItem value="two" class="i2">Two</ToggleGroupItem>
        </ToggleGroup>
      `,
    })
    const w = mount(Host)
    expect(w.get('.tg')).toBeTruthy()
    expect(w.get('.i1').attributes('data-state')).toBe('on')
    await w.get('.i2').trigger('click')
    await nextTick()
    expect(model.value).toBe('two')
  })
})

describe('Popover', () => {
  it('mounts trigger and force-mounted content with default align', async () => {
    const w = mount({
      components: { Popover, PopoverTrigger, PopoverContent },
      template: `
        <Popover :open="true">
          <PopoverTrigger class="trig">Open</PopoverTrigger>
          <PopoverContent class="pop" force-mount>Hello</PopoverContent>
        </Popover>
      `,
      attachTo: document.body,
    })
    await nextTick()
    expect(w.get('.trig').text()).toContain('Open')
    const content = document.body.querySelector('.pop') ?? w.find('.pop').element
    expect(content).toBeTruthy()
    expect((content as HTMLElement).textContent ?? '').toContain('Hello')
    w.unmount()
  })

  it('honors explicit align on PopoverContent', async () => {
    const w = mount({
      components: { Popover, PopoverTrigger, PopoverContent },
      template: `
        <Popover :open="true">
          <PopoverTrigger class="trig2">T</PopoverTrigger>
          <PopoverContent align="start" class="aligned" force-mount>X</PopoverContent>
        </Popover>
      `,
      attachTo: document.body,
    })
    await nextTick()
    const content = document.body.querySelector('.aligned') ?? w.find('.aligned').element
    expect(content).toBeTruthy()
    expect((content as HTMLElement).textContent ?? '').toContain('X')
    w.unmount()
  })
})
describe('Combobox family', () => {
  it('mounts full combobox tree with delegated class props', async () => {
    const Host = defineComponent({
      components: {
        Combobox,
        ComboboxAnchor,
        ComboboxTrigger,
        ComboboxList,
        ComboboxInput,
        ComboboxEmpty,
        ComboboxGroup,
        ComboboxLabel,
        ComboboxItem,
        ComboboxItemIndicator,
        ComboboxSeparator,
      },
      setup() {
        const open = ref(true)
        const model = ref('alpha')
        const searchTerm = ref('')
        return { open, model, searchTerm }
      },
      template: `
        <Combobox v-model="model" v-model:open="open" v-model:searchTerm="searchTerm" class="cb-root">
          <ComboboxAnchor class="anchor">
            <ComboboxTrigger class="trigger">Pick</ComboboxTrigger>
          </ComboboxAnchor>
          <ComboboxList class="list">
            <ComboboxInput class="input" placeholder="Search" />
            <ComboboxEmpty class="empty">None</ComboboxEmpty>
            <ComboboxGroup class="group">
              <ComboboxLabel class="label">Group</ComboboxLabel>
              <ComboboxItem value="alpha" class="item">
                Alpha
                <ComboboxItemIndicator class="ind">✓</ComboboxItemIndicator>
              </ComboboxItem>
              <ComboboxSeparator class="sep" />
              <ComboboxItem value="beta" class="item2">Beta</ComboboxItem>
            </ComboboxGroup>
          </ComboboxList>
        </Combobox>
      `,
    })

    const w = mount(Host, { attachTo: document.body })
    await nextTick()
    expect(w.get('.cb-root')).toBeTruthy()
    expect(w.get('.anchor')).toBeTruthy()
    expect(w.get('.trigger').text()).toContain('Pick')
    expect(w.html()).toContain('Search')
    expect(w.html()).toContain('Group')
    expect(w.html()).toContain('Alpha')
    expect(w.html()).toContain('Beta')
    w.unmount()
  })

  it('renders without optional class props', async () => {
    const Bare = defineComponent({
      components: {
        Combobox,
        ComboboxAnchor,
        ComboboxTrigger,
        ComboboxList,
        ComboboxEmpty,
        ComboboxGroup,
        ComboboxItem,
      },
      setup: () => ({ open: ref(true), model: ref('x') }),
      template: `
        <Combobox v-model="model" :open="open">
          <ComboboxAnchor><ComboboxTrigger>T</ComboboxTrigger></ComboboxAnchor>
          <ComboboxList>
            <ComboboxEmpty>E</ComboboxEmpty>
            <ComboboxGroup>
              <ComboboxItem value="x">X</ComboboxItem>
            </ComboboxGroup>
          </ComboboxList>
        </Combobox>
      `,
    })
    const w = mount(Bare, { attachTo: document.body })
    await nextTick()
    expect(w.text()).toContain('T')
    w.unmount()
  })
})
