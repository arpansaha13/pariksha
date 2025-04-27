import type { PaperQuestionCounts } from '~/types'

export function countTotalQuestions(question_counts: PaperQuestionCounts) {
  let count = 0
  count += question_counts.mcq ?? 0
  count += question_counts.short ?? 0
  count += question_counts.long ?? 0
  return count
}

export function calcHours(durationMinutes: number) {
  return Math.floor(durationMinutes / 60)
}

export function calcRemainderMinutes(durationMinutes: number) {
  return durationMinutes % 60
}

export function convertToMinutes(hours = 0, minutes = 0) {
  return hours * 60 + minutes
}
