package structs

type PrepareQuestionsPayload struct {
	ExamID  int64 `json:"examId"`
	PaperID int64 `json:"paperId"`
}

type AutoEndExamPayload struct {
	ExamID        int64 `json:"exam_id"`
	ParticipantID int64 `json:"participant_id"`
}

type DeleteExamsPayload struct {
	ExamIDs []int64 `json:"examIds"`
}
