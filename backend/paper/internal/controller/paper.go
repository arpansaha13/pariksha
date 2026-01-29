package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/middleware"
	"pariksha/paper/internal/service"
)

type Paper struct {
	paperSvc service.IPaperService
}

func NewPaper(s service.IPaperService) *Paper {
	return &Paper{
		paperSvc: s,
	}
}

func (c *Paper) HandleGetUserPapers(ctx context.Context, _ *emptypb.Empty) (*proto.PaperList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return c.paperSvc.GetUserPapers(ctx, userID)
}

func (c *Paper) HandleCreatePaper(ctx context.Context, _ *emptypb.Empty) (*proto.CreatePaperResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return c.paperSvc.CreatePaper(ctx, userID)
}

func (c *Paper) HandleGetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	return c.paperSvc.GetPaper(ctx, req.PaperHash)
}

func (c *Paper) HandleUpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*emptypb.Empty, error) {
	if err := c.paperSvc.UpdatePaper(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Paper) HandleGetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	perm, err := middleware.GetPermissionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.PaperPermissionsResponse{
		CanRead:  perm.CanRead(),
		CanWrite: perm.CanWrite(),
	}, nil
}

func (c *Paper) HandleDeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*emptypb.Empty, error) {
	if err := c.paperSvc.DeletePapers(ctx, req.PaperHashes); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
