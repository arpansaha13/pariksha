package dtos

type ExamResultDto struct {
	ID           int64  `json:"id"`
	ScoreAwarded int32  `json:"score_awarded"`
	Comments     string `json:"comments"`
}
