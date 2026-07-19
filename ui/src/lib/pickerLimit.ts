export type PickerOption = {
  value: string
  label: string
}

export const limitPickerOptions = <T extends PickerOption>(
  matches: T[],
  active: T | undefined,
  limit: number
): T[] => {
  const limited = matches.slice(0, limit)
  if (active && !limited.some((option) => option.value === active.value)) {
    if (limited.length < limit) limited.push(active)
    else limited[limited.length - 1] = active
  }
  return limited
}
