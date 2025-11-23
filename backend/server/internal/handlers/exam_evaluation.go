package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetAnswerEvaluationData(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	questionHash, err := getQuestionIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	resp, err := examService.Client().GetAnswerEvaluationData(ctx, &proto.ParticipantQuestionRequest{
		ParticipantId: participantID,
		QuestionHash:  questionHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	answer := dtos.GetAnswerEvaluationDataResponseDto{
		ID:           resp.AnswerId,
		QuestionID:   questionHash,
		ScoreAwarded: resp.ScoreAwarded,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func UpdateAnswerForEvaluation(w http.ResponseWriter, r *http.Request) {
	answerId, err := getInt64FromVars(mux.Vars(r), "answerId")
	if err != nil {
		http.Error(w, "Invalid answer ID", http.StatusBadRequest)
		return
	}

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
	ctx := examService.CreateMetadata(r.Context(), userID)

	req := &proto.UpdateAnswerRequest{
		AnswerId:  answerId,
		Evaluated: updateDTO.Evaluated,
		NewScore:  ptr.Int32(*updateDTO.NewScore),
	}

	// Send update request to exam service
	resp, err := examService.Client().UpdateAnswerForEvaluation(ctx, req)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	updatedEvaluationData := dtos.GetAnswerEvaluationDataResponseDto{
		ID:           resp.AnswerId,
		QuestionID:   resp.QuestionHash,
		ScoreAwarded: resp.ScoreAwarded,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedEvaluationData)
}

func MarkParticipantAsEvaluated(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	resp, err := examService.Client().MarkParticipantAsEvaluated(ctx, &proto.ParticipantRequest{
		ParticipantId: participantID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dtos.EvaluationStatusResponseDto{
		UnevaluatedCount: resp.UnevaluatedCount,
	})
}

func GetAnswerForEvaluation(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	questionHash, err := getQuestionIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	answer, err := examService.Client().GetAnswerForEvaluation(ctx, &proto.ParticipantQuestionRequest{
		ParticipantId: participantID,
		QuestionHash:  questionHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.AnswerMinimalResponseDto{
		ID:         answer.AnswerId,
		Answer:     answer.Answer,
		QuestionID: questionHash,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
