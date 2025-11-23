package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/config/validate"
	"pariksha/gateway/internal/dtos"
	"pariksha/gateway/internal/interservice"
	"pariksha/gateway/internal/middlewares"
	"pariksha/gateway/internal/services"
)

func GetExamParticipants(w http.ResponseWriter, r *http.Request) {
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	participants, err := examService.Client().GetExamParticipants(ctx, &proto.ExamRequest{
		ExamHash: examHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := make([]dtos.ExamParticipantResponseDto, len(participants.Participants))
	for i, p := range participants.Participants {
		// Get user details from auth service
		userResp, err := interservice.GetUser(&proto.GetUserRequest{
			UserId: p.UserId,
		})

		response[i] = dtos.ExamParticipantResponseDto{
			ID:           p.ParticipantId,
			UserID:       p.UserId,
			Status:       p.Status,
			ScoreAwarded: p.ScoreAwarded,
		}

		if err == nil {
			response[i].FirstName = userResp.FirstName
			response[i].LastName = userResp.LastName
			response[i].Email = userResp.Email
		}

		if p.StartedAt != nil {
			startedAt := p.StartedAt.AsTime()
			response[i].StartedAt = &startedAt
		}
		if p.EndedAt != nil {
			endedAt := p.EndedAt.AsTime()
			response[i].EndedAt = &endedAt
		}
		if p.ScheduledEndTime != nil {
			scheduledEndTime := p.ScheduledEndTime.AsTime()
			response[i].ScheduledEndTime = &scheduledEndTime
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExamParticipant gets the participant data for the current user
func GetExamParticipant(w http.ResponseWriter, r *http.Request) {
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	participant, err := examService.Client().GetExamParticipant(ctx, &proto.GetExamParticipantRequest{
		ExamHash: examHash,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.GetExamParticipantResponseDto{
		ID:           participant.ParticipantId,
		ScoreAwarded: participant.ScoreAwarded,
	}

	if participant.StartedAt != nil {
		response.StartedAt = participant.StartedAt.AsTime()
	}
	if participant.ScheduledEndTime != nil {
		response.ScheduledEndTime = participant.ScheduledEndTime.AsTime()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AddExamParticipant(w http.ResponseWriter, r *http.Request) {
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var participantDto dtos.AddExamParticipantDto
	if err := json.NewDecoder(r.Body).Decode(&participantDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(participantDto)
	if errs != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	// First create/update user if email is provided
	if participantDto.Email != "" {
		userResp, err := interservice.UpsertUser(&proto.UpsertUserRequest{
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
		ExamHash: examHash,
		UserId:   participantDto.UserID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.AddExamParticipantResponseDto{
		ID:           participant.ParticipantId,
		UserID:       participant.UserId,
		Status:       participant.Status,
		ScoreAwarded: participant.ScoreAwarded,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func RemoveExamParticipant(w http.ResponseWriter, r *http.Request) {
	examHash, err := getExamIdFromVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	_, err = examService.Client().RemoveExamParticipant(ctx, &proto.RemoveParticipantRequest{
		ExamHash:      examHash,
		ParticipantId: participantID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetParticipantById(w http.ResponseWriter, r *http.Request) {
	participantID, err := getInt64FromVars(mux.Vars(r), "participantId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(r.Context(), userID)

	participant, err := examService.Client().GetParticipantById(ctx, &proto.ParticipantRequest{
		ParticipantId: participantID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.ParticipantDetailResponseDto{
		ID:     participant.ParticipantId,
		Status: participant.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
