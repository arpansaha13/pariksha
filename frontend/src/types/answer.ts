import type { QuestionMcq, QuestionShort, QuestionType } from './question'

export interface MCQAnswer {
  optionIndex: number | undefined
}

export interface SubjectiveAnswer {
  text: string
}

export interface Answer {
  id: number
  exam_participant_id: number
  question_id: number

  /** `null` indicates that the question is unanswered */
  answer: MCQAnswer | SubjectiveAnswer | null
  score_awarded: number
  comments: string
}

export type AnswerMinimal = Pick<Answer, 'id' | 'answer' | 'question_id'>

type QuestionAnswerMCQ = {
  readonly type: QuestionType.MCQ
  readonly question: {
    readonly id: number
    readonly order: number
    readonly category_id: number
    readonly max_score: number
    readonly content: QuestionMcq['question']
  }
  readonly answer: {
    readonly id: number
    readonly content: MCQAnswer | null
  } | null
}

type QuestionAnswerSubjective = {
  readonly type: QuestionType.SUBJECTIVE
  readonly question: {
    readonly id: number
    readonly order: number
    readonly category_id: number
    readonly max_score: number
    readonly content: QuestionShort['question']
  }
  readonly answer: {
    readonly id: number
    readonly content: SubjectiveAnswer | null
  } | null
}

export type QuestionAnswer = QuestionAnswerMCQ | QuestionAnswerSubjective
