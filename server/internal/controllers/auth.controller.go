package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
	"github.com/arpansaha13/pariksha/internal/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var loginDto dtos.LoginDto

	if err := json.NewDecoder(r.Body).Decode(&loginDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("Login attempt with email: %s\n", loginDto.Email)
	fmt.Fprintf(w, "Login endpoint reached!")
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var signUpDto dtos.SignUpDto

	if err := json.NewDecoder(r.Body).Decode(&signUpDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingUser models.User
	if result := db.DB.Where("email = ?", signUpDto.Email).First(&existingUser); result.Error == nil {
		http.Error(w, "This email is already registered", http.StatusConflict)
		return
	}

	otp, _ := utils.GenerateOTP(constants.VERIFICATION_OTP_LENGTH)
	linkHash, _ := utils.GenerateBase64(constants.VERIFICATION_HASH_LENGTH)

	otpExpiresInMinutes, _ := strconv.Atoi(
		utils.GetEnvWithDefault(
			"OTP_EXPIRES_IN_MINUTES",
			constants.DEFAULT_OTP_EXPIRES_IN_MINUTES,
		),
	)
	otpExpiresAt := time.Now().Add(time.Duration(otpExpiresInMinutes) * time.Minute)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUpDto.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	unverifiedUser := models.UnverifiedUser{
		Hash:         linkHash,
		OTP:          otp,
		OTPExpiresAt: otpExpiresAt,
		Email:        signUpDto.Email,
		Password:     string(hashedPassword),
	}

	if result := db.DB.Create(&unverifiedUser); result.Error != nil {
		http.Error(w, "Failed to create user. Please try again later.", http.StatusInternalServerError)
		return
	}

	msg := utils.CreateVerificationMail(signUpDto.Email, otp, linkHash, otpExpiresInMinutes)

	if err := utils.SendEmail(signUpDto.Email, msg); err != nil {
		// Log the error but don't return it to the user since the account was created
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
