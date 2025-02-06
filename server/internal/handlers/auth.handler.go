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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", signUpDto.Email).Take(&user).Error; err == nil && user.Verified {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUpDto.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Create or update user
		err := tx.Where("email = ?", signUpDto.Email).Take(&user).Error
		if err == nil {
			user.Password.String = string(hashedPassword)

			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				Email:    signUpDto.Email,
				Username: strings.Split(signUpDto.Email, "@")[0],
				Password: sql.NullString{String: string(hashedPassword), Valid: true},
				Verified: false,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		} else {
			return err
		}

		otp, _ := utils.GenerateOTP(constants.VERIFICATION_OTP_LENGTH)
		otpExpiresInMinutes, _ := strconv.Atoi(
			utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES", constants.DEFAULT_OTP_EXPIRES_IN_MINUTES),
		)
		otpExpiresAt := time.Now().Add(time.Duration(otpExpiresInMinutes) * time.Minute)

		// Create or update OTP entry
		otpEntry := models.Otp{
			Email:        signUpDto.Email,
			OTP:          otp,
			OTPExpiresAt: otpExpiresAt,
			Purpose:      constants.OTP_PURPOSE_SIGNUP,
		}
		if err := tx.Save(&otpEntry).Error; err != nil {
			return err
		}

		msg := utils.CreateVerificationMail(signUpDto.Email, otp, otpExpiresInMinutes)
		if err := utils.SendEmail(signUpDto.Email, msg); err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to process signup", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func Verification(w http.ResponseWriter, r *http.Request) {
	var verificationDto dtos.VerificationDto
	if err := json.NewDecoder(r.Body).Decode(&verificationDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(verificationDto)
	if errs != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var otpEntry models.Otp
		if err := tx.Where("email = ?", verificationDto.Email).Take(&otpEntry).Error; err != nil {
			http.Error(w, "Invalid email", http.StatusUnauthorized)
			return err
		}

		if verificationDto.OTP != otpEntry.OTP || time.Now().After(otpEntry.OTPExpiresAt) {
			http.Error(w, "Invalid or expired OTP", http.StatusUnauthorized)
			return errors.New("invalid or expired otp")
		}

		if otpEntry.Purpose == constants.OTP_PURPOSE_SIGNUP {
			if err := tx.Model(&models.User{}).Where("email = ?", verificationDto.Email).
				Update("verified", true).Error; err != nil {
				http.Error(w, "Failed to verify user", http.StatusInternalServerError)
				return err
			}
		}

		if err := tx.Delete(&otpEntry).Error; err != nil {
			http.Error(w, "Failed to delete OTP entry", http.StatusInternalServerError)
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
