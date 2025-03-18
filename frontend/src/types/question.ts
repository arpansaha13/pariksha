export enum QuestionType {
  MCQ = 'MCQ',
  SHORT = 'SHORT',
  LONG = 'LONG',
}

export interface QuestionCategory {
  id: number
  name: string
  order: number
}

export interface QuestionMcq {
  id: number
  question: {
    statement: string
    options: string[]
  }
  category?: QuestionCategory
  type: QuestionType.MCQ
  tags: string[]
  paper_id: number
  max_score: number
  correct_answer?: string
}

export interface QuestionShort {
  id: number
  question: {
    statement: string
  }
  category?: QuestionCategory
  type: QuestionType.SHORT
  tags: string[]
  paper_id: number
  max_score: number
  correct_answer?: string
}

export interface QuestionLong {
  id: number
  question: {
    statement: string
  }
  category?: QuestionCategory
  type: QuestionType.LONG
  tags: string[]
  paper_id: number
  max_score: number
  correct_answer?: string
}

export type Question = QuestionMcq | QuestionShort | QuestionLong
