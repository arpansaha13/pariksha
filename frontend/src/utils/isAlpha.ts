/**
 * Checks if a string contains only alphabetic characters (A-Z, a-z)
 * @param str - The string to check
 * @returns boolean - True if string contains only alphabets, false otherwise
 */
export const isAlpha = (str: string): boolean => {
  if (!str) return false
  return /^[A-Za-z]+$/.test(str)
}
