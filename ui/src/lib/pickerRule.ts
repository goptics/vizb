// Threshold for the chart-type picker in the settings panel. ≤ 3 chart types
// -> button row (SelectionTabs, full-width); > 3 -> combobox (Selector) with
// icon + name. Picked as a UX heuristic: 4 buttons in a row are cramped, 5
// don't fit. The exact value is a single constant — tune later if 3 turns
// out wrong. The function is exported separately so SettingsPanel's test can
// exercise it as a pure unit (vitest config has no Vue plugin; mounted
// component tests would need a heavier harness).
export const CHART_PICKER_TAB_THRESHOLD = 3

export const shouldUseTabPicker = (chartTypeCount: number): boolean =>
  chartTypeCount <= CHART_PICKER_TAB_THRESHOLD
