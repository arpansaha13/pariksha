package controllers

import (
	"context"
	"encoding/json"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/services"
)

type Paper struct {
	paperSvc *services.Paper
}

func NewPaper(s *services.Paper) *Paper {
	return &Paper{
		paperSvc: s,
	}
}

func (c *Paper) HandleGetUserPapers(ctx context.Context, _ *proto.Empty) (*proto.PaperList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return c.paperSvc.GetUserPapers(ctx, userID)
}

func (c *Paper) HandleCreatePaper(ctx context.Context, _ *proto.Empty) (*proto.PaperResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return c.paperSvc.CreatePaper(ctx, userID)
}

func (c *Paper) HandleGetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	var paper models.Paper
	if err := db.DB.Where("hash = ?", req.PaperHash).First(&paper).Error; err != nil {
		return nil, utils.HandleDBError(err, "paper not found")
	}

	return paperToProto(paper), nil
}

func (c *Paper) HandleUpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*proto.Empty, error) {
	if err := c.paperSvc.UpdatePaper(ctx, req); err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

func (c *Paper) HandleGetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	perm, err := interceptors.GetPermissionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.PaperPermissionsResponse{
		CanRead:  perm.CanRead(),
		CanWrite: perm.CanWrite(),
	}, nil
}

func (c *Paper) HandleDeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*proto.Empty, error) {
	if err := c.paperSvc.DeletePapers(ctx, req.PaperHashes); err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

// Helper function to convert Paper model to proto response
func paperToProto(paper models.Paper) *proto.PaperResponse {
	var questionCounts proto.QuestionCount
	json.Unmarshal(paper.QuestionCounts, &questionCounts)

	return &proto.PaperResponse{
		PaperHash:       paper.Hash,
		Title:           paper.Title,
		MaxScore:        int32(paper.MaxScore),
		DurationMinutes: int32(paper.DurationMinutes),
		QuestionCounts:  &questionCounts,
		CreatedBy:       int64(paper.CreatedBy),
	}
}
