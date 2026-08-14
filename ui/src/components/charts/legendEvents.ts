export type LegendSelectChangedEvent = { selected: Record<string, boolean> }

const isBooleanRecord = (value: unknown): value is Record<string, boolean> =>
  !!value &&
  typeof value === 'object' &&
  !Array.isArray(value) &&
  Object.values(value).every((entry) => typeof entry === 'boolean')

export const createLegendSelectChangedForwarder =
  (forward: (event: LegendSelectChangedEvent) => void) => (event: unknown) => {
    if (isBooleanRecord((event as { selected?: unknown } | null)?.selected)) {
      forward(event as LegendSelectChangedEvent)
    }
  }
