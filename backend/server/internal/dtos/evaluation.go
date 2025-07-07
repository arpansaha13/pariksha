package dtos

type GetAnswerEvaluationDataResponseDto struct {
	ID           int64  `json:"id"`
	QuestionID   string `json:"question_id"`
	ScoreAwarded int32  `json:"score_awarded"`
}

type UpdateAnswerForEvaluationDto struct {
	NewScore  *int  `json:"new_score"`
	Evaluated *bool `json:"evaluated"`
}

type EvaluationStatusResponseDto struct {
	UnevaluatedCount int32 `json:"unevaluated_count"`
}
