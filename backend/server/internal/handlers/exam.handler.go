package handlers

import (
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
	userID := r.Context().Value(middlewares.UserIDKey).(int)

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
			ID:                 int(exam.Id),
			Title:              exam.Title,
			StartsAt:           exam.StartsAt.AsTime(),
			EndsAt:             exam.EndsAt.AsTime(),
			CreatedBy:          int(exam.CreatedBy),
			Type:               exam.Type,
			MaxCandidatesCount: int(exam.MaxCandidatesCount),
			PaperID:            int(exam.PaperId),
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

	userID := r.Context().Value(middlewares.UserIDKey).(int)

	// Verify paper exists first
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	_, err := paperService.Client().GetPaper(paperCtx, &proto.PaperRequest{
		PaperId: int32(examDto.PaperID),
	})
	if err != nil {
		http.Error(w, "Paper not found", http.StatusNotFound)
		return
	}

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	exam, err := examService.Client().CreateExam(ctx, &proto.CreateExamRequest{
		Title:              examDto.Title,
		StartsAt:           timestamppb.New(examDto.StartsAt),
		EndsAt:             timestamppb.New(examDto.EndsAt),
		MaxCandidatesCount: int32(examDto.MaxCandidatesCount),
		Type:               examDto.Type,
		PaperId:            int32(examDto.PaperID),
	})
	if err != nil {
		http.Error(w, "Failed to create exam", http.StatusInternalServerError)
		return
	}

	response := dtos.ExamResponse{
		ID:                 int(exam.Id),
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          int(exam.CreatedBy),
		Type:               exam.Type,
		MaxCandidatesCount: int(exam.MaxCandidatesCount),
		PaperID:            int(exam.PaperId),
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
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	req := &proto.UpdateExamRequest{
		ExamId: int32(examID),
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
	if examDto.MaxCandidatesCount != 0 {
		maxCount := int32(examDto.MaxCandidatesCount)
		req.MaxCandidatesCount = &maxCount
	}

	exam, err := examService.Client().UpdateExam(ctx, req)
	if err != nil {
		http.Error(w, "Failed to update exam", http.StatusInternalServerError)
		return
	}

	response := dtos.ExamResponse{
		ID:                 int(exam.Id),
		Title:              exam.Title,
		StartsAt:           exam.StartsAt.AsTime(),
		EndsAt:             exam.EndsAt.AsTime(),
		CreatedBy:          int(exam.CreatedBy),
		Type:               exam.Type,
		MaxCandidatesCount: int(exam.MaxCandidatesCount),
		PaperID:            int(exam.PaperId),
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

	userID := r.Context().Value(middlewares.UserIDKey).(int)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	participants, err := examService.Client().GetExamParticipants(ctx, &proto.ExamRequest{
		ExamId: int32(examID),
	})
	if err != nil {
		http.Error(w, "Failed to fetch participants", http.StatusInternalServerError)
		return
	}

	response := make([]dtos.ExamParticipantResponse, len(participants.Participants))
	for i, p := range participants.Participants {
		response[i] = dtos.ExamParticipantResponse{
			ID:           int(p.Id),
			UserID:       int(p.UserId),
			FirstName:    p.FirstName,
			LastName:     p.LastName,
			Email:        p.Email,
			Status:       int(p.Status),
			ScoreAwarded: int(p.ScoreAwarded),
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

func AddExamParticipants(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	examID, err := strconv.Atoi(params["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	var participantsDto []dtos.AddExamParticipantDto
	if err := json.NewDecoder(r.Body).Decode(&participantsDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	participants := make([]*proto.AddParticipant, len(participantsDto))
	for i, p := range participantsDto {
		participants[i] = &proto.AddParticipant{
			UserId:    int32(p.UserID),
			Email:     p.Email,
			FirstName: p.FirstName,
			LastName:  p.LastName,
		}
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	response, err := examService.Client().AddExamParticipants(ctx, &proto.AddParticipantsRequest{
		ExamId:       int32(examID),
		Participants: participants,
	})
	if err != nil {
		http.Error(w, "Failed to add participants", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dtos.AddExamParticipantResponse{
		AddedCount:     int(response.AddedCount),
		OmittedCount:   int(response.OmittedCount),
		MaxLimitReason: response.GetMaxLimitReason(),
	})
}

func RemoveExamParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, _ := strconv.Atoi(vars["examId"])
	participantID, _ := strconv.Atoi(vars["participantId"])

	userID := r.Context().Value(middlewares.UserIDKey).(int)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err := examService.Client().RemoveExamParticipant(ctx, &proto.RemoveParticipantRequest{
		ExamId:        int32(examID),
		ParticipantId: int32(participantID),
	})
	if err != nil {
		http.Error(w, "Failed to remove participant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper function to create a new exam participant
func createExamParticipant(tx *gorm.DB, examID, userID int) (*models.ExamParticipant, error) {
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

	userID := r.Context().Value(middlewares.UserIDKey).(int)

	// Get paper details first
	paperService := services.GetPaperService()
	paperCtx := paperService.CreateMetadata(userID)

	paper, err := paperService.Client().GetPaper(paperCtx, &proto.PaperRequest{
		PaperId: int32(examID),
	})
	if err != nil {
		http.Error(w, "Failed to get paper details", http.StatusInternalServerError)
		return
	}

	// Start exam
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err = examService.Client().StartExam(ctx, &proto.StartExamRequest{
		ExamId:          int32(examID),
		DurationMinutes: paper.DurationMinutes,
	})
	if err != nil {
		http.Error(w, "Failed to start exam", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func EndExam(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int)
	examID, _ := strconv.Atoi(mux.Vars(r)["examId"])

	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err := examService.Client().EndExam(ctx, &proto.EndExamRequest{
		ExamId: int32(examID),
	})
	if err != nil {
		http.Error(w, "Failed to end exam", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
