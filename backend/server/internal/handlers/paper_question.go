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
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetPaperQuestions(ctx, &proto.PaperRequest{
		PaperHash: paperHash,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Convert proto response to HTTP response
	httpResponse := make([]dtos.QuestionMinimalResponseDto, len(response.Questions))
	for i, q := range response.Questions {
		httpResponse[i] = dtos.QuestionMinimalResponseDto{
			ID:         q.QuestionHash,
			CategoryID: q.CategoryId,
			PaperID:    q.PaperHash,
			Order:      q.Order,
			Question:   q.RawQuestion,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResponse)
}

func GetPaperQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionHash, err := getQuestionIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetPaperQuestion(ctx, &proto.PaperQuestionRequest{
		PaperHash:    paperHash,
		QuestionHash: questionHash,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	httpResponse := dtos.QuestionResponseDto{
		ID:         response.QuestionHash,
		Question:   response.RawQuestion,
		CategoryID: response.CategoryId,
		Type:       response.Type,
		PaperID:    response.PaperHash,
		MaxScore:   response.MaxScore,
		TestCases:  nil,
	}

	if response.Type == proto.QuestionType_CODING {
		testCases := make([]dtos.PaperTestCaseDto, 0, len(response.TestCases))
		for _, tc := range response.TestCases {
			testCases = append(testCases, dtos.PaperTestCaseDto{
				Inputs:      tc.Inputs,
				Output:      tc.Output,
				Explanation: tc.Explanation,
				Hidden:      tc.Hidden,
				Order:       tc.Order,
			})
		}

		httpResponse.TestCases = &testCases
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResponse)
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	// Get raw question bytes
	questionBytes, err := questionDto.Question.MarshalJSON()
	if err != nil {
		http.Error(w, "Invalid question data", http.StatusBadRequest)
		return
	}

	requestObj := proto.CreatePaperQuestionRequest{
		PaperHash:   paperHash,
		RawQuestion: questionBytes,
		CategoryId:  questionDto.CategoryID,
		Type:        questionDto.Type,
		MaxScore:    int32(questionDto.MaxScore),
	}

	response, err := paperService.Client().CreatePaperQuestion(ctx, &requestObj)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	responseDto := dtos.CreateQuestionResponseDto{
		ID: response.QuestionHash,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseDto)
}

func UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionHash, err := getQuestionIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var updateDto dtos.UpdateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&updateDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	request := &proto.UpdatePaperQuestionRequest{
		PaperHash:    paperHash,
		QuestionHash: questionHash,
	}

	// Set optional fields only if they are provided
	if updateDto.Type != 0 {
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

	if updateDto.MaxScore != 0 {
		maxScore := int32(updateDto.MaxScore)
		request.MaxScore = &maxScore
	}

	response, err := paperService.Client().UpdatePaperQuestion(ctx, request)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	responseDto := dtos.UpdateQuestionResponseDto{
		ID: response.QuestionHash,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseDto)
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	vars := mux.Vars(r)

	paperHash, err := getPaperIdFromVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionHash, err := getQuestionIdFromVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	_, err = paperService.Client().DeletePaperQuestion(ctx, &proto.PaperQuestionRequest{
		PaperHash:    paperHash,
		QuestionHash: questionHash,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ReorderQuestions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	vars := mux.Vars(r)

	paperHash, err := getPaperIdFromVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	categoryID, err := getInt64FromVars(vars, "category_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	_, err = paperService.Client().ReorderPaperQuestions(ctx, &proto.ReorderPaperQuestionsRequest{
		PaperHash:      paperHash,
		CategoryId:     categoryID,
		QuestionHashes: reorderDto.Questions,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpsertPaperTestCases(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionHash, err := getQuestionIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var testCasesDto dtos.UpsertTestCasesDto
	if err := json.NewDecoder(r.Body).Decode(&testCasesDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(testCasesDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	testCases := make([]*proto.UpsertTestCase, len(testCasesDto.TestCases))
	for i, tc := range testCasesDto.TestCases {
		testCases[i] = &proto.UpsertTestCase{
			Inputs:      tc.Inputs,
			Output:      tc.Output,
			Explanation: tc.Explanation,
			Hidden:      tc.Hidden,
		}
	}

	_, err = paperService.Client().UpsertPaperTestCases(ctx, &proto.UpsertPaperTestCasesRequest{
		PaperHash:    paperHash,
		QuestionHash: questionHash,
		TestCases:    testCases,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
