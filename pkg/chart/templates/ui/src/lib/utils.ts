import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * Utility function to merge Tailwind CSS classes
 * Used for conditional styling with shadcn components
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
