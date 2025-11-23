package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/dtos"
	"pariksha/gateway/internal/middlewares"
	"pariksha/gateway/internal/services"
)

func GetExamQuestions(w http.ResponseWriter, r *http.Request) {
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	// Get question IDs from exam service
	questions, err := examService.Client().GetExamQuestions(ctx, &proto.ExamRequest{
		ExamHash: examHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := make([]dtos.ExamQuestionMinimalResponseDto, len(questions.Questions))
	for i, q := range questions.Questions {
		response[i] = dtos.ExamQuestionMinimalResponseDto{
			QuestionID: q.QuestionHash,
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
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	// Get category IDs from exam service
	categories, err := examService.Client().GetExamCategories(ctx, &proto.ExamRequest{
		ExamHash: examHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Convert to response format using order from exam categories
	response := make([]dtos.ExamCategoriesResponseDto, len(categories.Categories))
	for i, c := range categories.Categories {
		response[i] = dtos.ExamCategoriesResponseDto{
			CategoryID: c.CategoryId,
			Name:       c.Name,
			Order:      c.Order,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetExamQuestion(w http.ResponseWriter, r *http.Request) {
	questionHash, err := getQuestionIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	examService := services.GetExamService()
	question, err := examService.Client().GetExamQuestion(context.Background(), &proto.ExamQuestionRequest{
		QuestionHash: questionHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.ExamQuestionResponseDto{
		ID:       question.QuestionHash,
		Type:     question.Type,
		Question: question.RawQuestion,
	}

	if len(question.TestCases) > 0 {
		testCases := make([]dtos.PaperTestCaseDto, len(question.TestCases))
		for i, tc := range question.TestCases {
			testCases[i] = dtos.PaperTestCaseDto{
				Inputs:      tc.Inputs,
				Output:      tc.Output,
				Order:       tc.Order,
				Hidden:      tc.Hidden,
				Explanation: tc.Explanation,
			}
		}
		response.TestCases = &testCases
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
