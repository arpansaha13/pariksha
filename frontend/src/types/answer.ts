export interface MCQAnswer {
  optionIndex: number | undefined
}

export interface GeneralAnswer {
  text: string
}

export interface Answer {
  id: number
  exam_participant_id: number
  question_id: number
  answer: MCQAnswer | GeneralAnswer | null
  score_awarded: number
  comments: string
}

export type AnswerMinimal = Pick<Answer, 'id' | 'answer' | 'question_id'>
