enum PaperOwnership {
  OWNER = 'OWNER',
  SHARED = 'SHARED',
}

interface PaperQuestionCounts {
  mcq: number
  short: number
  long: number
}

export interface Paper {
  id: number
  title: string
  maxScore: number
  questionCounts: PaperQuestionCounts
  paperOwnership: {
    id: number
    path: string
    type: PaperOwnership
  }
}
