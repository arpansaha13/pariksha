export interface PaperQuestionCounts {
  mcq: number
  subjective: number
}

export interface Paper {
  id: PaperId
  title: string
  max_score: number
  question_counts: PaperQuestionCounts
  duration_minutes: number
}

export interface PaperPermission {
  can_read: boolean
  can_write: boolean
}
