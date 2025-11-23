package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/config/validate"
	"pariksha/gateway/internal/dtos"
	"pariksha/gateway/internal/services"
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

	// Convert DTO test cases to proto test cases
	testCases := make([]*proto.TestCase, len(requestDto.TestCases))
	for i, tc := range requestDto.TestCases {
		testCases[i] = &proto.TestCase{
			Inputs:         tc.Inputs,
			ExpectedOutput: tc.ExpectedOutput,
		}
	}

	engineService := services.GetEngineService()
	response, err := engineService.Client().RunCode(context.Background(), &proto.RunCodeRequest{
		QuestionHash: requestDto.QuestionID,
		Code:         requestDto.Code,
		Environment:  requestDto.Environment,
		TestCases:    testCases,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	testCaseResults := make([]dtos.TestCaseResult, len(response.Results))

	for i, result := range response.Results {
		testCaseResults[i] = dtos.TestCaseResult{
			Inputs:         result.Inputs,
			Output:         result.Output,
			ExpectedOutput: result.ExpectedOutput,
			Stdout:         result.Stdout,
			Error:          result.Error,
			Status:         result.Status,
			ExecutionTime:  result.ExecutionTime,
		}
	}

	responseDto := dtos.RunCodeResponseDto{
		Compilation: dtos.CompilationResult{
			Success: response.Compilation.Success,
			Stderr:  response.Compilation.Stderr,
		},
		Results: testCaseResults,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseDto)
}
