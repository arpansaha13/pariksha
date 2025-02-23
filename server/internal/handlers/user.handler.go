package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/arpansaha13/common/pkg/models"
	"github.com/arpansaha13/pariksha/internal/config/db"
	"github.com/arpansaha13/pariksha/internal/config/validate"
	"github.com/arpansaha13/pariksha/internal/dtos"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var user models.User
	if err := db.DB.Take(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var userDto dtos.UpdateUserDto
	if err := json.NewDecoder(r.Body).Decode(&userDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(userDto)
	if errs != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	var user models.User
	if err := db.DB.Take(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find user", http.StatusInternalServerError)
		}
		return
	}

	isUpdated := false
	notUpdatedFields := make(map[string]string)

	if userDto.Username != "" && userDto.Username != user.Username {
		var existingUser models.User
		if err := db.DB.Where("username = ?", userDto.Username).First(&existingUser).Error; err == nil {
			notUpdatedFields["Username"] = "Username is already taken"
		} else {
			user.Username = userDto.Username
			isUpdated = true
		}
	}
	if userDto.FirstName != "" && userDto.FirstName != user.FirstName.String {
		user.FirstName = sql.NullString{String: userDto.FirstName, Valid: true}
		isUpdated = true
	}
	if userDto.LastName != "" && userDto.LastName != user.LastName.String {
		user.LastName = sql.NullString{String: userDto.LastName, Valid: true}
		isUpdated = true
	}

	if isUpdated {
		if err := db.DB.Save(&user).Error; err != nil {
			http.Error(w, "Failed to update user info", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"updated_fields":     userDto,
		"not_updated_fields": notUpdatedFields,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
