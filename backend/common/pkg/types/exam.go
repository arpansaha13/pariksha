package types

type ExamQueuePayload struct {
	ExamID  int `json:"examId"`
	PaperID int `json:"paperId"`
}
