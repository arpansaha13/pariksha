import type {
  QuestionMcqContent,
  QuestionSubjectiveContent,
  QuestionType,
} from './question'
export interface MCQAnswer {
  optionIndex: number | undefined
}

export interface SubjectiveAnswer {
  text: string
}

export interface Answer {
  id: AnswerId
  exam_participant_id: ExamParticipantId
  question_id: QuestionId

  /** `null` indicates that the question is unanswered */
  answer: MCQAnswer | SubjectiveAnswer | null
  score_awarded: number
  comments: string
}

export type AnswerMinimal = Pick<Answer, 'id' | 'answer' | 'question_id'>

type QuestionAnswerMCQ = {
  readonly type: QuestionType.MCQ
  readonly question: {
    readonly id: QuestionId
    readonly order: number
    readonly category_id: number
    readonly max_score: number
    readonly content: QuestionMcqContent
  }
  readonly answer: {
    readonly id: AnswerId
    readonly content: MCQAnswer | null
  } | null
}

type QuestionAnswerSubjective = {
  readonly type: QuestionType.SUBJECTIVE
  readonly question: {
    readonly id: QuestionId
    readonly order: number
    readonly category_id: number
    readonly max_score: number
    readonly content: QuestionSubjectiveContent
  }
  readonly answer: {
    readonly id: AnswerId
    readonly content: SubjectiveAnswer | null
  } | null
}

export type QuestionAnswer = QuestionAnswerMCQ | QuestionAnswerSubjective
