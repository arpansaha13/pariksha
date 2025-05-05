package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetExamParticipants(w http.ResponseWriter, r *http.Request) {
	examID, err := getInt64FromVars(mux.Vars(r), "examId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	participants, err := examService.Client().GetExamParticipants(ctx, &proto.ExamRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	authService := services.GetAuthService()
	authCtx := context.Background()

	response := make([]dtos.ExamParticipantResponseDto, len(participants.Participants))
	for i, p := range participants.Participants {
		// Get user details from auth service
		userResp, err := authService.Client().GetUser(authCtx, &proto.GetUserRequest{
			UserId: p.UserId,
		})

		response[i] = dtos.ExamParticipantResponseDto{
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
	examID, err := getInt64FromVars(mux.Vars(r), "examId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	participant, err := examService.Client().GetExamParticipant(ctx, &proto.GetExamParticipantRequest{
		ExamId: examID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.GetExamParticipantResponseDto{
		ID: participant.ParticipantId,
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
	examID, err := getInt64FromVars(mux.Vars(r), "examId")
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
		ExamId: examID,
		UserId: participantDto.UserID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.AddExamParticipantResponseDto{
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
	examID, err := getInt64FromVars(vars, "examId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	participantID, err := getInt64FromVars(vars, "participantId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	examService := services.GetExamService()
	ctx := examService.CreateMetadata(userID)

	_, err = examService.Client().RemoveExamParticipant(ctx, &proto.RemoveParticipantRequest{
		ExamId:        examID,
		ParticipantId: participantID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
