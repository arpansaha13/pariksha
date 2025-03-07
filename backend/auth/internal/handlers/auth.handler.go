package handlers

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/auth/internal/config/db"
	"pariksha/auth/internal/config/env"
	"pariksha/auth/internal/services"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
)

type AuthServer struct {
	proto.UnimplementedAuthServiceServer
}

func generateUniqueUsername(tx *gorm.DB, emailPrefix string) (string, error) {
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		randomHash, err := utils.GenerateAlphaNum(6)
		if err != nil {
			return "", err
		}

		username := emailPrefix + "_" + randomHash

		var exists int64
		if err := tx.Model(&models.User{}).Where("username = ?", username).Count(&exists).Error; err != nil {
			return "", err
		}

		if exists == 0 {
			return username, nil
		}
	}

	return "", status.Error(codes.Internal, "failed to generate unique username after maximum retries")
}

func createSession(userID int) (*models.Session, error) {
	sessionKey := uuid.New()
	sessionExpiresAt := time.Now().Add(time.Duration(env.SESSION_EXPIRES_IN_HOURS) * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     sessionExpiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(env.JWT_SECRET_KEY))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	csrfToken, err := utils.GenerateBase64String(constants.CSRF_TOKEN_LENGTH)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate csrf token")
	}

	session := &models.Session{
		Key:       sessionKey,
		Token:     tokenString,
		ExpiresAt: sessionExpiresAt,
		CsrfToken: csrfToken,
	}

	if err := db.Sessions.Create(session).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	return session, nil
}

