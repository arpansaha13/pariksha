package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetUserPapers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int)
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
	userID := r.Context().Value(middlewares.UserIDKey).(int)
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
	paperID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	_, err := paperService.Client().UpdatePaper(ctx, &proto.UpdatePaperRequest{
		PaperId: int32(paperID),
		Title:   paperDto.Title,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetPaper(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	response, err := paperService.Client().GetPaper(ctx, &proto.PaperRequest{
		PaperId: int32(paperID),
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
	paperID, _ := strconv.Atoi(vars["id"])
	userID := r.Context().Value(middlewares.UserIDKey).(int)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	_, err := paperService.Client().CheckPaperAccess(ctx, &proto.PaperRequest{
		PaperId: int32(paperID),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
