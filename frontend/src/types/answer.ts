export interface MCQAnswer {
  optionIndex: number | undefined
}

export interface GeneralAnswer {
  text: string
}

export type AnswerData = MCQAnswer | GeneralAnswer

export interface Answer {
  id: number
  exam_participant_id: number
  question_id: number
  answer: AnswerData
  score_awarded: number
  comments: string
}

export type AnswerMinimal = Pick<Answer, 'id' | 'answer' | 'question_id'>
