package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/config/validate"
	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
)

func CreateExam(w http.ResponseWriter, r *http.Request) {
	var examDto dtos.CreateExamDto
	if err := json.NewDecoder(r.Body).Decode(&examDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(examDto)
	if errs != nil {
		http.Error(w, "Invald request body", http.StatusBadRequest)
		return
	}

	var paper models.Paper
	if err := db.DB.Take(&paper, examDto.PaperID).Error; err != nil {
		http.Error(w, "Paper not found", http.StatusNotFound)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int)

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

	var participants []models.ExamParticipant
	if err := db.DB.Preload("User").Where("exam_id = ?", examID).Find(&participants).Error; err != nil {
		http.Error(w, "Failed to fetch participants", http.StatusInternalServerError)
		return
	}

	response := make([]dtos.ExamParticipantResponse, len(participants))
	for i, p := range participants {
		response[i] = dtos.ExamParticipantResponse{
			ID:           p.ID,
			UserID:       p.UserID,
			FirstName:    p.User.FirstName.String,
			LastName:     p.User.LastName.String,
			Email:        p.User.Email,
			Status:       p.Status,
			ScoreAwarded: p.ScoreAwarded,
		}
		if p.StartedAt.Valid {
			response[i].StartedAt = p.StartedAt.Time
		}
		if p.EndedAt.Valid {
			response[i].EndedAt = p.EndedAt.Time
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

	var exam models.Exam
	if err := db.DB.Take(&exam, examID).Error; err != nil {
		http.Error(w, "Failed to find exam", http.StatusNotFound)
		return
	}

	// Get current counts
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		http.Error(w, "Failed to get participant counts", http.StatusInternalServerError)
		return
	}

	var examParticipants []models.ExamParticipant
	addedCount := 0
	omittedCount := 0
	maxLimitReached := false

	for _, participantDto := range participantsDto {
		currTotalParticipants := counts.Invited + counts.Started + counts.Ended
		if currTotalParticipants == exam.MaxCandidatesCount {
			maxLimitReached = true
			omittedCount++
			continue
		}

		var userID int

		if participantDto.UserID != 0 {
			userID = participantDto.UserID
		} else if participantDto.Email != "" {
			// Create a guest user
			username := strings.Split(participantDto.Email, "@")[0]
			user := models.User{
				Email:    participantDto.Email,
				IsGuest:  true,
				Username: username,
			}

			if participantDto.FirstName != "" {
				user.FirstName = sql.NullString{String: participantDto.FirstName, Valid: true}
			}

			if participantDto.LastName != "" {
				user.FirstName = sql.NullString{String: participantDto.LastName, Valid: true}
			}

			if err := db.DB.Create(&user).Error; err != nil {
				http.Error(w, "Failed to create guest user", http.StatusInternalServerError)
				return
			}
			userID = user.ID
		} else {
			// Either user_id or email must be provided
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Add the participant to the exam
		participant := models.ExamParticipant{
			ExamID: examID,
			UserID: userID,
		}

		examParticipants = append(examParticipants, participant)
		counts.Invited++
		addedCount++
	}

	if len(examParticipants) > 0 {
		if err := db.DB.Create(&examParticipants).Error; err != nil {
			http.Error(w, "Failed to add participants", http.StatusInternalServerError)
			return
		}

		exam.ParticipantCounts, err = json.Marshal(counts)
		if err != nil {
			http.Error(w, "Failed to marshal counts", http.StatusInternalServerError)
			return
		}

		// Save exam with updated counts
		if err := db.DB.Save(&exam).Error; err != nil {
			http.Error(w, "Failed to update exam", http.StatusInternalServerError)
			return
		}
	}

	response := dtos.AddExamParticipantResponse{
		AddedCount:   addedCount,
		OmittedCount: omittedCount,
	}

	if maxLimitReached {
		response.MaxLimitReason = "Maximum participant limit reached for the exam"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func RemoveExamParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID := vars["examId"]
	participantID := vars["participantId"]

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var exam models.Exam
		if err := db.DB.Take(&exam, examID).Error; err != nil {
			http.Error(w, "Failed to find exam", http.StatusNotFound)
			return err
		}

		if exam.StartsAt.Before(time.Now()) {
			http.Error(w, "Cannot remove participant after exam has started", http.StatusBadRequest)
			return errors.New("cannot remove participant after exam has started")
		}

		counts, err := exam.GetParticipantCounts()
		if err != nil {
			http.Error(w, "Failed to get participant counts", http.StatusInternalServerError)
			return err
		}

		var participant models.ExamParticipant
		if err := db.DB.Take(&participant, participantID).Error; err != nil {
			http.Error(w, "Participant not found", http.StatusNotFound)
			return err
		}

		// Decrement count based on participant's status
		switch participant.Status {
		case constants.PARTICIPANT_STATUS_INVITED:
			counts.Invited--
		case constants.PARTICIPANT_STATUS_STARTED:
			counts.Started--
		case constants.PARTICIPANT_STATUS_ENDED:
			counts.Ended--
		case constants.PARTICIPANT_STATUS_UNATTENDED:
			counts.Unattended--
		}

		if exam.ParticipantCounts, err = json.Marshal(counts); err != nil {
			http.Error(w, "Failed to marshal counts", http.StatusInternalServerError)
			return err
		}

		if err := tx.Save(&exam).Error; err != nil {
			return err
		}
		if err := tx.Delete(&participant).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to remove participant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func StartExam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.Atoi(vars["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	participantID, err := strconv.Atoi(vars["participantId"])
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var participant models.ExamParticipant
		if err := tx.Take(&participant, participantID).Error; err != nil {
			http.Error(w, "Participant not found", http.StatusNotFound)
			return err
		}

		if participant.Status != constants.PARTICIPANT_STATUS_INVITED {
			http.Error(w, "Participant has already started the exam", http.StatusBadRequest)
			return err
		}

		if participant.ExamID != examID {
			http.Error(w, "Participant does not belong to this exam", http.StatusBadRequest)
			return err
		}

		var exam models.Exam
		if err := tx.Preload("Paper").Take(&exam, examID).Error; err != nil {
			http.Error(w, "Exam not found", http.StatusNotFound)
			return err
		}

		now := time.Now()

		if exam.StartsAt.After(now) {
			http.Error(w, "Exam has not started yet", http.StatusBadRequest)
			return err
		}

		if exam.EndsAt.Before(now) {
			http.Error(w, "Exam has ended", http.StatusBadRequest)
			return err
		}

		counts, err := exam.GetParticipantCounts()
		if err != nil {
			http.Error(w, "Failed to get participant counts", http.StatusInternalServerError)
			return err
		}

		scheduledEndTime := now.Add(time.Duration(exam.Paper.DurationMinutes) * time.Minute)

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
		http.Error(w, "Failed to update exam and participants", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
