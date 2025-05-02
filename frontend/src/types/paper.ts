export interface PaperQuestionCounts {
  mcq: number
  short: number
  long: number
}

export interface Paper {
  id: number
  title: string
  max_score: number
  question_counts: PaperQuestionCounts
  duration_minutes: number
}
