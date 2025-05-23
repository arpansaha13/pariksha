package handlers

import (
	"encoding/json"
	"net/http"

	"google.golang.org/protobuf/types/known/timestamppb"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
	"pariksha/server/internal/utils"
)

func GetUserExams(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	examList, err := examService.Client().GetUserExams(ctx, &proto.Empty{})
	if err != nil {
		http.Error(w, "Failed to retrieve exams", http.StatusInternalServerError)
		return
	}

	response := make([]dtos.ExamResponseDto, len(examList.Exams))
	for i, exam := range examList.Exams {
		encryptedExamID, _ := utils.EncryptID(exam.Id)
		response[i] = dtos.ExamResponseDto{
			ID:                 encryptedExamID,
			Title:              exam.Title,
			StartsAt:           exam.StartsAt.AsTime(),
			EndsAt:             exam.EndsAt.AsTime(),
			CreatedBy:          exam.CreatedBy,
			Type:               exam.Type,
			MaxCandidatesCount: exam.MaxCandidatesCount,
			PaperID:            exam.PaperId,
			DurationMinutes:    exam.DurationMinutes,
			MaxScore:           exam.MaxScore,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateExam(w http.ResponseWriter, r *http.Request) {
	var examDto dtos.CreateExamDto
	if err := json.NewDecoder(r.Body).Decode(&examDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(examDto)
	if errs != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	// Verify paper exists first
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	_, err := paperService.Client().GetPaper(paperCtx, &proto.PaperRequest{
		PaperId: examDto.PaperID,
	})
	if err != nil {
		http.Error(w, "Paper not found", http.StatusNotFound)
		return
	}

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	createExamRequest := &proto.CreateExamRequest{
		Title:              examDto.Title,
		StartsAt:           timestamppb.New(examDto.StartsAt),
		EndsAt:             timestamppb.New(examDto.EndsAt),
		MaxCandidatesCount: 10,
		Type:               nil,
		PaperId:            examDto.PaperID,
		DurationMinutes:    examDto.DurationMinutes,
	}

	if examDto.Type != "" {
		createExamRequest.Type = &examDto.Type
	}

	exam, err := examService.Client().CreateExam(ctx, createExamRequest)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	encryptedExamID, _ := utils.EncryptID(exam.Id)
	response := dtos.ExamResponseDto{
		ID:                 encryptedExamID,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		PaperID:            exam.PaperId,
		DurationMinutes:    exam.DurationMinutes,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func UpdateExam(w http.ResponseWriter, r *http.Request) {
	var examDto dtos.UpdateExamDto
	if err := json.NewDecoder(r.Body).Decode(&examDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(examDto)
	if errs != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	examID := r.Context().Value(middlewares.DecryptedExamID).(int64)
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	req := &proto.UpdateExamRequest{
		ExamId: examID,
	}

	if examDto.Title != "" {
		req.Title = &examDto.Title
	}
	if !examDto.StartsAt.IsZero() {
		req.StartsAt = timestamppb.New(examDto.StartsAt)
	}
	if !examDto.EndsAt.IsZero() {
		req.EndsAt = timestamppb.New(examDto.EndsAt)
	}
	if examDto.Type != "" {
		req.Type = &examDto.Type
	}
	if examDto.DurationMinutes != nil {
		req.DurationMinutes = examDto.DurationMinutes
	}

	exam, err := examService.Client().UpdateExam(ctx, req)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	encryptedExamID, _ := utils.EncryptID(exam.Id)
	response := dtos.ExamResponseDto{
		ID:                 encryptedExamID,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		PaperID:            exam.PaperId,
		DurationMinutes:    exam.DurationMinutes,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func StartExam(w http.ResponseWriter, r *http.Request) {
	examID := r.Context().Value(middlewares.DecryptedExamID).(int64)
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err := examService.Client().StartExam(ctx, &proto.StartExamRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func EndExam(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examID := r.Context().Value(middlewares.DecryptedExamID).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err := examService.Client().EndExam(ctx, &proto.EndExamRequest{
		ExamId: examID,
	})
	if err != nil {
		http.Error(w, "Failed to end exam", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetExam(w http.ResponseWriter, r *http.Request) {
	examID := r.Context().Value(middlewares.DecryptedExamID).(int64)
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	exam, err := examService.Client().GetExam(ctx, &proto.ExamRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	encryptedExamID, _ := utils.EncryptID(exam.Id)
	response := dtos.ExamResponseDto{
		ID:                 encryptedExamID,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		PaperID:            exam.PaperId,
		DurationMinutes:    exam.DurationMinutes,
		MaxScore:           exam.MaxScore,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetExamPermission(w http.ResponseWriter, r *http.Request) {
	examID := r.Context().Value(middlewares.DecryptedExamID).(int64)
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	permission, err := examService.Client().GetExamPermission(ctx, &proto.ExamRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.ExamPermissionResponseDto{
		CanRead:        permission.CanRead,
		CanWrite:       permission.CanWrite,
		CanParticipate: permission.CanParticipate,
		CanEvaluate:    permission.CanEvaluate,
	}

	if permission.ParticipantStatus != nil {
		status := int(permission.GetParticipantStatus())
		response.ParticipantStatus = &status
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func DeleteExams(w http.ResponseWriter, r *http.Request) {
	var deleteExamsDto dtos.DeleteExamsDto
	if err := json.NewDecoder(r.Body).Decode(&deleteExamsDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(deleteExamsDto)
	if errs != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	// Decrypt exam IDs
	decryptedExamIds := make([]int64, len(deleteExamsDto.ExamIds))
	for i, encryptedID := range deleteExamsDto.ExamIds {
		decryptedID, err := utils.DecryptID(encryptedID)
		if err != nil {
			http.Error(w, "Invalid exam ID", http.StatusBadRequest)
			return
		}
		decryptedExamIds[i] = decryptedID
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err := examService.Client().DeleteExams(ctx, &proto.DeleteExamsRequest{
		ExamIds: decryptedExamIds,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
