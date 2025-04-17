package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
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

	response := make([]dtos.ExamResponse, len(examList.Exams))
	for i, exam := range examList.Exams {
		response[i] = dtos.ExamResponse{
			ID:                 exam.Id,
			Title:              exam.Title,
			StartsAt:           exam.StartsAt.AsTime(),
			EndsAt:             exam.EndsAt.AsTime(),
			CreatedBy:          exam.CreatedBy,
			Type:               exam.Type,
			MaxCandidatesCount: int(exam.MaxCandidatesCount),
			PaperID:            exam.PaperId,
			DurationMinutes:    exam.DurationMinutes,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateExam(w http.ResponseWriter, r *http.Request) {
	var examDto dtos.CreateExamDto
	if err := json.NewDecoder(r.Body).Decode(&examDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(examDto)
	if errs != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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

	response := dtos.ExamResponse{
		ID:                 exam.Id,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: int(exam.MaxCandidatesCount),
		PaperID:            exam.PaperId,
		DurationMinutes:    exam.DurationMinutes,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func UpdateExam(w http.ResponseWriter, r *http.Request) {
	var examDto dtos.UpdateExamDto
	if err := json.NewDecoder(r.Body).Decode(&examDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(examDto)
	if errs != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	examID, _ := strconv.Atoi(vars["examId"])
	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	req := &proto.UpdateExamRequest{
		ExamId: int64(examID),
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

	response := dtos.ExamResponse{
		ID:                 exam.Id,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: int(exam.MaxCandidatesCount),
		PaperID:            exam.PaperId,
		DurationMinutes:    exam.DurationMinutes,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetExamParticipants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.Atoi(vars["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	participants, err := examService.Client().GetExamParticipants(ctx, &proto.ExamRequest{
		ExamId: int64(examID),
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	authService := services.GetAuthService()
	authCtx := context.Background()

	response := make([]dtos.ExamParticipantResponse, len(participants.Participants))
	for i, p := range participants.Participants {
		// Get user details from auth service
		userResp, err := authService.Client().GetUser(authCtx, &proto.GetUserRequest{
			UserId: p.UserId,
		})

		response[i] = dtos.ExamParticipantResponse{
			ID:           p.Id,
			UserID:       p.UserId,
			Status:       int(p.Status),
			ScoreAwarded: int(p.ScoreAwarded),
		}

		if err == nil {
			response[i].FirstName = userResp.FirstName
			response[i].LastName = userResp.LastName
			response[i].Email = userResp.Email
		}

		if p.StartedAt != nil {
			response[i].StartedAt = p.StartedAt.AsTime()
		}
		if p.EndedAt != nil {
			response[i].EndedAt = p.EndedAt.AsTime()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AddExamParticipant(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	examID, err := strconv.Atoi(params["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	var participantDto dtos.AddExamParticipantDto
	if err := json.NewDecoder(r.Body).Decode(&participantDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(participantDto)
	if errs != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	// First create/update user if email is provided
	if participantDto.Email != "" {
		authService := services.GetAuthService()
		authCtx := context.Background()

		userResp, err := authService.Client().UpsertUser(authCtx, &proto.UpsertUserRequest{
			Email:     participantDto.Email,
			FirstName: &participantDto.FirstName,
			LastName:  &participantDto.LastName,
		})
		if err != nil {
			http.Error(w, "Failed to create/update user", http.StatusInternalServerError)
			return
		}
		participantDto.UserID = userResp.Id
	}

	// Add participant
	participant, err := examService.Client().AddExamParticipant(ctx, &proto.AddParticipantRequest{
		ExamId: int64(examID),
		UserId: participantDto.UserID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.AddExamParticipantResponse{
		ID:           participant.Id,
		UserID:       participant.UserId,
		Status:       int(participant.Status),
		ScoreAwarded: int(participant.ScoreAwarded),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func RemoveExamParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, _ := strconv.Atoi(vars["examId"])
	participantID, _ := strconv.Atoi(vars["participantId"])

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err := examService.Client().RemoveExamParticipant(ctx, &proto.RemoveParticipantRequest{
		ExamId:        int64(examID),
		ParticipantId: int64(participantID),
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper function to create a new exam participant
func createExamParticipant(tx *gorm.DB, examID int64, userID int64) (*models.ExamParticipant, error) {
	participant := models.ExamParticipant{
		ExamID: examID,
		UserID: userID,
	}

	if err := tx.Create(&participant).Error; err != nil {
		return nil, err
	}
	return &participant, nil
}

func StartExam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.Atoi(vars["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err = examService.Client().StartExam(ctx, &proto.StartExamRequest{
		ExamId: int64(examID),
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func EndExam(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examID, _ := strconv.Atoi(mux.Vars(r)["examId"])

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err := examService.Client().EndExam(ctx, &proto.EndExamRequest{
		ExamId: int64(examID),
	})
	if err != nil {
		http.Error(w, "Failed to end exam", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetExam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.Atoi(vars["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	exam, err := examService.Client().GetExam(ctx, &proto.ExamRequest{
		ExamId: int64(examID),
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.ExamResponse{
		ID:                 exam.Id,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: int(exam.MaxCandidatesCount),
		PaperID:            exam.PaperId,
		DurationMinutes:    exam.DurationMinutes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CheckExamAccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.Atoi(vars["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	access, err := examService.Client().CheckExamAccess(ctx, &proto.ExamRequest{
		ExamId: int64(examID),
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.ExamAccessResponse{
		AccessType: access.AccessType.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CheckExamParticipant checks if the current user is a participant of the exam
func CheckExamParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.Atoi(vars["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err = examService.Client().CheckExamParticipant(ctx, &proto.CheckParticipantRequest{
		ExamId: int64(examID),
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
