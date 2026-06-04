import { computed } from 'vue'

// Returns a writable computed that reads from the store and writes via setter.
// Replaces the ref + watch + handleChange triples in settings components.
export function useSyncedSetting<T>(read: () => T, write: (val: T) => void) {
  return computed<T>({
    get: read,
    set: write,
  })
}
