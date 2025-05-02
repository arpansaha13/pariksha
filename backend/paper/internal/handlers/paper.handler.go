package handlers

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

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
		Select("papers.id, papers.title, papers.max_score, papers.duration_minutes, papers.question_counts").
		Joins("INNER JOIN permissions ON permissions.paper_id = papers.id").
		Where("permissions.user_id = ?", userID).
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

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		paper = models.Paper{} // Will use database default for Title

		if err := tx.Create(&paper).Error; err != nil {
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

		// Create permissions entry with write access
		permissions := models.PaperPermissions{
			PaperID: paper.ID,
			UserID:  userID,
		}
		permissions.SetWrite()

		if err := tx.Create(&permissions).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create paper")
	}

	return paperToProto(paper), nil
}

func (s *PaperServer) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	var paper models.Paper
	if err := db.DB.Take(&paper, req.PaperId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "paper not found")
		}
		return nil, status.Error(codes.Internal, "failed to retrieve paper")
	}

	return paperToProto(paper), nil
}

func (s *PaperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*proto.Empty, error) {
	var paper models.Paper
	err := db.DB.Take(&paper, req.PaperId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "paper not found")
		}
		return nil, status.Error(codes.Internal, "failed to find paper")
	}

	isUpdated := false

	if req.Title != nil && req.GetTitle() != paper.Title {
		paper.Title = req.GetTitle()
		isUpdated = true
	}

	if req.DurationMinutes != nil && int(req.GetDurationMinutes()) != paper.DurationMinutes {
		paper.DurationMinutes = int(req.GetDurationMinutes())
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
	return &proto.Empty{}, nil
}