func (s *AuthServer) LoginWithPassword(ctx context.Context, req *proto.LoginWithPasswordRequest) (*proto.UserResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).Take(&user).Error; err != nil || !user.Verified {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	session, err := createSession(user.ID)
	if err != nil {
		return nil, err
	}

	// Add session info to response headers
	md := metadata.Pairs(
		"session-key", session.Key.String(),
		"csrf-token", session.CsrfToken,
		"expires-at", session.ExpiresAt.Format(time.RFC3339),
	)
	grpc.SetHeader(ctx, md)

	return &proto.UserResponse{
		Id:        int32(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
	}, nil
}

func (s *AuthServer) InitiateLoginWithOtp(ctx context.Context, req *proto.LoginWithOtpRequest) (*proto.EmptyResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	// Check if user exists and is verified
	var user models.User
	if err := db.DB.Where("email = ?", req.Email).Take(&user).Error; err != nil || !user.Verified {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	otp, _ := utils.GenerateOTP(constants.VERIFICATION_OTP_LENGTH)
	otpExpiresInMinutes := env.OTP_EXPIRES_IN_MINUTES
	otpExpiresAt := time.Now().Add(time.Duration(otpExpiresInMinutes) * time.Minute)

	otpEntry := models.Otp{
		Email:        req.Email,
		OTP:          otp,
		OTPExpiresAt: otpExpiresAt,
		Purpose:      constants.OTP_PURPOSE_LOGIN,
	}

	if err := db.DB.Save(&otpEntry).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create OTP")
	}

	services.MailService.SendLoginOtpMail(&types.MailRequestLoginOtp{
		To:               req.Email,
		Otp:              otp,
		ExpiresInMinutes: otpExpiresInMinutes,
	})

	return &proto.EmptyResponse{}, nil
}

func (s *AuthServer) SignUp(ctx context.Context, req *proto.SignUpRequest) (*proto.EmptyResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).Take(&user).Error; err == nil && user.Verified {
		return nil, status.Error(codes.AlreadyExists, "email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Create or update user
		err := tx.Where("email = ?", req.Email).Take(&user).Error
		if err == nil {
			// Update password if user already exists
			user.Password.String = string(hashedPassword)
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			username, err := generateUniqueUsername(tx, strings.Split(req.Email, "@")[0])
			if err != nil {
				return err
			}

			user = models.User{
				Email:    req.Email,
				Username: username,
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
		otpExpiresAt := time.Now().Add(time.Duration(env.OTP_EXPIRES_IN_MINUTES) * time.Minute)

		otpEntry := models.Otp{
			Email:        req.Email,
			OTP:          otp,
			OTPExpiresAt: otpExpiresAt,
			Purpose:      constants.OTP_PURPOSE_SIGNUP,
		}
		if err := tx.Save(&otpEntry).Error; err != nil {
			return err
		}

		services.MailService.SendVerificationMail(&types.MailRequestVerification{
			To:               req.Email,
			Otp:              otp,
			ExpiresInMinutes: env.OTP_EXPIRES_IN_MINUTES,
		})

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to process signup")
	}

	return &proto.EmptyResponse{}, nil
}

func (s *AuthServer) VerifySignup(ctx context.Context, req *proto.VerificationRequest) (*proto.UserResponse, error) {
	if req.Email == "" || req.Otp == "" {
		return nil, status.Error(codes.InvalidArgument, "email and OTP are required")
	}

	var otpEntry models.Otp
	var user models.User
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("email = ? AND purpose = ?", req.Email, constants.OTP_PURPOSE_SIGNUP).Take(&otpEntry).Error; err != nil {
			return status.Error(codes.Unauthenticated, "invalid email")
		}

		if req.Otp != otpEntry.OTP || time.Now().After(otpEntry.OTPExpiresAt) {
			return status.Error(codes.Unauthenticated, "invalid or expired OTP")
		}

		if err := tx.Where("email = ?", req.Email).Take(&user).Error; err != nil {
			return status.Error(codes.NotFound, "user not found")
		}

		if err := tx.Model(&user).Update("verified", true).Error; err != nil {
			return status.Error(codes.Internal, "failed to verify user")
		}

		if err := tx.Delete(&otpEntry).Error; err != nil {
			return status.Error(codes.Internal, "failed to delete OTP entry")
		}

		session, err := createSession(user.ID)
		if err != nil {
			return err
		}

		md := metadata.Pairs(
			"session-key", session.Key.String(),
			"csrf-token", session.CsrfToken,
			"expires-at", session.ExpiresAt.Format(time.RFC3339),
		)
		grpc.SetHeader(ctx, md)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.UserResponse{
		Id:        int32(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
	}, nil
}

func (s *AuthServer) VerifyLoginOtp(ctx context.Context, req *proto.VerificationRequest) (*proto.UserResponse, error) {
	if req.Email == "" || req.Otp == "" {
		return nil, status.Error(codes.InvalidArgument, "email and OTP are required")
	}

	var otpEntry models.Otp
	var user models.User
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("email = ? AND purpose = ?", req.Email, constants.OTP_PURPOSE_LOGIN).Take(&otpEntry).Error; err != nil {
			return status.Error(codes.NotFound, "invalid email")
		}

		if req.Otp != otpEntry.OTP || time.Now().After(otpEntry.OTPExpiresAt) {
			return status.Error(codes.InvalidArgument, "invalid or expired OTP")
		}

		if err := tx.Where("email = ?", req.Email).Take(&user).Error; err != nil {
			return status.Error(codes.NotFound, "user not found")
		}

		if !user.Verified {
			return status.Error(codes.PermissionDenied, "user not verified")
		}

		session, err := createSession(user.ID)
		if err != nil {
			return err
		}

		md := metadata.Pairs(
			"session-key", session.Key.String(),
			"csrf-token", session.CsrfToken,
			"expires-at", session.ExpiresAt.Format(time.RFC3339),
		)
		grpc.SetHeader(ctx, md)

		if err := tx.Delete(&otpEntry).Error; err != nil {
			return status.Error(codes.Internal, "failed to delete OTP entry")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.UserResponse{
		Id:        int32(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
	}, nil
}

func (s *AuthServer) ForgotPassword(ctx context.Context, req *proto.ForgotPasswordRequest) (*proto.EmptyResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	var user models.User
	if err := db.DB.Where("email = ? AND verified = ?", req.Email, true).Take(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "email not found or not verified")
	}

	otp, _ := utils.GenerateOTP(constants.VERIFICATION_OTP_LENGTH)
	otpExpiresAt := time.Now().Add(time.Duration(env.OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD) * time.Minute)

	otpEntry := models.Otp{
		Email:        req.Email,
		OTP:          otp,
		OTPExpiresAt: otpExpiresAt,
		Purpose:      constants.OTP_PURPOSE_FORGOT_PASSWORD,
	}

	if err := db.DB.Save(&otpEntry).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create OTP")
	}

	services.MailService.SendForgotPasswordMail(&types.MailRequestForgotPassword{
		To:               req.Email,
		Otp:              otp,
		ExpiresInMinutes: env.OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD,
	})

	return &proto.EmptyResponse{}, nil
}

func (s *AuthServer) ResetPassword(ctx context.Context, req *proto.ResetPasswordRequest) (*proto.EmptyResponse, error) {
	if req.Email == "" || req.NewPassword == "" || req.Otp == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}

	var otpEntry models.Otp
	if err := db.DB.Where("email = ? AND otp = ? AND purpose = ?",
		req.Email, req.Otp, constants.OTP_PURPOSE_FORGOT_PASSWORD).Take(&otpEntry).Error; err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid or expired OTP")
	}

	if time.Now().After(otpEntry.OTPExpiresAt) {
		return nil, status.Error(codes.InvalidArgument, "OTP has expired")
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).Take(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if !user.Verified {
		return nil, status.Error(codes.PermissionDenied, "user not verified")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash new password")
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("password", sql.NullString{String: string(hashedPassword), Valid: true}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&otpEntry).Error; err != nil {
			return err
		}

		services.MailService.SendResetPasswordMail(&types.MailRequestResetPassword{To: req.Email})
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to reset password")
	}

	return &proto.EmptyResponse{}, nil
}

func (s *AuthServer) Authenticate(ctx context.Context, req *proto.AuthenticateRequest) (*proto.AuthenticateResponse, error) {
	if req.SessionKey == "" || req.CsrfToken == "" {
		return nil, status.Error(codes.InvalidArgument, "session key and csrf token are required")
	}

	var session models.Session
	if err := db.Sessions.Where("key = ?", req.SessionKey).Take(&session).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid session")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, status.Error(codes.Unauthenticated, "session expired")
	}

	if session.CsrfToken != req.CsrfToken {
		return nil, status.Error(codes.Unauthenticated, "invalid csrf token")
	}

	token, err := jwt.Parse(session.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(env.JWT_SECRET_KEY), nil
	})
	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid token claims")
	}

	userID := int32(claims["user_id"].(float64))
	return &proto.AuthenticateResponse{UserId: userID}, nil
}
