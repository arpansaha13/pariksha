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

	response := make([]dtos.ExamResultDto, len(results.Results))

	for i, answer := range results.Results {
		response[i] = dtos.ExamResultDto{
			ID:           answer.AnswerId,
			ScoreAwarded: answer.ScoreAwarded,
			Comments:     answer.Comments,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
