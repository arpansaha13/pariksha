package handlers

import (
	"encoding/json"
	"net/http"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
	"pariksha/server/internal/utils"
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

	papers := make([]dtos.PaperResponseDto, len(response.Papers))
	for i, p := range response.Papers {
		encryptedID, err := utils.EncryptID(p.Id)
		if err != nil {
			http.Error(w, "Failed to process paper IDs", http.StatusInternalServerError)
			return
		}
		papers[i] = dtos.PaperResponseDto{
			ID:              encryptedID,
			Title:           p.Title,
			MaxScore:        p.MaxScore,
			DurationMinutes: p.DurationMinutes,
			QuestionCounts: dtos.QuestionCountDto{
				MCQ:        p.QuestionCounts.Mcq,
				Subjective: p.QuestionCounts.Subjective,
				Coding:     p.QuestionCounts.Coding,
			},
			CreatedBy: p.CreatedBy,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(papers)
}

func GetPaper(w http.ResponseWriter, r *http.Request) {
	paperID := r.Context().Value(middlewares.DecryptedPaperID).(int64)
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

	encryptedID, err := utils.EncryptID(response.Id)
	if err != nil {
		http.Error(w, "Failed to process paper ID", http.StatusInternalServerError)
		return
	}

	paperDto := dtos.PaperResponseDto{
		ID:              encryptedID,
		Title:           response.Title,
		MaxScore:        response.MaxScore,
		DurationMinutes: response.DurationMinutes,
		QuestionCounts: dtos.QuestionCountDto{
			MCQ:        response.QuestionCounts.Mcq,
			Subjective: response.QuestionCounts.Subjective,
			Coding:     response.QuestionCounts.Coding,
		},
		CreatedBy: response.CreatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paperDto)
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

	encryptedID, err := utils.EncryptID(response.Id)
	json.NewEncoder(w).Encode(dtos.CreatePaperResponseDto{
		ID:        encryptedID,
		Title:     response.Title,
		CreatedBy: response.CreatedBy,
	})
}

func UpdatePaper(w http.ResponseWriter, r *http.Request) {
	var paperDto dtos.UpdatePaperDto
	if err := json.NewDecoder(r.Body).Decode(&paperDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(paperDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	paperID := r.Context().Value(middlewares.DecryptedPaperID).(int64)
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
	_, err := paperService.Client().UpdatePaper(ctx, updatePaperRequest)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetPaperPermissions(w http.ResponseWriter, r *http.Request) {
	paperID := r.Context().Value(middlewares.DecryptedPaperID).(int64)
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	paperService := services.GetPaperService()

	ctx := paperService.CreateMetadata(userID)
	response, err := paperService.Client().GetPaperPermissions(ctx, &proto.PaperRequest{
		PaperId: paperID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	permissions := dtos.PaperPermissionsDto{
		CanRead:  response.CanRead,
		CanWrite: response.CanWrite,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

func DeletePapers(w http.ResponseWriter, r *http.Request) {
	var deletePaperDto dtos.DeletePaperDto
	if err := json.NewDecoder(r.Body).Decode(&deletePaperDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(deletePaperDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	// Decrypt paper IDs
	decryptedPaperIds := make([]int64, len(deletePaperDto.PaperIDs))
	for i, encryptedID := range deletePaperDto.PaperIDs {
		decryptedID, err := utils.DecryptID(encryptedID)
		if err != nil {
			http.Error(w, "Invalid paper ID", http.StatusBadRequest)
			return
		}
		decryptedPaperIds[i] = decryptedID
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	_, err := paperService.Client().DeletePapers(ctx, &proto.DeletePapersRequest{
		PaperIds: decryptedPaperIds,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
