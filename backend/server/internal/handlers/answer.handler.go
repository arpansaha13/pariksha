package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"context"
	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetParticipantAnswers(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	resp, err := examService.Client().GetParticipantAnswers(ctx, &proto.ParticipantRequest{
		ParticipantId: participantID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	var response []dtos.AnswerResponse
	for _, answer := range resp.Answers {
		response = append(response, dtos.AnswerResponse{
			ID:                answer.Id,
			ExamParticipantID: answer.ExamParticipantId,
			QuestionID:        answer.QuestionId,
			Answer:            answer.Answer,
			Comments:          answer.Comments,
			ScoreAwarded:      int(answer.ScoreAwarded),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetAnswer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := getInt64FromVars(vars, "examId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionID, err := getInt64FromVars(vars, "questionId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	resp, err := examService.Client().GetAnswer(ctx, &proto.GetAnswerRequest{
		ExamId:     examID,
		QuestionId: questionID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.PartialAnswerResponse{
		ID:         resp.Id,
		Answer:     resp.Answer,
		QuestionID: resp.QuestionId,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpsertAnswer(w http.ResponseWriter, r *http.Request) {
	var answerDTO dtos.UpsertAnswerDto
	if err := json.NewDecoder(r.Body).Decode(&answerDTO); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(answerDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examID, err := getInt64FromVars(mux.Vars(r), "examId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch question from paper service to get question type
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	questionResp, err := paperService.Client().GetExamQuestion(paperCtx, &proto.QuestionRequest{
		QuestionId: answerDTO.QuestionID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Create metadata with question type
	md := metadata.New(map[string]string{
		"user_id":       strconv.FormatInt(userID, 10),
		"question_type": questionResp.Type,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	examService := services.GetExamService()
	resp, err := examService.Client().UpsertAnswer(ctx, &proto.UpsertAnswersRequest{
		ExamId: examID,
		Answer: &proto.Answer{
			Answer:      answerDTO.Answer,
			SubmittedAt: timestamppb.Now(),
			QuestionId:  answerDTO.QuestionID,
		},
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"answer_id": resp.AnswerId})
}

func UpdateAnswerForEvaluation(w http.ResponseWriter, r *http.Request) {
	var updateDTO dtos.UpdateAnswerForEvaluationDTO
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(updateDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	// Fetch answer to get question ID
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	answerResp, err := examService.Client().GetAnswerById(ctx, &proto.GetAnswerByIdRequest{
		AnswerId: updateDTO.AnswerID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Fetch question details
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)
	questionResp, err := paperService.Client().GetQuestion(paperCtx, &proto.QuestionRequest{
		QuestionId: answerResp.QuestionId,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Create metadata with question max score
	md := metadata.New(map[string]string{
		"user_id":        strconv.FormatInt(userID, 10),
		"question_score": strconv.Itoa(int(questionResp.MaxScore)),
	})
	ctx = metadata.NewOutgoingContext(context.Background(), md)

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
	_, err = examService.Client().UpdateAnswerForEvaluation(ctx, req)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func MarkAsEvaluated(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	resp, err := examService.Client().MarkAsEvaluated(ctx, &proto.ParticipantRequest{
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
