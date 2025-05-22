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

func GetPaperQuestions(w http.ResponseWriter, r *http.Request) {
	paperID, err := getInt64FromVars(mux.Vars(r), "paperId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetPaperQuestions(ctx, &proto.PaperRequest{
		PaperId: paperID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Convert proto response to HTTP response
	httpResponse := make([]dtos.QuestionMinimalResponseDto, len(response.Questions))
	for i, q := range response.Questions {
		httpResponse[i] = dtos.QuestionMinimalResponseDto{
			ID:         q.Id,
			CategoryID: q.CategoryId,
			PaperID:    q.PaperId,
			Order:      q.Order,
			Question:   q.RawQuestion,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResponse)
}

func GetPaperQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := getInt64FromVars(mux.Vars(r), "questionId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetPaperQuestion(ctx, &proto.QuestionRequest{
		QuestionId: questionID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	tags, _ := json.Marshal(response.Tags)

	httpResponse := dtos.QuestionResponseDto{
		ID:            response.Id,
		Question:      response.RawQuestion,
		CategoryID:    response.CategoryId,
		Type:          response.Type,
		Tags:          tags,
		PaperID:       response.PaperId,
		MaxScore:      response.MaxScore,
		CorrectAnswer: response.GetCorrectAnswer(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResponse)
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	paperID, err := getInt64FromVars(mux.Vars(r), "paperId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	var questionDto dtos.CreateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&questionDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(questionDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	var tags []string
	if err := json.Unmarshal(questionDto.Tags, &tags); err != nil {
		http.Error(w, "Invalid tags data", http.StatusBadRequest)
		return
	}

	// Get raw question bytes
	questionBytes, err := questionDto.Question.MarshalJSON()
	if err != nil {
		http.Error(w, "Invalid question data", http.StatusBadRequest)
		return
	}

	requestObj := proto.CreateQuestionRequest{
		PaperId:       paperID,
		RawQuestion:   questionBytes,
		CategoryId:    questionDto.CategoryID,
		Type:          questionDto.Type,
		Tags:          tags,
		MaxScore:      int32(questionDto.MaxScore),
		CorrectAnswer: &questionDto.CorrectAnswer,
	}

	response, err := paperService.Client().CreateQuestion(ctx, &requestObj)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	responseDto := dtos.CreateQuestionResponseDto{
		ID: response.Id,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseDto)
}

func UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := getInt64FromVars(mux.Vars(r), "questionId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	var updateDto dtos.UpdateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&updateDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	request := &proto.UpdateQuestionRequest{
		QuestionId: questionID,
	}

	// Set optional fields only if they are provided
	if updateDto.Type != "" {
		request.Type = &updateDto.Type
	}

	if updateDto.Question != nil {
		questionBytes, err := updateDto.Question.MarshalJSON()
		if err != nil {
			http.Error(w, "Invalid question data", http.StatusBadRequest)
			return
		}
		request.RawQuestion = questionBytes
	}

	if updateDto.CategoryID != 0 {
		request.CategoryId = &updateDto.CategoryID
	}

	if updateDto.MaxScore != 0 {
		maxScore := int32(updateDto.MaxScore)
		request.MaxScore = &maxScore
	}

	if updateDto.Tags != nil {
		var tags []string
		if err := json.Unmarshal(updateDto.Tags, &tags); err != nil {
			http.Error(w, "Invalid tags data", http.StatusBadRequest)
			return
		}
		request.Tags = tags
	}

	if updateDto.CorrectAnswer != "" {
		request.CorrectAnswer = &updateDto.CorrectAnswer
	}

	response, err := paperService.Client().UpdateQuestion(ctx, request)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	responseDto := dtos.UpdateQuestionResponseDto{
		ID: response.QuestionId,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseDto)
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := getInt64FromVars(mux.Vars(r), "questionId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	_, err = paperService.Client().DeleteQuestion(ctx, &proto.QuestionRequest{
		QuestionId: questionID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ReorderQuestions(w http.ResponseWriter, r *http.Request) {
	categoryID, err := getInt64FromVars(mux.Vars(r), "category_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	var reorderDto dtos.ReorderQuestionsDto
	if err := json.NewDecoder(r.Body).Decode(&reorderDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(reorderDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	questionIDs := make([]int64, len(reorderDto.Questions))
	for i, id := range reorderDto.Questions {
		questionIDs[i] = id
	}

	_, err = paperService.Client().ReorderQuestions(ctx, &proto.ReorderQuestionsRequest{
		CategoryId:  categoryID,
		QuestionIds: questionIDs,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
