package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
	ctx := examService.CreateMetadata(r.Context(), userID)

	resp, err := examService.Client().GetParticipantAnswers(ctx, &proto.ParticipantRequest{
		ParticipantId: participantID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := make([]dtos.AnswerListItemDto, len(resp.Answers))
	for i, result := range resp.Answers {
		response[i] = dtos.AnswerListItemDto{
			Type: result.QuestionType,
			Question: dtos.AnswerListQuestionDto{
				ID:         result.QuestionHash,
				Order:      result.Order,
				CategoryID: result.CategoryId,
				Content:    result.Question,
				MaxScore:   result.MaxScore,
			},
			Answer: nil,
		}

		// AnswerId will be 0 is question is unanswered
		if result.AnswerId != 0 {
			response[i].Answer = &dtos.AnswerListAnswerDto{
				ID:      result.AnswerId,
				Content: result.Answer,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetAnswerForExam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	examHash, err := getExamIdFromVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionHash, err := getQuestionIdFromVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	resp, err := examService.Client().GetAnswerForExam(ctx, &proto.GetAnswerRequest{
		ExamHash:     examHash,
		QuestionHash: questionHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.AnswerMinimalResponseDto{
		ID:         resp.AnswerId,
		Answer:     resp.Answer,
		QuestionID: questionHash,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpsertAnswer(w http.ResponseWriter, r *http.Request) {
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	questionHash := answerDTO.QuestionID

	upsertAnswerRequest := &proto.UpsertAnswersRequest{
		ExamHash: examHash,
		Answer: &proto.Answer{
			Answer:       nil,
			SubmittedAt:  timestamppb.Now(),
			QuestionHash: questionHash,
		},
	}

	if answerDTO.Answer != nil {
		upsertAnswerRequest.Answer.Answer = *answerDTO.Answer
	}

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	resp, err := examService.Client().UpsertAnswer(ctx, upsertAnswerRequest)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(dtos.AnswerMinimalResponseDto{
		ID:         resp.AnswerId,
		Answer:     resp.Answer,
		QuestionID: questionHash,
	})
}
