import { describe, it, expect, vi, beforeEach } from 'vitest'
import { TrackedMockWorker } from './workers/__test-utils__/workerHarness'

// The real Dashboard constructs a TransformWorker in useChartPipeline; happy-dom
// has no Worker global, so route it through the shared test harness.
vi.mock('./workers/transform.worker.ts?worker&inline', () => ({
  default: TrackedMockWorker,
}))

describe('App bootstrap', () => {
  beforeEach(() => {
    document.body.innerHTML = '<div id="app"></div>'
  })

  it('mounts the real Dashboard without crashing', async () => {
    // Boots the app exactly as production does (createApp(App).mount('#app')).
    await import('./main')
    await new Promise((resolve) => setTimeout(resolve, 0))

    // With no VIZB_DATA the shell still renders (empty dataset placeholder),
    // proving App → Dashboard boots end to end. The worker-backed pipeline runs
    // without a Worker global crash (the mock worker is constructed in its place).
    expect(document.querySelector('[data-testid="page-shell"]')).not.toBeNull()
    expect(TrackedMockWorker.instances.length).toBeGreaterThan(0)
  })
})
