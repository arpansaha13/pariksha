package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetPaperQuestions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetPaperQuestions(ctx, &proto.PaperRequest{
		PaperId: int32(paperID),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Convert proto response to HTTP response
	httpResponse := make([]dtos.QuestionMinimalResponse, len(response.Questions))
	for i, q := range response.Questions {
		var questionData json.RawMessage
		switch q := q.Question.(type) {
		case *proto.QuestionMinimal_Mcq:
			data, _ := json.Marshal(q.Mcq)
			questionData = data
		case *proto.QuestionMinimal_General:
			data, _ := json.Marshal(q.General)
			questionData = data
		}

		httpResponse[i] = dtos.QuestionMinimalResponse{
			ID:         int(q.Id),
			CategoryID: int(q.CategoryId),
			PaperID:    int(q.PaperId),
			Order:      int(q.Order),
			Question:   questionData,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResponse)
}

func GetQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetQuestion(ctx, &proto.QuestionRequest{
		QuestionId: int32(questionID),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Convert proto response to HTTP response
	var questionData json.RawMessage
	switch q := response.Question.(type) {
	case *proto.QuestionResponse_Mcq:
		data, _ := json.Marshal(q.Mcq)
		questionData = data
	case *proto.QuestionResponse_General:
		data, _ := json.Marshal(q.General)
		questionData = data
	}

	tags, _ := json.Marshal(response.Tags)

	httpResponse := dtos.QuestionResponse{
		ID:       int(response.Id),
		Question: questionData,
		Category: &dtos.QuestionCategoryResponse{
			ID:    int(response.Category.Id),
			Name:  response.Category.Name,
			Order: int(response.Category.Order),
		},
		Type:          response.Type,
		Tags:          tags,
		PaperID:       int(response.PaperId),
		MaxScore:      int(response.MaxScore),
		CorrectAnswer: *response.CorrectAnswer,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(httpResponse)
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	var questionDto dtos.CreateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&questionDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(questionDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	var tags []string
	if err := json.Unmarshal(questionDto.Tags, &tags); err != nil {
		http.Error(w, "Invalid tags data", http.StatusBadRequest)
		return
	}

	requestObj := proto.CreateQuestionRequest{
		PaperId:       int32(paperID),
		Question:      nil,
		CategoryId:    int32(questionDto.CategoryID),
		Type:          questionDto.Type,
		Tags:          tags,
		MaxScore:      int32(questionDto.MaxScore),
		CorrectAnswer: &questionDto.CorrectAnswer,
	}

	switch questionDto.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq structs.MCQQuestion
		if err := json.Unmarshal(questionDto.Question, &mcq); err != nil {
			http.Error(w, "Invalid question data", http.StatusBadRequest)
			return
		}
		requestObj.Question = &proto.CreateQuestionRequest_Mcq{
			Mcq: &proto.McqQuestion{
				Statement: mcq.Statement,
				Options:   mcq.Options,
			},
		}
	default:
		var general structs.GeneralQuestion
		if err := json.Unmarshal(questionDto.Question, &general); err != nil {
			http.Error(w, "Invalid question data", http.StatusBadRequest)
			return
		}
		requestObj.Question = &proto.CreateQuestionRequest_General{
			General: &proto.GeneralQuestion{
				Statement: general.Statement,
			},
		}
	}

	response, err := paperService.Client().CreateQuestion(ctx, &requestObj)

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(protoQuestionToResponse(response))
}

func UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	var updateDto dtos.UpdateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&updateDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	request := &proto.UpdateQuestionRequest{
		QuestionId: int32(questionID),
	}

	// Set optional fields only if they are provided
	if updateDto.Type != "" {
		request.Type = &updateDto.Type
	}

	if updateDto.Question != nil {
		switch updateDto.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq structs.MCQQuestion
			if err := json.Unmarshal(updateDto.Question, &mcq); err != nil {
				http.Error(w, "Invalid question data", http.StatusBadRequest)
				return
			}
			request.Question = &proto.UpdateQuestionRequest_Mcq{
				Mcq: &proto.McqQuestion{
					Statement: mcq.Statement,
					Options:   mcq.Options,
				},
			}
		default:
			var general structs.GeneralQuestion
			if err := json.Unmarshal(updateDto.Question, &general); err != nil {
				http.Error(w, "Invalid question data", http.StatusBadRequest)
				return
			}
			request.Question = &proto.UpdateQuestionRequest_General{
				General: &proto.GeneralQuestion{
					Statement: general.Statement,
				},
			}
		}
	}

	if updateDto.CategoryID != 0 {
		categoryId := int32(updateDto.CategoryID)
		request.CategoryId = &categoryId
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

	_, err := paperService.Client().UpdateQuestion(ctx, request)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	_, err := paperService.Client().DeleteQuestion(ctx, &proto.QuestionRequest{
		QuestionId: int32(questionID),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ReorderQuestions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, _ := strconv.Atoi(vars["category_id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	var reorderDto dtos.ReorderQuestionsDto
	if err := json.NewDecoder(r.Body).Decode(&reorderDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(reorderDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	questionIDs := make([]int32, len(reorderDto.Questions))
	for i, id := range reorderDto.Questions {
		questionIDs[i] = int32(id)
	}

	_, err := paperService.Client().ReorderQuestions(ctx, &proto.ReorderQuestionsRequest{
		CategoryId:  int32(categoryID),
		QuestionIds: questionIDs,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func protoQuestionToResponse(resp *proto.QuestionResponse) dtos.QuestionResponse {
	tags, _ := json.Marshal(resp.Tags)

	response := dtos.QuestionResponse{
		ID:            int(resp.Id),
		Type:          resp.Type,
		Tags:          tags,
		PaperID:       int(resp.PaperId),
		MaxScore:      int(resp.MaxScore),
		CorrectAnswer: resp.GetCorrectAnswer(),
	}

	switch q := resp.Question.(type) {
	case *proto.QuestionResponse_Mcq:
		data, _ := json.Marshal(q.Mcq)
		response.Question = data
	case *proto.QuestionResponse_General:
		data, _ := json.Marshal(q.General)
		response.Question = data
	}

	if resp.Category != nil {
		response.Category = &dtos.QuestionCategoryResponse{
			ID:    int(resp.Category.Id),
			Name:  resp.Category.Name,
			Order: int(resp.Category.Order),
		}
	}

	return response
}
