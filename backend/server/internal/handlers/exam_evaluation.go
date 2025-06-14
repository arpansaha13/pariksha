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

	// Get question ID from paper service
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	questionIDResp, err := paperService.Client().GetQuestionIds(paperCtx, &proto.GetQuestionIdsRequest{
		QuestionHashes: []string{questionHash},
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	if len(questionIDResp.QuestionIds) == 0 {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Call exam service with question ID
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	resp, err := examService.Client().GetAnswerEvaluationData(ctx, &proto.ParticipantQuestionRequest{
		ParticipantId: participantID,
		QuestionId:    questionIDResp.QuestionIds[0],
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	answer := dtos.GetAnswerEvaluationDataResponseDto{
		ID:           resp.AnswerId,
		QuestionID:   questionHash,
		ScoreAwarded: resp.ScoreAwarded,
		Comments:     resp.Comments,
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
	ctx := examService.CreateMetadata(userID)

	req := &proto.UpdateAnswerRequest{
		AnswerId: answerId,
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
	resp, err := examService.Client().UpdateAnswerForEvaluation(ctx, req)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Get question hashes from paper service
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	hashes, err := paperService.Client().GetQuestionHashes(paperCtx, &proto.GetQuestionHashesRequest{
		QuestionIds: []int64{resp.QuestionId},
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	updatedEvaluationData := dtos.GetAnswerEvaluationDataResponseDto{
		ID:           resp.AnswerId,
		QuestionID:   hashes.QuestionHashes[0],
		ScoreAwarded: resp.ScoreAwarded,
		Comments:     resp.Comments,
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
	ctx := examService.CreateMetadata(userID)

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

	// Get question ID from paper service
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	questionIDResp, err := paperService.Client().GetQuestionIds(paperCtx, &proto.GetQuestionIdsRequest{
		QuestionHashes: []string{questionHash},
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	if len(questionIDResp.QuestionIds) == 0 {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Call exam service with question ID
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	answer, err := examService.Client().GetAnswerForEvaluation(ctx, &proto.ParticipantQuestionRequest{
		ParticipantId: participantID,
		QuestionId:    questionIDResp.QuestionIds[0],
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
