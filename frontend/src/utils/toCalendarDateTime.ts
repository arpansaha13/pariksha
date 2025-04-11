import { CalendarDateTime } from '@internationalized/date'

export function toCalendarDateTime(date: Date | string): CalendarDateTime {
  const dateObj = typeof date === 'string' ? new Date(date) : date

  return new CalendarDateTime(
    dateObj.getFullYear(),
    dateObj.getMonth() + 1, // getMonth() is 0-indexed
    dateObj.getDate()
  )
}
