package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
)

func (s *PaperServer) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.GetBoilerplateResponse, error) {
	var boilerplate models.Boilerplate
	err := db.DB.Where("question_id = ? AND language_id = ?", req.QuestionId, req.LanguageId).Take(&boilerplate).Error
	if err != nil {
		return nil, status.Error(codes.NotFound, "boilerplate not found")
	}

	return &proto.GetBoilerplateResponse{
		Code: boilerplate.Code,
	}, nil
}
