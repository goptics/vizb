import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

const mount = vi.fn()
const createApp = vi.fn<(root: unknown) => { mount: typeof mount }>(() => ({ mount }))

vi.mock('vue', async () => {
  const actual = await vi.importActual<typeof import('vue')>('vue')
  return {
    ...actual,
    createApp,
  }
})

vi.mock('./App.vue', () => ({
  default: { name: 'AppStub' },
}))

vi.mock('./assets/globals.css', () => ({}))

describe('main.ts', () => {
  beforeEach(() => {
    mount.mockClear()
    createApp.mockClear()
    vi.resetModules()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('creates the app and mounts #app', async () => {
    await import('./main')
    expect(createApp).toHaveBeenCalledTimes(1)
    expect(createApp.mock.calls[0]?.[0]).toMatchObject({ name: 'AppStub' })
    expect(mount).toHaveBeenCalledWith('#app')
  })
})
