/**
 * Logs a warning message in development mode.
 *
 * If `import.meta.dev` is `true`, this function emits a stack trace
 * with the provided warning message. Useful for debugging purposes.
 *
 * @param {string} message - The warning message to be logged.
 * @returns {void}
 */
export function logWarning(message: string): void {
  if (import.meta.dev) {
    console.trace(message)
  }
}

/**
 * Logs a warning if a variable is null or undefined in development mode.
 *
 * This function checks if `import.meta.dev` is `true` and emits a stack trace
 * with the provided variable name. It is useful for debugging purposes.
 *
 * @param {string} variable_name - The name of the variable being checked.
 * @returns {void}
 */
export function logNullOrUndefinedWarning(variable_name: string): void {
  if (import.meta.dev) {
    console.trace(`${variable_name} is null or undefined`)
  }
}
