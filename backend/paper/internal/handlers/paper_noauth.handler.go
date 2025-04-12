package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
)

func (s *PaperServer) GetPaperDuration(ctx context.Context, req *proto.PaperRequest) (*proto.PaperDurationResponse, error) {
	var paper models.Paper
	if err := db.DB.Select("duration_minutes").Take(&paper, req.PaperId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "paper not found")
	}

	return &proto.PaperDurationResponse{
		DurationMinutes: int32(paper.DurationMinutes),
	}, nil
}
