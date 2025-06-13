package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/dtos"
)

func handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch st.Code() {
	case codes.InvalidArgument, codes.FailedPrecondition:
		http.Error(w, st.Message(), http.StatusBadRequest)
	case codes.Unauthenticated:
		http.Error(w, st.Message(), http.StatusUnauthorized)
	case codes.AlreadyExists:
		http.Error(w, st.Message(), http.StatusConflict)
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
	case codes.PermissionDenied:
		http.Error(w, st.Message(), http.StatusForbidden)
	default:
		http.Error(w, st.Message(), http.StatusInternalServerError)
	}
}

// GetInt64FromVars parses an int64 value from a map using the given key.
// Returns error if the value is missing or cannot be parsed.
func getInt64FromVars(vars map[string]string, key string) (int64, error) {
	val, ok := vars[key]
	if !ok {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}

	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: must be a number", key)
	}

	return id, nil
}

// getHashFromVars parses an string value from a map using the given key.
// Returns error if the value is missing or empty.
func getHashFromVars(vars map[string]string, key string) (string, error) {
	hash, ok := vars[key]
	if !ok || hash == "" {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}

	return hash, nil
}

func getQuestionIdFromVars(vars map[string]string) (string, error) {
	questionID, err := getHashFromVars(vars, "questionId")
	if err != nil {
		return "", fmt.Errorf("Missing question ID")
	}
	return questionID, nil
}

func getPaperIdFromVars(vars map[string]string) (string, error) {
	paperID, err := getHashFromVars(vars, "paperId")
	if err != nil {
		return "", fmt.Errorf("Missing paper ID")
	}
	return paperID, nil
}

func getExamIdFromVars(vars map[string]string) (string, error) {
	examID, err := getHashFromVars(vars, "examId")
	if err != nil {
		return "", fmt.Errorf("Missing exam ID")
	}
	return examID, nil
}

func mapUserProfileToDto(profile *proto.UserProfileResponse) dtos.UserResponseDto {
	return dtos.UserResponseDto{
		ID:        profile.Id,
		Username:  profile.Username,
		Email:     profile.Email,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
	}
}
