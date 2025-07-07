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
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().GetPaperCategories(ctx, &proto.PaperRequest{
		PaperHash: paperHash,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	categories := make([]dtos.QuestionCategoryResponseDto, len(response.Categories))
	for i, c := range response.Categories {
		categories[i] = dtos.QuestionCategoryResponseDto{
			ID:    c.Id,
			Name:  c.Name,
			Order: c.Order,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	response, err := paperService.Client().CreatePaperCategory(ctx, &proto.CreatePaperCategoryRequest{
		PaperHash: paperHash,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dtos.QuestionCategoryResponseDto{
		ID:    response.Id,
		Name:  response.Name,
		Order: response.Order,
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
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(categoryDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	_, err = paperService.Client().UpdatePaperCategory(ctx, &proto.UpdatePaperCategoryRequest{
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

	_, err = paperService.Client().DeletePaperCategory(ctx, &proto.PaperCategoryRequest{
		CategoryId: categoryID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ReorderCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	paperHash, err := getPaperIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var reorderDto dtos.ReorderCategoryDto
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

	categoryIDs := make([]int64, len(reorderDto.Categories))
	for i, id := range reorderDto.Categories {
		categoryIDs[i] = id
	}

	_, err = paperService.Client().ReorderPaperCategories(ctx, &proto.ReorderPaperCategoriesRequest{
		PaperHash:   paperHash,
		CategoryIds: categoryIDs,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
