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
