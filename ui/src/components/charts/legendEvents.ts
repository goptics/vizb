export interface LegendSelectChangedEvent {
  selected: Record<string, boolean>
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value)

const isBooleanRecord = (value: unknown): value is Record<string, boolean> =>
  isRecord(value) && Object.values(value).every((entry) => typeof entry === 'boolean')

export const isLegendSelectChangedEvent = (event: unknown): event is LegendSelectChangedEvent =>
  isRecord(event) && isBooleanRecord(event.selected)

export const createLegendSelectChangedForwarder = (
  forward: (event: LegendSelectChangedEvent) => void
) => {
  return (event: unknown) => {
    if (isLegendSelectChangedEvent(event)) {
      forward(event)
    }
  }
}
