package handlers

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"
)

type PaperServer struct {
	proto.UnimplementedPaperServiceServer
}

func (s *PaperServer) GetUserPapers(ctx context.Context, _ *proto.Empty) (*proto.PaperList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var papers []models.Paper
	err = db.DB.
		Joins("INNER JOIN paper_ownerships ON paper_ownerships.paper_id = papers.id").
		Where("paper_ownerships.user_id = ?", userID).
		Preload("PaperOwnership", "user_id = ?", userID).
		Find(&papers).Error

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve papers")
	}

	response := &proto.PaperList{
		Papers: make([]*proto.PaperResponse, len(papers)),
	}

	for i, paper := range papers {
		response.Papers[i] = paperToProto(paper)
	}

	return response, nil
}

func (s *PaperServer) CreatePaper(ctx context.Context, _ *proto.Empty) (*proto.PaperResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var paper models.Paper
	var paperOwnership models.PaperOwnership

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		paper = models.Paper{} // Will use database default for Title

		if err := tx.Create(&paper).Error; err != nil {
			return err
		}

		paperOwnership = models.PaperOwnership{
			UserID:  userID,
			PaperID: paper.ID,
			Type:    constants.PAPER_OWNERSHIP_TYPE_OWNER,
		}

		if err := tx.Create(&paperOwnership).Error; err != nil {
			return err
		}

		// Create default category for questions in this paper
		defaultCategory := models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: paper.ID, Valid: true},
			Name:    "Category 1",
			Order:   1,
		}
		if err := tx.Create(&defaultCategory).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create paper")
	}

	paper.PaperOwnership = paperOwnership // For creating response object in `paperToProto`
	return paperToProto(paper), nil
}

func (s *PaperServer) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var paper models.Paper
	err = db.DB.Preload("PaperOwnership").Take(&paper, req.PaperId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "paper not found")
		}
		return nil, status.Error(codes.Internal, "failed to retrieve paper")
	}

	if err := verifyPaperAccess(nil, paper.ID, userID, ""); err != nil {
		return nil, err
	}

	return paperToProto(paper), nil
}

func (s *PaperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var paper models.Paper
	err = db.DB.Take(&paper, req.PaperId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "paper not found")
		}
		return nil, status.Error(codes.Internal, "failed to find paper")
	}

	if err := verifyPaperAccess(nil, paper.ID, userID, constants.PAPER_OWNERSHIP_TYPE_OWNER); err != nil {
		return nil, err
	}

	isUpdated := false

	if req.Title != "" {
		paper.Title = req.Title
		isUpdated = true
	}

	if isUpdated {
		if err := db.DB.Save(&paper).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to update paper")
		}
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) CheckPaperAccess(ctx context.Context, req *proto.PaperRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Check if paper exists
	exists, err := paperExists(req.PaperId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, status.Error(codes.NotFound, "paper not found")
	}

	// Check if user has owner access
	if err := verifyPaperAccess(nil, req.PaperId, userID, constants.PAPER_OWNERSHIP_TYPE_OWNER); err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
