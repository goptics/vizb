export type PickerOption = {
  value: string
  label: string
}

export const limitPickerOptions = <T extends PickerOption>(
  matches: T[],
  active: T | undefined,
  limit: number
): T[] => {
  if (limit <= 0) {
    if (!active || matches.some((option) => option.value === active.value)) return matches
    return [...matches, active]
  }

  const limited = matches.slice(0, limit)
  if (active && !limited.some((option) => option.value === active.value)) {
    if (limited.length < limit) limited.push(active)
    else limited[limited.length - 1] = active
  }
  return limited
}

export const filterPickerOptions = <T extends PickerOption>(options: T[], search: string): T[] => {
  const query = search.trim().toLowerCase()
  if (!query) return options
  return options.filter((option) => option.label.toLowerCase().includes(query))
}
