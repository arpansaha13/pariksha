package dtos

type GetAnswerEvaluationDataResponseDto struct {
	ID           int64  `json:"id"`
	QuestionID   int64  `json:"question_id"`
	ScoreAwarded int32  `json:"score_awarded"`
	Comments     string `json:"comments"`
}

type UpdateAnswerForEvaluationDto struct {
	NewScore  *int    `json:"new_score"`
	Evaluated *bool   `json:"evaluated"`
	Comments  *string `json:"comments"`
}

type EvaluationStatusResponseDto struct {
	UnevaluatedCount int32 `json:"unevaluated_count"`
}
