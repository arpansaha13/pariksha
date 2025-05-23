export enum QuestionId {
  /** Special case for add question */
  ADD = 0,
}

export enum QuestionType {
  MCQ = 'MCQ',
  SUBJECTIVE = 'SUBJECTIVE',
  CODING = 'CODING',
}

export interface QuestionCategory {
  id: number
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
  id: number
  order: number
  category_id: number
  tags: string[]
  paper_id: number
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
