package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetAuthUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := services.GetAuthService().Client().GetUser(r.Context(), &proto.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	userDto := mapUserProfileToDto(resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDto)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := getInt64FromVars(mux.Vars(r), "userId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := services.GetAuthService().Client().GetUser(r.Context(), &proto.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	userDto := mapUserProfileToDto(resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDto)
}

func UpdateAuthUser(w http.ResponseWriter, r *http.Request) {
	var userDto dtos.UpdateUserDto
	if err := json.NewDecoder(r.Body).Decode(&userDto); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(userDto)
	if errs != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	req := &proto.UpdateUserRequest{
		UserId: userID,
	}

	if userDto.Username != "" {
		req.Username = &userDto.Username
	}
	if userDto.FirstName != "" {
		req.FirstName = &userDto.FirstName
	}
	if userDto.LastName != "" {
		req.LastName = &userDto.LastName
	}

	resp, err := services.GetAuthService().Client().UpdateUser(r.Context(), req)

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
