package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
	"github.com/arpansaha13/pariksha/internal/repositories"
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
	userRepository := repositories.GetUserRepository()
	unverifiedUserRepository := repositories.GetUnverifiedUserRepository()

	var err error
	var signUpDto dtos.SignUpDto

	if err := json.NewDecoder(r.Body).Decode(&signUpDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if _, err = userRepository.FindOneByEmail(signUpDto.Email); err == nil {
		http.Error(w, "This email is already registered", http.StatusConflict)
		return
	}

	otp, _ := utils.GenerateOTP(constants.VERIFICATION_OTP_LENGTH)
	linkHash, _ := utils.GenerateAlphaNum(constants.VERIFICATION_HASH_LENGTH)

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

	if err := unverifiedUserRepository.Create(&unverifiedUser); err != nil {
		fmt.Println(err)
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

func Verification(w http.ResponseWriter, r *http.Request) {
	unverifiedUserRepository := repositories.GetUnverifiedUserRepository()
	userRepository := repositories.GetUserRepository()

	params := mux.Vars(r)
	hash := params["hash"]

	var err error

	var verificationDto dtos.VerificationDto
	if err := json.NewDecoder(r.Body).Decode(&verificationDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var unverifiedUser *models.UnverifiedUser
	if unverifiedUser, err = unverifiedUserRepository.FindOne(hash); err != nil {
		http.Error(w, "Invalid or expired link", constants.StatusInvalidToken)
		return
	}

	if verificationDto.OTP != unverifiedUser.OTP || time.Now().After(unverifiedUser.OTPExpiresAt) {
		http.Error(w, "Invalid or expired OTP", http.StatusUnauthorized)
		return
	}

	newUser := models.User{
		Email:    unverifiedUser.Email,
		Password: unverifiedUser.Password,
		Username: strings.Split(unverifiedUser.Email, "@")[0],
	}

	if err = userRepository.Create(&newUser); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	if err = unverifiedUserRepository.DeleteByPointer(unverifiedUser); err != nil {
		http.Error(w, "Failed to delete unverified user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
