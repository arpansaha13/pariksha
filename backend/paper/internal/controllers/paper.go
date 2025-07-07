package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
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

func (c *Paper) HandleGetUserPapers(ctx context.Context, _ *emptypb.Empty) (*proto.PaperList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return c.paperSvc.GetUserPapers(userID)
}

func (c *Paper) HandleCreatePaper(ctx context.Context, _ *emptypb.Empty) (*proto.CreatePaperResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return c.paperSvc.CreatePaper(userID)
}

func (c *Paper) HandleGetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	return c.paperSvc.GetPaper(req.PaperHash)
}

func (c *Paper) HandleUpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*emptypb.Empty, error) {
	if err := c.paperSvc.UpdatePaper(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
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

func (c *Paper) HandleDeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*emptypb.Empty, error) {
	if err := c.paperSvc.DeletePapers(ctx, req.PaperHashes); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
