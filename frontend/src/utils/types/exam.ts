import type { Question, QuestionCategory } from './question'

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
  id: ExamId
  title: string
  starts_at: string
  ends_at: string
  created_by: UserId
  type: ExamAccessType
  max_candidates_count: number
  max_score: number
  duration_minutes: number
}

export interface ExamPermission {
  can_read: boolean
  can_write: boolean
  can_participate: boolean
  can_evaluate: boolean
  participant_status?: ExamParticipantStatus
}

export type ExamQuestion = Pick<Question, 'id' | 'question'>

export type ExamQuestionMinimal = Pick<
  Question,
  'id' | 'category_id' | 'order' | 'max_score' | 'type'
>

export type ExamCategory = Pick<QuestionCategory, 'id' | 'name' | 'order'>

export interface ExamParticipant {
  readonly id: ExamParticipantId
  readonly score_awarded: number
  readonly started_at: string
  readonly scheduled_end_time: string
}

export interface ExamParticipantById {
  readonly id: ExamParticipantId
  readonly status: ExamParticipantStatus
}

export interface ExamParticipantResponse {
  id: ExamParticipantId
  user_id: UserId
  status: ExamParticipantStatus
  score_awarded: number
  started_at?: string
  ended_at?: string
  scheduled_end_time?: string
  first_name?: string
  last_name?: string
  email?: string
}

export type ExamResult = {
  readonly id: AnswerId
  readonly score_awarded: number
  readonly comments: string
}
