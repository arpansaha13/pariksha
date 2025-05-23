/**
 * Helper function for array comparison
 */
export function arrayEquals<T>(a: T[], b: T[]): boolean {
  return (
    Array.isArray(a) &&
    Array.isArray(b) &&
    a.length === b.length &&
    a.every((val, index) => val === b[index])
  )
}

/**
 * Checks if a string contains only alphabetic characters (A-Z, a-z)
 * @param str - The string to check
 * @returns boolean - True if string contains only alphabets, false otherwise
 */
export const isAlpha = (str: string): boolean => {
  if (!str) return false
  return /^[A-Za-z]+$/.test(str)
}
