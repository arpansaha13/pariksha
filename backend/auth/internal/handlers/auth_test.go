package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"pariksha/auth/internal/config/db"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
)

func TestSignUp(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.SignUpRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Successful signup",
			req: &proto.SignUpRequest{
				Email:    "new@example.com",
				Password: "newPass123",
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var user models.User
				err := db.DB.Where("email = ?", "new@example.com").First(&user).Error
				assert.NoError(t, err)
				assert.False(t, user.Verified)

				var otp models.Otp
				err = db.DB.Where("email = ? AND purpose = ?", "new@example.com", constants.OTP_PURPOSE_SIGNUP).First(&otp).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, otp.OTP)
				assert.True(t, otp.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Email already exists and verified",
			setup: func(t *testing.T) {
				createTestUser(t, "existing@example.com", true)
			},
			req: &proto.SignUpRequest{
				Email:    "existing@example.com",
				Password: "pass123",
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "Unverified user signup again",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.SignUpRequest{
				Email:    testUnverifiedEmail,
				Password: "newPass123",
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var otp models.Otp
				err := db.DB.Where("email = ? AND purpose = ?", testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP).First(&otp).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, otp.OTP)
			},
		},
		{
			name: "Missing email",
			req: &proto.SignUpRequest{
				Password: "pass123",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing password",
			req: &proto.SignUpRequest{
				Email: "test@example.com",
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.SignUp(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestVerifySignup(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.VerificationRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, resp *proto.UserResponse, md metadata.MD)
	}{
		{
			name: "Successful verification",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UserResponse, md metadata.MD) {
				// Verify user response
				assert.Equal(t, testUnverifiedEmail, resp.Email)
				assert.NotEmpty(t, resp.Username)

				// Verify user is now verified
				var user models.User
				err := db.DB.Where("email = ?", testUnverifiedEmail).First(&user).Error
				assert.NoError(t, err)
				assert.True(t, user.Verified)

				// Verify session was created
				assert.NotEmpty(t, md.Get(constants.HEADER_SESSION_KEY))
				assert.NotEmpty(t, md.Get(constants.HEADER_CSRF_TOKEN))
				assert.NotEmpty(t, md.Get(constants.HEADER_EXPIRES_AT))

				// Verify OTP was deleted
				var otp models.Otp
				err = db.DB.Where("email = ?", testUnverifiedEmail).First(&otp).Error
				assert.Error(t, err)
			},
		},
		{
			name: "Expired OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, true)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Invalid OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   invalidOTP,
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Missing email",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Otp: validOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Non-existent email",
			setup: func(t *testing.T) {
				createTestOTP(t, "nonexistent@example.com", constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: "nonexistent@example.com",
				Otp:   validOTP,
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			var md metadata.MD
			resp, err := client.VerifySignup(context.Background(), tt.req, grpc.Header(&md))

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, resp, md)
				}
			}

			clearTables(t)
		})
	}
}

