export enum QuestionId {
  /** Special case for add question */
  ADD = 0,
}

export enum QuestionType {
  MCQ = 'MCQ',
  SUBJECTIVE = 'SUBJECTIVE',
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
  order: number
  category_id: number
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
  order: number
  category_id: number
  type: QuestionType.SUBJECTIVE
  tags: string[]
  paper_id: number
  max_score: number
  correct_answer: string | null | undefined
}

export type Question = QuestionMcq | QuestionShort

export interface QuestionMinimal {
  id: number
  category_id: number
  order: number
  paper_id: number
  question: Question['question']
}
