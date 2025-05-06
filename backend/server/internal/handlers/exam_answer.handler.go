package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

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

	var response []dtos.AnswerResponseDto
	for _, answer := range resp.Answers {
		response = append(response, dtos.AnswerResponseDto{
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

func GetAnswerForExam(w http.ResponseWriter, r *http.Request) {
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

	resp, err := examService.Client().GetAnswerForExam(ctx, &proto.GetAnswerRequest{
		ExamId:     examID,
		QuestionId: questionID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.AnswerMinimalResponseDto{
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
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
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

	upsertAnswerRequest := &proto.UpsertAnswersRequest{
		ExamId: examID,
		Answer: &proto.Answer{
			Answer:      nil,
			SubmittedAt: timestamppb.Now(),
			QuestionId:  answerDTO.QuestionID,
		},
	}

	if answerDTO.Answer != nil {
		upsertAnswerRequest.Answer.Answer = *answerDTO.Answer
	}

	examService := services.GetExamService()
	resp, err := examService.Client().UpsertAnswer(ctx, upsertAnswerRequest)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"answer_id": resp.AnswerId})
}
