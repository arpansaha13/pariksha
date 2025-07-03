package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/env"
	"pariksha/server/internal/services"
)

func setCookiesFromMetadata(w http.ResponseWriter, md metadata.MD) {
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

	var header metadata.MD
	authService := services.GetAuthService()
	response, err := authService.Client().LoginWithPassword(
		context.Background(),
		&loginReq,
		grpc.Header(&header),
	)
	if err != nil {
		handleGRPCError(w, err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setCookiesFromMetadata(w, header)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var signUpReq proto.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&signUpReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	authService := services.GetAuthService()
	_, err := authService.Client().SignUp(context.Background(), &signUpReq)
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

	var header metadata.MD
	authService := services.GetAuthService()
	response, err := authService.Client().VerifySignUp(
		context.Background(),
		&verificationReq,
		grpc.Header(&header),
	)
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	setCookiesFromMetadata(w, header)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func LoginWithOtp(w http.ResponseWriter, r *http.Request) {
	var loginOtpReq proto.LoginWithOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&loginOtpReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	authService := services.GetAuthService()
	_, err := authService.Client().InitiateLoginWithOtp(context.Background(), &loginOtpReq)
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

	var header metadata.MD
	authService := services.GetAuthService()
	response, err := authService.Client().VerifyLoginOtp(
		context.Background(),
		&verificationReq,
		grpc.Header(&header),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setCookiesFromMetadata(w, header)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var forgotPasswordReq proto.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&forgotPasswordReq); err != nil {
		http.Error(w, INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	authService := services.GetAuthService()
	_, err := authService.Client().ForgotPassword(context.Background(), &forgotPasswordReq)
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

	authService := services.GetAuthService()
	_, err := authService.Client().ResetPassword(context.Background(), &resetPasswordReq)
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

	authService := services.GetAuthService()
	_, err = authService.Client().Logout(context.Background(), &proto.LogoutRequest{
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
