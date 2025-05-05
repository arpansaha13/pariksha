package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pariksha/common/pkg/proto"
	"pariksha/server/internal/dtos"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func mapUserProfileToDto(profile *proto.UserProfileResponse) dtos.UserResponseDto {
	return dtos.UserResponseDto{
		ID:        profile.Id,
		Username:  profile.Username,
		Email:     profile.Email,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
	}
}

func mapQuestionToDto(resp *proto.QuestionResponse) dtos.QuestionResponseDto {
	tags, _ := json.Marshal(resp.Tags)

	response := dtos.QuestionResponseDto{
		ID:            resp.Id,
		Type:          resp.Type,
		Tags:          tags,
		PaperID:       resp.PaperId,
		MaxScore:      int(resp.MaxScore),
		CategoryID:    resp.CategoryId,
		CorrectAnswer: resp.GetCorrectAnswer(),
	}

	switch q := resp.Question.(type) {
	case *proto.QuestionResponse_Mcq:
		data, _ := json.Marshal(q.Mcq)
		response.Question = data
	case *proto.QuestionResponse_General:
		data, _ := json.Marshal(q.General)
		response.Question = data
	}

	return response
}
