import { vi } from 'vitest'

/**
 * Side-effect import: stubs every settings control SFC that fieldRegistry
 * (and SettingsPanel) pull in. Import this module at the top of the test file
 * BEFORE dynamically importing fieldRegistry/SettingsPanel.
 *
 * ```ts
 * import '@/test-utils/mockSettingsControls'
 * const { getRenderableFields } = await import('../composables/settings/fieldRegistry')
 * ```
 *
 * vi.mock calls MUST stay at module top-level (Vitest hoist rule).
 */
vi.mock('@/components/settings/SortControl.vue', () => ({
  default: { name: 'SortControl' },
}))
vi.mock('@/components/settings/ScaleControl.vue', () => ({
  default: { name: 'ScaleControl' },
}))
vi.mock('@/components/settings/BooleanControl.vue', () => ({
  default: { name: 'BooleanControl' },
}))
vi.mock('@/components/settings/SwapControl.vue', () => ({
  default: { name: 'SwapControl' },
}))
