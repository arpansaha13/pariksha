package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/config/validate"
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

	errs := validate.Do.Struct(loginDto)
	if errs != nil {
		http.Error(w, "Invald request body", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", loginDto.Email).Take(&user).Error; err != nil || !user.Verified {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(loginDto.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	sessionKey := uuid.New()
	sessionExpiresInHours, _ := strconv.Atoi(
		utils.GetEnvWithDefault("SESSION_EXPIRES_IN_HOURS", constants.DEFAULT_SESSION_EXPIRES_IN_HOURS),
	)
	sessionExpiresAt := time.Now().Add(time.Duration(sessionExpiresInHours) * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     sessionExpiresAt.Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	session := models.Session{
		Key:       sessionKey,
		Token:     tokenString,
		ExpiresAt: sessionExpiresAt,
	}

	if err := db.DB.Create(&session).Error; err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     utils.GetEnvWithDefault("SESSION_COOKIE_NAME", constants.DEFAULT_SESSION_COOKIE_NAME),
		Value:    sessionKey.String(),
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"firstName": user.FirstName.String,
		"lastName":  user.LastName.String,
	})
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var signUpDto dtos.SignUpDto

	if err := json.NewDecoder(r.Body).Decode(&signUpDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(signUpDto)
	if errs != nil {
		http.Error(w, "Invald request body", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", signUpDto.Email).Take(&user).Error; err == nil && user.Verified {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	otp, _ := utils.GenerateOTP(constants.VERIFICATION_OTP_LENGTH)
	linkHash, _ := utils.GenerateAlphaNum(constants.VERIFICATION_HASH_LENGTH)

	otpExpiresInMinutes, _ := strconv.Atoi(
		utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES", constants.DEFAULT_OTP_EXPIRES_IN_MINUTES),
	)
	otpExpiresAt := time.Now().Add(time.Duration(otpExpiresInMinutes) * time.Minute)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUpDto.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// check if the email already exists in UnverifiedUsers
	var unverifiedUser models.UnverifiedUser
	err = db.DB.Where("email = ?", signUpDto.Email).Take(&unverifiedUser).Error

	if err == nil {
		// if yes, then just update the otp in that row

		unverifiedUser.OTP = otp
		unverifiedUser.OTPExpiresAt = otpExpiresAt
		if err := db.DB.Save(&unverifiedUser).Error; err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to update OTP. Please try again later.", http.StatusInternalServerError)
			return
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// else create a new entry.

		newUnverifiedUser := models.UnverifiedUser{
			Hash:         linkHash,
			OTP:          otp,
			OTPExpiresAt: otpExpiresAt,
			Email:        signUpDto.Email,
			Password:     string(hashedPassword),
		}
		if err := db.DB.Create(&newUnverifiedUser).Error; err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to create user. Please try again later.", http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println(err)
		http.Error(w, "Failed to process request. Please try again later.", http.StatusInternalServerError)
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
	params := mux.Vars(r)
	hash := params["hash"]

	var verificationDto dtos.VerificationDto
	if err := json.NewDecoder(r.Body).Decode(&verificationDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(verificationDto)
	if errs != nil {
		http.Error(w, "Invald request body", http.StatusBadRequest)
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var unverifiedUser *models.UnverifiedUser
		if err := tx.Where("hash = ?", hash).Take(&unverifiedUser).Error; err != nil {
			http.Error(w, "Invalid or expired link", constants.StatusInvalidToken)
			return err
		}

		if verificationDto.OTP != unverifiedUser.OTP || time.Now().After(unverifiedUser.OTPExpiresAt) {
			http.Error(w, "Invalid or expired OTP", http.StatusUnauthorized)
			return errors.New("invalid or expired otp")
		}

		var newUser *models.User

		if err := tx.Model(models.User{}).Where("email = ?", unverifiedUser.Email).Take(&newUser).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Something went wrong!", http.StatusInternalServerError)
				return err
			}

			newUser = &models.User{
				Email:    unverifiedUser.Email,
				Password: sql.NullString{String: unverifiedUser.Password, Valid: true},
				Username: strings.Split(unverifiedUser.Email, "@")[0],
			}

			if err := tx.Create(newUser).Error; err != nil {
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return err
			}
		} else {
			newUser.Verified = true
			newUser.Password = sql.NullString{String: unverifiedUser.Password, Valid: true}

			if err := tx.Save(&newUser).Error; err != nil {
				http.Error(w, "Failed to save user", http.StatusInternalServerError)
				return err
			}
		}

		if err := tx.Delete(unverifiedUser).Error; err != nil {
			http.Error(w, "Failed to delete unverified user", http.StatusInternalServerError)
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
