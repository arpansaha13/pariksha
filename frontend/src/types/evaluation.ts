import type { Answer } from './answer'

export interface EvaluationAnswer {
  id: Answer['id']
  question_id: Answer['question_id']
  score_awarded: Answer['score_awarded']
  comments: Answer['comments']
}