func TestForgotPassword(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.ForgotPasswordRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Success - OTP created",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.ForgotPasswordRequest{
				Email: testVerifiedEmail,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var otp models.Otp
				err := db.DB.Where("email = ? AND purpose = ?",
					testVerifiedEmail,
					constants.OTP_PURPOSE_FORGOT_PASSWORD,
				).First(&otp).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, otp.OTP)
				assert.True(t, otp.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.ForgotPasswordRequest{
				Email: testUnverifiedEmail,
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Non-existent email",
			req: &proto.ForgotPasswordRequest{
				Email: "nonexistent@example.com",
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Missing email",
			req: &proto.ForgotPasswordRequest{
				Email: "",
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.ForgotPassword(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestResetPassword(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.ResetPasswordRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Successful password reset",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_FORGOT_PASSWORD, false)
			},
			req: &proto.ResetPasswordRequest{
				Email:       testVerifiedEmail,
				NewPassword: "newPass123",
				Otp:         validOTP,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var user models.User
				err := db.DB.Where("email = ?", testVerifiedEmail).First(&user).Error
				assert.NoError(t, err)
				err = bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte("newPass123"))
				assert.NoError(t, err)

				var otp models.Otp
				err = db.DB.Where("email = ?", testVerifiedEmail).First(&otp).Error
				assert.Error(t, err)
			},
		},
		{
			name: "Expired OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_FORGOT_PASSWORD, true)
			},
			req: &proto.ResetPasswordRequest{
				Email:       testVerifiedEmail,
				NewPassword: "newPass123",
				Otp:         validOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_FORGOT_PASSWORD, false)
			},
			req: &proto.ResetPasswordRequest{
				Email:       testUnverifiedEmail,
				NewPassword: "newPass123",
				Otp:         validOTP,
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Missing fields",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.ResetPasswordRequest{
				Email: testVerifiedEmail,
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.ResetPassword(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestLoginWithPassword(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.LoginWithPasswordRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, resp *proto.UserResponse, md metadata.MD)
	}{
		{
			name: "Successful login",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.LoginWithPasswordRequest{
				Email:    testVerifiedEmail,
				Password: "testPass123",
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UserResponse, md metadata.MD) {
				assert.Equal(t, testVerifiedEmail, resp.Email)
				assert.Equal(t, "verified", resp.Username)
				assert.NotEmpty(t, md.Get(constants.HEADER_SESSION_KEY))
				assert.NotEmpty(t, md.Get(constants.HEADER_CSRF_TOKEN))
				assert.NotEmpty(t, md.Get(constants.HEADER_EXPIRES_AT))

				sessionKey := md.Get(constants.HEADER_SESSION_KEY)[0]
				var session models.Session
				err := db.Sessions.Where("key = ?", sessionKey).First(&session).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, session.Token)
				assert.True(t, session.ExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Missing credentials",
			req: &proto.LoginWithPasswordRequest{
				Email:    "",
				Password: "",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid password",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.LoginWithPasswordRequest{
				Email:    testVerifiedEmail,
				Password: "wrongPass",
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Non-existent user",
			req: &proto.LoginWithPasswordRequest{
				Email:    "nonexistent@example.com",
				Password: "testPass123",
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.LoginWithPasswordRequest{
				Email:    testUnverifiedEmail,
				Password: "testPass123",
			},
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			var md metadata.MD
			resp, err := client.LoginWithPassword(context.Background(), tt.req, grpc.Header(&md))

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, resp, md)
				}
			}

			clearTables(t)
		})
	}
}

func TestInitiateLoginWithOtp(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.LoginWithOtpRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Success - OTP created",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.LoginWithOtpRequest{
				Email: testVerifiedEmail,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var otpEntry models.Otp
				result := db.DB.Where("email = ? AND purpose = ?",
					testVerifiedEmail,
					constants.OTP_PURPOSE_LOGIN,
				).First(&otpEntry)
				assert.NoError(t, result.Error)
				assert.NotEmpty(t, otpEntry.OTP)
				assert.True(t, otpEntry.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Missing email",
			req: &proto.LoginWithOtpRequest{
				Email: "",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.LoginWithOtpRequest{
				Email: testUnverifiedEmail,
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Non-existent user",
			req: &proto.LoginWithOtpRequest{
				Email: "nonexistent@example.com",
			},
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.InitiateLoginWithOtp(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestVerifyLoginOtp(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.VerificationRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, resp *proto.UserResponse, md metadata.MD)
	}{
		{
			name: "Success - Valid OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_LOGIN, false)
			},
			req: &proto.VerificationRequest{
				Email: testVerifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UserResponse, md metadata.MD) {
				assert.Equal(t, testVerifiedEmail, resp.Email)
				assert.NotEmpty(t, md.Get(constants.HEADER_SESSION_KEY))
				assert.NotEmpty(t, md.Get(constants.HEADER_CSRF_TOKEN))

				var otpEntry models.Otp
				result := db.DB.Where("email = ?", testVerifiedEmail).First(&otpEntry)
				assert.Error(t, result.Error)
			},
		},
		{
			name: "Invalid OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_LOGIN, true)
			},
			req: &proto.VerificationRequest{
				Email: testVerifiedEmail,
				Otp:   invalidOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_LOGIN, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Missing fields",
			req: &proto.VerificationRequest{
				Email: "",
				Otp:   "",
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			var md metadata.MD
			resp, err := client.VerifyLoginOtp(context.Background(), tt.req, grpc.Header(&md))

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, resp, md)
				}
			}

			clearTables(t)
		})
	}
}

func TestLogout(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		req          *proto.LogoutRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, sessionKey string)
	}{
		{
			name: "Successful logout",
			setup: func(t *testing.T) string {
				// Create a session that expires in future
				session := &models.Session{
					Key:       uuid.New(),
					Token:     "valid_token",
					ExpiresAt: time.Now().Add(24 * time.Hour),
					CsrfToken: "csrf_token",
				}
				err := db.Sessions.Create(session).Error
				assert.NoError(t, err)
				return session.Key.String()
			},
			req:          &proto.LogoutRequest{}, // SessionKey will be set from setup
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, sessionKey string) {
				var session models.Session
				err := db.Sessions.Where("key = ?", sessionKey).First(&session).Error
				assert.NoError(t, err)
				assert.True(t, session.ExpiresAt.Before(time.Now()), "session should be expired")
			},
		},
		{
			name: "Missing session key",
			req: &proto.LogoutRequest{
				SessionKey: "",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Non-existent session",
			req: &proto.LogoutRequest{
				SessionKey: uuid.New().String(),
			},
			expectedCode: codes.OK, // We return success even if session doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sessionKey string
			if tt.setup != nil {
				sessionKey = tt.setup(t)
				tt.req.SessionKey = sessionKey
			}

			resp, err := client.Logout(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, sessionKey)
				}
			}

			clearTables(t)
		})
	}
}
