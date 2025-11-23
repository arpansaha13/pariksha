package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/dtos"
	"pariksha/gateway/internal/middlewares"
	"pariksha/gateway/internal/services"
)

// GetExamResults handles HTTP request to get all questions and answers for an exam
func GetExamResults(w http.ResponseWriter, r *http.Request) {
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	results, err := examService.Client().GetExamResults(ctx, &proto.ExamRequest{
		ExamHash: examHash,
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
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
