package structs

import (
	"pariksha/common/pkg/types"
)

type PrepareQuestionsPayload struct {
	ExamID  types.ExamID  `json:"examId"`
	PaperID types.PaperID `json:"paperId"`
}

type AutoEndExamPayload struct {
	ExamID        types.ExamID        `json:"exam_id"`
	ParticipantID types.ParticipantID `json:"participant_id"`
}

type DeleteExamsPayload struct {
	ExamIDs []types.ExamID `json:"examIds"`
}
