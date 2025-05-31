// _____________________________USER______________________________
export interface User {
  id: UserId
  username: string
  email: string
  first_name?: string
  last_name?: string
}

// ____________________________PAPER______________________________
export interface PaperQuestionCounts {
  mcq: number
  subjective: number
}

export interface Paper {
  id: PaperId
  title: string
  max_score: number
  question_counts: PaperQuestionCounts
  duration_minutes: number
}

export interface PaperPermission {
  can_read: boolean
  can_write: boolean
}

// _____________________________EXAM______________________________
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

// ___________________________QUESTION____________________________
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

export enum QuestionCodingContentPrimitiveInputTypes {
  NUMBER = 1,
  STRING = 2,
  BOOLEAN = 3,
}

export enum QuestionCodingContentCompositeInputTypes {
  ARRAY = 4,
}

export type QuestionCodingContentInputTypes =
  | QuestionCodingContentPrimitiveInputTypes
  | QuestionCodingContentCompositeInputTypes

type QuestionCodingContentParameterItem = {
  /** only for input_type = object */
  property_name?: string

  /** Only primitive types can be used inside composite types */
  type: QuestionCodingContentPrimitiveInputTypes
}

export interface QuestionCodingContentParameter {
  type: QuestionCodingContentInputTypes
  items?: QuestionCodingContentParameterItem[]
}

export interface QuestionCodingContentInputDefinition
  extends QuestionCodingContentParameter {
  variable_name: string
}

export interface QuestionCodingContent {
  title: string
  statement: string
  input_definitions: QuestionCodingContentInputDefinition[]
  output_definition: QuestionCodingContentParameter
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

// __________________________EVALUATION___________________________
export interface EvaluationAnswer {
  id: AnswerId
  question_id: Answer['question_id']
  score_awarded: Answer['score_awarded']
  comments: Answer['comments']
}

// ____________________________ANSWER_____________________________
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

// ___________________QUESTION ANSWER COMBINED____________________
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
