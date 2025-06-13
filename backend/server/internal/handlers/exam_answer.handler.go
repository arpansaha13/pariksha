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
				ID:         questionDetail.QuestionHash,
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

	resp, err := examService.Client().GetAnswerForExam(ctx, &proto.GetAnswerRequest{
		ExamHash:   examHash,
		QuestionId: questionIDResp.QuestionIds[0],
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

	// Get question ID from paper service
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)
	questionHash := answerDTO.QuestionID

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
	upsertAnswerRequest := &proto.UpsertAnswersRequest{
		ExamHash: examHash,
		Answer: &proto.Answer{
			Answer:      nil,
			SubmittedAt: timestamppb.Now(),
			QuestionId:  questionIDResp.QuestionIds[0],
		},
	}

	if answerDTO.Answer != nil {
		upsertAnswerRequest.Answer.Answer = *answerDTO.Answer
	}

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

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
