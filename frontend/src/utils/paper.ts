export function countTotalQuestions(question_counts: PaperQuestionCounts) {
  return _sum(Object.values(question_counts))
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
