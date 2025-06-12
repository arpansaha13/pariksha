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
	"pariksha/server/internal/utils"
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

	// Get question details from paper service
	questionIDs := make([]int64, len(resp.Answers))
	for i, result := range resp.Answers {
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
		questionDetailsMap[q.QuestionId] = q
	}

	response := make([]dtos.AnswerListItemDto, len(resp.Answers))
	for i, result := range resp.Answers {
		questionDetail := questionDetailsMap[result.QuestionId]
		var questionBytes []byte
		if questionDetail != nil {
			questionBytes = questionDetail.RawQuestion
		}

		response[i] = dtos.AnswerListItemDto{
			Type: result.QuestionType,
			Question: dtos.AnswerListQuestionDto{
				ID:         result.QuestionId,
				Order:      result.Order,
				CategoryID: result.CategoryId,
				Content:    questionBytes,
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
	examID := r.Context().Value(middlewares.DecryptedExamID).(int64)
	questionID := r.Context().Value(middlewares.DecryptedQuestionID).(int64)

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

	encryptedQuestionId, _ := utils.EncryptID(resp.QuestionId)
	response := dtos.AnswerMinimalResponseDto{
		ID:         resp.AnswerId,
		Answer:     resp.Answer,
		QuestionID: encryptedQuestionId,
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

	examID := r.Context().Value(middlewares.DecryptedExamID).(int64)

	decryptedQuestionId, err := utils.DecryptID(answerDTO.QuestionID)
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}
	upsertAnswerRequest := &proto.UpsertAnswersRequest{
		ExamId: examID,
		Answer: &proto.Answer{
			Answer:      nil,
			SubmittedAt: timestamppb.Now(),
			QuestionId:  decryptedQuestionId,
		},
	}

	if answerDTO.Answer != nil {
		upsertAnswerRequest.Answer.Answer = *answerDTO.Answer
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	resp, err := examService.Client().UpsertAnswer(ctx, upsertAnswerRequest)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encryptedQuestionId, _ := utils.EncryptID(resp.QuestionId)
	json.NewEncoder(w).Encode(dtos.AnswerMinimalResponseDto{
		ID:         resp.AnswerId,
		Answer:     resp.Answer,
		QuestionID: encryptedQuestionId,
	})
}
