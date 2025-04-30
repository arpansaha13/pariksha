package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetUserPapers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	response, err := paperService.Client().GetUserPapers(ctx, &proto.Empty{})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.Papers)
}

func CreatePaper(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	response, err := paperService.Client().CreatePaper(ctx, &proto.Empty{})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func UpdatePaper(w http.ResponseWriter, r *http.Request) {
	var paperDto dtos.UpdatePaperDto
	if err := json.NewDecoder(r.Body).Decode(&paperDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(paperDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	paperID, err := getInt64FromVars(vars, "paperId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	paperService := services.GetPaperService()

	updatePaperRequest := &proto.UpdatePaperRequest{
		PaperId: paperID,
	}

	if paperDto.Title != "" {
		updatePaperRequest.Title = &paperDto.Title
	}

	if paperDto.DurationMinutes > 0 {
		updatePaperRequest.DurationMinutes = ptr.Int32(paperDto.DurationMinutes)
	}

	ctx := paperService.CreateMetadata(userID)
	_, err = paperService.Client().UpdatePaper(ctx, updatePaperRequest)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetPaper(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID, err := getInt64FromVars(vars, "paperId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	response, err := paperService.Client().GetPaper(ctx, &proto.PaperRequest{
		PaperId: paperID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CheckPaperAccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID, err := getInt64FromVars(vars, "paperId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	_, err = paperService.Client().CheckPaperAccess(ctx, &proto.PaperRequest{
		PaperId: paperID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
