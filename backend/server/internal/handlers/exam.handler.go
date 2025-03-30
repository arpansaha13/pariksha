package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/db"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetUserExams(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	var exams []models.Exam
	if err := db.DB.Where("created_by = ?", userID).Find(&exams).Error; err != nil {
		http.Error(w, "Failed to retrieve exams", http.StatusInternalServerError)
		return
	}

	var response []dtos.ExamResponse
	for _, exam := range exams {
		response = append(response, dtos.ExamResponse{
			ID:                 exam.ID,
			Title:              exam.Title,
			StartsAt:           exam.StartsAt,
			EndsAt:             exam.EndsAt,
			CreatedBy:          exam.CreatedBy,
			Type:               exam.Type,
			MaxCandidatesCount: exam.MaxCandidatesCount,
			PaperID:            exam.PaperID,
		})
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
	paperService := services.GetPaperService()
	ctx := paperService.CreateMetadata(userID)

	// Verify paper exists
	_, err := paperService.Client().GetPaper(ctx, &proto.PaperRequest{
		PaperId: int32(examDto.PaperID),
	})
	if err != nil {
		http.Error(w, "Paper not found", http.StatusNotFound)
		return
	}

	exam := models.Exam{
		Title:              examDto.Title,
		StartsAt:           examDto.StartsAt,
		EndsAt:             examDto.EndsAt,
		CreatedBy:          userID,
		Type:               examDto.Type,
		MaxCandidatesCount: examDto.MaxCandidatesCount,
		PaperID:            examDto.PaperID,
	}

	if err := db.DB.Create(&exam).Error; err != nil {
		http.Error(w, "Failed to create exam", http.StatusInternalServerError)
		return
	}

	response := dtos.ExamResponse{
		ID:                 exam.ID,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt,
		EndsAt:             exam.EndsAt,
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		PaperID:            exam.PaperID,
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
	examID := vars["examId"]

	var exam models.Exam
	if err := db.DB.Take(&exam, examID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Exam not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find exam", http.StatusInternalServerError)
		}
		return
	}

	isUpdated := false
	now := time.Now()
	notUpdatedFields := make(map[string]string)

	if examDto.Title != "" && examDto.Title != exam.Title {
		exam.Title = examDto.Title
		isUpdated = true
	}

	if now.After(exam.EndsAt) {
		notUpdatedFields["StartsAt"] = "Cannot update StartsAt after the exam has ended"
		notUpdatedFields["EndsAt"] = "Cannot update EndsAt after the exam has ended"
		notUpdatedFields["Type"] = "Cannot update Type after the exam has ended"
		notUpdatedFields["MaxCandidatesCount"] = "Cannot update MaxCandidatesCount after the exam has ended"
	} else {
		if examDto.StartsAt != (time.Time{}) && examDto.StartsAt != exam.StartsAt {
			if now.After(exam.StartsAt) {
				notUpdatedFields["StartsAt"] = "Cannot update StartsAt after the exam has started"
			} else if examDto.StartsAt.Before(now) {
				http.Error(w, "StartsAt cannot be a time in the past", http.StatusBadRequest)
				return
			} else {
				exam.StartsAt = examDto.StartsAt
				isUpdated = true
			}
		}

		if examDto.EndsAt != (time.Time{}) && examDto.EndsAt != exam.EndsAt {
			if examDto.EndsAt.Before(exam.StartsAt) || examDto.EndsAt.Equal(exam.StartsAt) {
				http.Error(w, "EndsAt cannot be less than or equal to StartsAt", http.StatusBadRequest)
				return
			} else {
				exam.EndsAt = examDto.EndsAt
				isUpdated = true
			}
		}

		if examDto.Type != "" && examDto.Type != exam.Type {
			if now.After(exam.StartsAt) {
				notUpdatedFields["Type"] = "Cannot update Type after the exam has started"
			} else {
				exam.Type = examDto.Type
				isUpdated = true
			}
		}

		if examDto.MaxCandidatesCount != 0 && examDto.MaxCandidatesCount != exam.MaxCandidatesCount {
			exam.MaxCandidatesCount = examDto.MaxCandidatesCount
			isUpdated = true
		}
	}

	if isUpdated {
		if err := db.DB.Save(&exam).Error; err != nil {
			http.Error(w, "Failed to update exam", http.StatusInternalServerError)
			return
		}
	}

	response := dtos.ExamResponse{
		ID:                 exam.ID,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt,
		EndsAt:             exam.EndsAt,
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		PaperID:            exam.PaperID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exam":               response,
		"not_updated_fields": notUpdatedFields,
	})
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
	paperService := services.GetPaperService()

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var exam models.Exam
		if err := tx.Take(&exam, examID).Error; err != nil {
			http.Error(w, "Exam not found", http.StatusNotFound)
			return err
		}

		ctx := paperService.CreateMetadata(userID)
		paper, err := paperService.Client().GetPaper(ctx, &proto.PaperRequest{
			PaperId: int32(exam.PaperID),
		})
		if err != nil {
			http.Error(w, "Failed to get paper details", http.StatusInternalServerError)
			return err
		}

		now := time.Now()

		// Check exam timing constraints
		if exam.StartsAt.After(now) {
			http.Error(w, "Exam has not started yet", http.StatusBadRequest)
			return errors.New("exam has not started")
		}

		if exam.EndsAt.Before(now) {
			http.Error(w, "Exam has ended", http.StatusBadRequest)
			return errors.New("exam has ended")
		}

		var participant models.ExamParticipant
		participantErr := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&participant).Error

		if exam.Type == constants.EXAM_TYPE_OPEN {
			if participantErr == gorm.ErrRecordNotFound {
				// For OPEN exams, create new participant if doesn't exist
				newParticipant, err := createExamParticipant(tx, examID, userID)
				if err != nil {
					http.Error(w, "Failed to create participant", http.StatusInternalServerError)
					return err
				}
				participant = *newParticipant
			} else if participantErr != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return participantErr
			}
		} else {
			// For INVITE exams, participant must exist
			if participantErr != nil {
				http.Error(w, "Participant not found", http.StatusNotFound)
				return participantErr
			}
		}

		// Check if participant has already started
		if participant.Status != constants.PARTICIPANT_STATUS_INVITED {
			http.Error(w, "Participant has already started the exam", http.StatusBadRequest)
			return errors.New("already started")
		}

		counts, err := exam.GetParticipantCounts()
		if err != nil {
			http.Error(w, "Failed to get participant counts", http.StatusInternalServerError)
			return err
		}

		scheduledEndTime := now.Add(time.Duration(paper.DurationMinutes) * time.Minute)

		// Update participant status and times
		participant.Status = constants.PARTICIPANT_STATUS_STARTED
		participant.StartedAt = sql.NullTime{Time: now, Valid: true}
		participant.ScheduledEndTime = sql.NullTime{Time: scheduledEndTime, Valid: true}

		// Update counts based on status change
		counts.Invited--
		counts.Started++

		exam.ParticipantCounts, err = json.Marshal(counts)
		if err != nil {
			http.Error(w, "Failed to marshal counts", http.StatusInternalServerError)
			return err
		}

		if err := tx.Save(&exam).Error; err != nil {
			return err
		}
		if err := tx.Save(&participant).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func EndExam(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int)
	examID := mux.Vars(r)["examId"]

	var examParticipant models.ExamParticipant
	if err := db.DB.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&examParticipant).Error; err != nil {
		http.Error(w, "Exam participant not found", http.StatusNotFound)
		return
	}

	examParticipant.Status = constants.PARTICIPANT_STATUS_ENDED
	examParticipant.EndedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := db.DB.Save(&examParticipant).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
