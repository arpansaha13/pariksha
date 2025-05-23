export const QUESTION_ID_ADD = 0 as QuestionId

export enum QuestionType {
  MCQ = 'MCQ',
  SUBJECTIVE = 'SUBJECTIVE',
  CODING = 'CODING',
}

export interface QuestionCategory {
  id: CategoryId
  name: string
  order: number
}

export interface QuestionMcqContent {
  statement: string
  options: string[]
}

export interface QuestionSubjectiveContent {
  statement: string
}

export interface QuestionCodingContentExample {
  input: string
  output: string
  explanation?: string
}

export interface QuestionCodingContent {
  title: string
  statement: string
  examples?: QuestionCodingContentExample[]
}

export interface BaseQuestion {
  id: QuestionId
  order: number
  category_id: CategoryId
  tags: string[]
  paper_id: PaperId
  max_score: number
  correct_answer?: string
}

export interface QuestionMcq extends BaseQuestion {
  type: QuestionType.MCQ
  question: QuestionMcqContent
}

export interface QuestionSubjective extends BaseQuestion {
  type: QuestionType.SUBJECTIVE
  question: QuestionSubjectiveContent
}

export interface QuestionCoding extends BaseQuestion {
  type: QuestionType.CODING
  question: QuestionCodingContent
}

export type Question = QuestionMcq | QuestionSubjective | QuestionCoding

export type QuestionMinimal = Pick<
  Question,
  'id' | 'category_id' | 'order' | 'paper_id' | 'question'
>
