package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"pariksha/common/pkg/proto"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"

	"github.com/gorilla/mux"
)

func GetExamQuestions(w http.ResponseWriter, r *http.Request) {
	examID, err := getInt64FromVars(mux.Vars(r), "examId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	// Get question IDs from exam service
	questions, err := examService.Client().GetExamQuestions(ctx, &proto.ExamRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := make([]dtos.ExamQuestionMinimalResponseDto, len(questions.Questions))
	for i, q := range questions.Questions {
		response[i] = dtos.ExamQuestionMinimalResponseDto{
			QuestionID: q.QuestionId,
			CategoryID: q.CategoryId,
			Type:       q.Type,
			Order:      q.Order,
			MaxScore:   q.MaxScore,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetExamCategories(w http.ResponseWriter, r *http.Request) {
	examID, err := getInt64FromVars(mux.Vars(r), "examId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	// Get category IDs from exam service
	categories, err := examService.Client().GetExamCategories(ctx, &proto.ExamRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Extract category IDs
	categoryIDs := make([]int64, len(categories.Categories))
	for i, c := range categories.Categories {
		categoryIDs[i] = c.CategoryId
	}

	// No categories to fetch
	if len(categoryIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]dtos.ExamCategoriesResponseDto{})
		return
	}

	// Get category data from paper service
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	categoryData, err := paperService.Client().GetCategoriesByIds(paperCtx, &proto.GetCategoriesByIdsRequest{
		CategoryIds: categoryIDs,
	})
	if err != nil {
		http.Error(w, "Failed to retrieve category data", http.StatusInternalServerError)
		return
	}

	// Create a map of category data
	categoryMap := make(map[int64]*proto.CategoryBatchItem)
	for _, c := range categoryData.Categories {
		categoryMap[c.Id] = c
	}

	// Convert to response format using order from exam categories
	response := make([]dtos.ExamCategoriesResponseDto, len(categories.Categories))
	for i, c := range categories.Categories {
		response[i] = dtos.ExamCategoriesResponseDto{
			CategoryID: c.CategoryId,
			Name:       categoryMap[c.CategoryId].Name,
			Order:      c.Order,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetExamQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := getInt64FromVars(mux.Vars(r), "questionId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	question, err := paperService.Client().GetExamQuestion(context.Background(), &proto.QuestionRequest{
		QuestionId: questionID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	var questionContent json.RawMessage
	switch q := question.Question.(type) {
	case *proto.QuestionResponse_Mcq:
		questionContent, _ = json.Marshal(q.Mcq)
	case *proto.QuestionResponse_Subjective:
		questionContent, _ = json.Marshal(q.Subjective)
	}

	response := dtos.ExamQuestionResponseDto{
		ID:       question.Id,
		Question: questionContent,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
