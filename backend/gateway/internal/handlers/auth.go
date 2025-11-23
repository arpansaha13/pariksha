package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/config/env"
	"pariksha/gateway/internal/interservice"
)

func setCookiesFromMetadata(w http.ResponseWriter, mdPtr *metadata.MD) error {
	if mdPtr == nil {
		return fmt.Errorf("mdPtr is nil")
	}
	md := *mdPtr

	sessionKey := md.Get(constants.HEADER_SESSION_KEY)[0]
	csrfToken := md.Get(constants.HEADER_CSRF_TOKEN)[0]
	expiresAt, _ := time.Parse(time.RFC3339, md.Get(constants.HEADER_EXPIRES_AT)[0])

	http.SetCookie(w, &http.Cookie{
		Name:     env.SESSION_COOKIE_NAME,
		Value:    sessionKey,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/", // Token wont reach frontend server middlewares if path is not root
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     env.CSRFTOKEN_COOKIE_NAME,
		Value:    csrfToken,
		Expires:  expiresAt,
		HttpOnly: false,
		Secure:   true,
		Path:     "/", // JavaScript can read cookie only if current path matches cookie path
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

type AuthCheckResponse struct {
	Valid bool `json:"valid"`
}

func CheckAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthCheckResponse{Valid: true})
}

func LoginWithPassword(w http.ResponseWriter, r *http.Request) {
	var loginReq proto.LoginWithPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	response, header, err := interservice.LoginWithPassword(
		&loginReq,
	)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	err = setCookiesFromMetadata(w, header)
	if err != nil {
		http.Error(w, constants.ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var signUpReq proto.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&signUpReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	_, err := interservice.SignUp(&signUpReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func VerifySignUp(w http.ResponseWriter, r *http.Request) {
	var verificationReq proto.VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&verificationReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	response, header, err := interservice.VerifySignUp(
		&verificationReq,
	)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	err = setCookiesFromMetadata(w, header)
	if err != nil {
		http.Error(w, constants.ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func LoginWithOtp(w http.ResponseWriter, r *http.Request) {
	var loginOtpReq proto.LoginWithOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&loginOtpReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	_, err := interservice.InitiateLoginWithOtp(&loginOtpReq)
	if err != nil {
		// Convert gRPC status error to appropriate HTTP status code
		st, ok := status.FromError(err)
		if !ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		case codes.Unauthenticated:
			http.Error(w, st.Message(), http.StatusUnauthorized)
		default:
			http.Error(w, st.Message(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func VerifyLoginWithOtp(w http.ResponseWriter, r *http.Request) {
	var verificationReq proto.VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&verificationReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	response, header, err := interservice.VerifyLoginOtp(
		&verificationReq,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = setCookiesFromMetadata(w, header)
	if err != nil {
		http.Error(w, constants.ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var forgotPasswordReq proto.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&forgotPasswordReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	_, err := interservice.ForgotPassword(&forgotPasswordReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var resetPasswordReq proto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&resetPasswordReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	_, err := interservice.ResetPassword(&resetPasswordReq)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(env.SESSION_COOKIE_NAME)
	if err != nil {
		http.Error(w, "No session found", http.StatusBadRequest)
		return
	}

	_, err = interservice.Logout(&proto.LogoutRequest{
		SessionKey: cookie.Value,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Delete cookies by setting expiration to past
	http.SetCookie(w, &http.Cookie{
		Name:     env.SESSION_COOKIE_NAME,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     env.CSRFTOKEN_COOKIE_NAME,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-24 * time.Hour),
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusNoContent)
}
