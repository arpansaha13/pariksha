export interface EvaluationAnswer {
  id: AnswerId
  question_id: Answer['question_id']
  score_awarded: Answer['score_awarded']
  comments: Answer['comments']
}
