import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import LoadError from './LoadError.vue'

describe('LoadError', () => {
  it('shows message text', () => {
    const w = mount(LoadError, { props: { message: 'network down' } })
    expect(w.text()).toContain('network down')
    expect(w.text()).toContain('Failed to load Data set')
  })

  it('hides Retry button without retry prop', () => {
    const w = mount(LoadError, { props: { message: 'boom' } })
    expect(w.find('button').exists()).toBe(false)
  })

  it('invokes retry prop on Retry click', async () => {
    const retry = vi.fn()
    const w = mount(LoadError, { props: { message: 'boom', retry } })
    await w.get('button').trigger('click')
    expect(retry).toHaveBeenCalledOnce()
    expect(w.get('button').text()).toBe('Retry')
  })
})
