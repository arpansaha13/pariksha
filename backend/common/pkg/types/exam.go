package types

type ExamQueuePayload struct {
	ExamID  int64 `json:"examId"`
	PaperID int64 `json:"paperId"`
}
