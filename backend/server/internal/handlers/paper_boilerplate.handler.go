package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/middlewares"
	"pariksha/server/internal/services"
)

func GetBoilerplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	questionID := r.Context().Value(middlewares.DecryptedQuestionID).(int64)

	languageID, err := getInt64FromVars(vars, "languageId")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int64)
	ctx := services.GetPaperService().CreateMetadata(userID)
	resp, err := services.GetPaperService().Client().GetBoilerplate(ctx, &proto.GetBoilerplateRequest{
		QuestionId: questionID,
		LanguageId: int32(languageID),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	response := dtos.GetBoilerplateResponseDto{
		Code: resp.Code,
	}

	json.NewEncoder(w).Encode(response)
}
