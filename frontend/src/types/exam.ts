import type { Question, QuestionCategory, QuestionMinimal } from './question'

export enum ExamAccessType {
  LINK = 'LINK',
  INVITE = 'INVITE',
}

export enum ExamParticipantStatus {
  UNATTENDED = 0,
  INVITED = 1,
  STARTED = 2,
  ENDED = 3,
  EVALUATED = 4,
}

export interface Exam {
  id: number
  title: string
  starts_at: string
  ends_at: string
  created_by: number
  type: ExamAccessType
  max_candidates_count: number
  paper_id: number
  duration_minutes: number
}

export enum ExamPermission {
  OWNER = 'OWNER',
  PARTICIPANT = 'PARTICIPANT',
}

export type ExamQuestion = Pick<
  Question,
  'id' | 'max_score' | 'question' | 'type' | 'category_id'
>

export type ExamQuestionMinimal = Pick<
  QuestionMinimal,
  'id' | 'category_id' | 'order'
>

export type ExamCategory = Pick<QuestionCategory, 'id' | 'name' | 'order'>
