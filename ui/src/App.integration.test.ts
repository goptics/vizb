import { describe, it, expect, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'

vi.mock('./views/Dashboard.vue', () => ({
  default: defineComponent({
    name: 'DashboardStub',
    setup() {
      return () => h('div', { 'data-testid': 'dashboard-stub' }, 'dashboard')
    },
  }),
}))

import App from './App.vue'

describe('App.vue', () => {
  it('renders the Dashboard root view', () => {
    const wrapper = mount(App)
    expect(wrapper.find('[data-testid="dashboard-stub"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('dashboard')
  })
})
