import { CalendarDateTime } from '@internationalized/date'

export function toCalendarDateTime(date: Date | string): CalendarDateTime {
  const dateObj = typeof date === 'string' ? new Date(date) : date

  return new CalendarDateTime(
    dateObj.getFullYear(),
    dateObj.getMonth() + 1, // getMonth() is 0-indexed
    dateObj.getDate(),
    dateObj.getHours(),
    dateObj.getMinutes()
  )
}

export function isCalendarBefore(
  a: CalendarDateTime | Ref<CalendarDateTime>,
  b: CalendarDateTime | Ref<CalendarDateTime>
): boolean {
  return unref(a).compare(unref(b)) < 0
}

export function isCalendarAfter(
  a: CalendarDateTime | Ref<CalendarDateTime>,
  b: CalendarDateTime | Ref<CalendarDateTime>
): boolean {
  return unref(a).compare(unref(b)) > 0
}
