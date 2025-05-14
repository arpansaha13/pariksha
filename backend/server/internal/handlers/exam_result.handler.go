package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

// GetExamResults handles HTTP request to get all questions and answers for an exam
func GetExamResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.ParseInt(vars["examId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid exam id", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	results, err := examService.Client().GetExamResults(ctx, &proto.ExamRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Get question details from paper service
	questionIDs := make([]int64, len(results.Results))
	for i, result := range results.Results {
		questionIDs[i] = result.QuestionId
	}

	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	questionDetails, err := paperService.Client().GetQuestionsByIds(paperCtx, &proto.GetQuestionsByIdsRequest{
		QuestionIds: questionIDs,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Create a map for quick lookup of question details
	questionDetailsMap := make(map[int64]*proto.QuestionBatchItem)
	for _, q := range questionDetails.Questions {
		questionDetailsMap[q.Id] = q
	}

	response := make([]dtos.ExamResultItemDTO, len(results.Results))

	for i, result := range results.Results {
		var questionBytes []byte
		questionDetail := questionDetailsMap[result.QuestionId]
		if questionDetail != nil {
			switch q := questionDetail.Question.(type) {
			case *proto.QuestionBatchItem_Mcq:
				questionBytes, _ = json.Marshal(q.Mcq)
			case *proto.QuestionBatchItem_General:
				questionBytes, _ = json.Marshal(q.General)
			}
		}

		response[i] = dtos.ExamResultItemDTO{
			Type: questionDetail.GetType(),
			Question: dtos.ExamResultQuestionDTO{
				ID:         result.QuestionId,
				Order:      result.Order,
				CategoryID: result.CategoryId,
				Content:    questionBytes,
				MaxScore:   result.MaxScore,
			},
			Answer: dtos.ExamResultAnswerDTO{
				Content:      result.Answer,
				ScoreAwarded: result.ScoreAwarded,
				Comments:     result.Comments,
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
