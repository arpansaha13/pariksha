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

func GetPaperCategories(w http.ResponseWriter, r *http.Request) {
	paperID, err := getInt64FromVars(mux.Vars(r), "paperId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetPaperCategories(ctx, &proto.PaperRequest{
		PaperId: paperID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	categories := make([]dtos.QuestionCategoryResponse, len(response.Categories))
	for i, c := range response.Categories {
		categories[i] = dtos.QuestionCategoryResponse{
			ID:    c.Id,
			Name:  c.Name,
			Order: int(c.Order),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	paperID, err := getInt64FromVars(mux.Vars(r), "paperId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().CreateCategory(ctx, &proto.CreateCategoryRequest{
		PaperId: paperID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dtos.QuestionCategoryResponse{
		ID:    response.Id,
		Name:  response.Name,
		Order: int(response.Order),
	})
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := getInt64FromVars(mux.Vars(r), "categoryId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	var categoryDto dtos.UpdateCategoryDto
	if err := json.NewDecoder(r.Body).Decode(&categoryDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(categoryDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	_, err = paperService.Client().UpdateCategory(ctx, &proto.UpdateCategoryRequest{
		CategoryId: categoryID,
		Name:       categoryDto.Name,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := getInt64FromVars(mux.Vars(r), "categoryId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	_, err = paperService.Client().DeleteCategory(ctx, &proto.CategoryRequest{
		CategoryId: categoryID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ReorderCategories(w http.ResponseWriter, r *http.Request) {
	paperID, err := getInt64FromVars(mux.Vars(r), "paperId")
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	var reorderDto dtos.ReorderCategoryDto
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

	categoryIDs := make([]int64, len(reorderDto.Categories))
	for i, id := range reorderDto.Categories {
		categoryIDs[i] = id
	}

	_, err = paperService.Client().ReorderCategories(ctx, &proto.ReorderCategoriesRequest{
		PaperId:     paperID,
		CategoryIds: categoryIDs,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
