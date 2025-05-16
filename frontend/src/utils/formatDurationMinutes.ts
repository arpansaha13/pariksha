import { isNullOrUndefined } from '@arpansaha13/utils'

export function formatDurationMinutes(durationMinutes: number) {
  if (isNullOrUndefined(durationMinutes)) return '0 minutes'

  const hours = calcHours(durationMinutes)
  const remainingMinutes = calcRemainderMinutes(durationMinutes)

  if (hours === 0) return `${remainingMinutes} minutes`
  if (remainingMinutes === 0)
    return `${hours} ${hours === 1 ? 'hour' : 'hours'}`
  if (hours === 1)
    return `${hours} hour ${remainingMinutes} ${remainingMinutes === 1 ? 'minute' : 'minutes'}`
  return `${hours} hours  ${remainingMinutes} minutes`
}
