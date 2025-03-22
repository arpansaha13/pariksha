export enum QuestionId {
  /** Special case for add question */
  ADD = 0,
}

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
  category: QuestionCategory
  type: QuestionType.MCQ
  tags: string[]
  paper_id: number
  max_score: number
  correct_answer: string | null | undefined
}

export interface QuestionShort {
  id: number
  question: {
    statement: string
  }
  category: QuestionCategory
  type: QuestionType.SHORT
  tags: string[]
  paper_id: number
  max_score: number
  correct_answer: string | null | undefined
}

export interface QuestionLong {
  id: number
  question: {
    statement: string
  }
  category: QuestionCategory
  type: QuestionType.LONG
  tags: string[]
  paper_id: number
  max_score: number
  correct_answer: string | null | undefined
}

export type Question = QuestionMcq | QuestionShort | QuestionLong

export interface QuestionMinimal {
  id: number
  category_id: number
  paper_id: number
  question: Question['question']
}
