package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
	"pariksha/server/internal/services"
)

func RunCode(w http.ResponseWriter, r *http.Request) {
	var requestDto dtos.RunCodeRequestDto
	if err := json.NewDecoder(r.Body).Decode(&requestDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(requestDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	engineService := services.GetEngineService()
	response, err := engineService.Client().RunCode(context.Background(), &proto.RunCodeRequest{
		Code:        requestDto.Code,
		Environment: requestDto.Environment,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	responseDto := dtos.RunCodeResponseDto{
		Stdout:        response.Stdout,
		Stderr:        response.Stderr,
		Error:         response.Error,
		ExitCode:      response.ExitCode,
		ExecutionTime: response.ExecutionTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseDto)
}
