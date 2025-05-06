package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetAnswerForEvaluation(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	questionID, err := getInt64FromVars(mux.Vars(r), "questionId")
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	resp, err := examService.Client().GetAnswerForEvaluation(ctx, &proto.GetAnswerForEvaluationRequest{
		ParticipantId: participantID,
		QuestionId:    questionID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	answer := dtos.AnswerResponseDto{
		ID:                resp.Id,
		ExamParticipantID: resp.ExamParticipantId,
		QuestionID:        resp.QuestionId,
		Answer:            resp.Answer,
		ScoreAwarded:      int(resp.ScoreAwarded),
		Comments:          resp.Comments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func UpdateAnswerForEvaluation(w http.ResponseWriter, r *http.Request) {
	var updateDTO dtos.UpdateAnswerForEvaluationDto
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(updateDTO); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	req := &proto.UpdateAnswerRequest{
		AnswerId: updateDTO.AnswerID,
	}

	if updateDTO.NewScore != nil {
		score := int32(*updateDTO.NewScore)
		req.NewScore = &score
	}
	if updateDTO.Evaluated != nil {
		req.Evaluated = updateDTO.Evaluated
	}
	if updateDTO.Comments != nil {
		req.Comments = updateDTO.Comments
	}

	// Send update request to exam service
	_, err := examService.Client().UpdateAnswerForEvaluation(ctx, req)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func MarkParticipantAsEvaluated(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	resp, err := examService.Client().MarkParticipantAsEvaluated(ctx, &proto.ParticipantRequest{
		ParticipantId: participantID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := map[string]int32{
		"unevaluatedCount": resp.UnevaluatedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
