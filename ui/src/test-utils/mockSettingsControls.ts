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
vi.mock('@/components/settings/StackControl.vue', () => ({
  default: { name: 'StackControl' },
}))
vi.mock('@/components/settings/ShowLabelsControl.vue', () => ({
  default: { name: 'ShowLabelsControl' },
}))
vi.mock('@/components/settings/SmoothControl.vue', () => ({
  default: { name: 'SmoothControl' },
}))
vi.mock('@/components/settings/HorizontalControl.vue', () => ({
  default: { name: 'HorizontalControl' },
}))
vi.mock('@/components/settings/ThreeDRotateControl.vue', () => ({
  default: { name: 'ThreeDRotateControl' },
}))
vi.mock('@/components/settings/ThreeDControl.vue', () => ({
  default: { name: 'ThreeDControl' },
}))
vi.mock('@/components/settings/ThreeDVisualMapControl.vue', () => ({
  default: { name: 'ThreeDVisualMapControl' },
}))
vi.mock('@/components/settings/VisualMapControl.vue', () => ({
  default: { name: 'VisualMapControl' },
}))
vi.mock('@/components/settings/SwapControl.vue', () => ({
  default: { name: 'SwapControl' },
}))
vi.mock('@/components/settings/BorderRadiusControl.vue', () => ({
  default: { name: 'BorderRadiusControl' },
}))
