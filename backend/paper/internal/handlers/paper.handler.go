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
	"pariksha/paper/internal/utils/validate"
)

type PaperServer struct {
	proto.UnimplementedPaperServer
}

func (s *PaperServer) GetUserPapers(ctx context.Context, _ *proto.Empty) (*proto.PaperList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var papers []models.Paper
	err = db.DB.
		Select("papers.id, papers.title, papers.max_score, papers.duration_minutes, papers.question_counts, papers.created_by").
		Joins("INNER JOIN permissions ON permissions.paper_id = papers.id").
		Where("permissions.user_id = ?", userID).
		Find(&papers).Error

	if err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
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
	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		paper = models.Paper{CreatedBy: userID} // Will use database default for Title
		if err := tx.Create(&paper).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Create default category
		defaultCategory := models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: paper.ID, Valid: true},
			Name:    "Category 1",
			Order:   1,
		}
		if err := tx.Create(&defaultCategory).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Create permissions entry with write access
		permissions := models.PaperPermission{
			PaperID: paper.ID,
			UserID:  userID,
		}
		permissions.SetWrite()

		if err := tx.Create(&permissions).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return paperToProto(paper), nil
}

func (s *PaperServer) GetPaper(ctx context.Context, req *proto.PaperRequest) (*proto.PaperResponse, error) {
	paper, err := utils.FindRecord[models.Paper](db.DB, req.PaperId, "paper not found")
	if err != nil {
		return nil, err
	}
	return paperToProto(*paper), nil
}

func (s *PaperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*proto.Empty, error) {
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		paper, err := utils.FindRecord[models.Paper](tx, req.PaperId, "paper not found")
		if err != nil {
			return err
		}

		isUpdated := false

		if req.Title != nil && req.GetTitle() != "" && req.GetTitle() != paper.Title {
			paper.Title = req.GetTitle()
			isUpdated = true
		}

		if req.DurationMinutes != nil {
			if err := validate.PaperDuration(req.GetDurationMinutes()); err != nil {
				return err
			}
			if int16(req.GetDurationMinutes()) != paper.DurationMinutes {
				paper.DurationMinutes = int16(req.GetDurationMinutes())
				isUpdated = true
			}
		}

		if isUpdated {
			if err := tx.Save(&paper).Error; err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) GetPaperPermissions(ctx context.Context, req *proto.PaperRequest) (*proto.PaperPermissionsResponse, error) {
	pc, err := NewPaperContext(ctx)
	if err != nil || pc.Permissions == nil {
		return nil, status.Error(codes.Internal, "permissions data not found in context")
	}

	return &proto.PaperPermissionsResponse{
		CanRead:  pc.Permissions.CanRead(),
		CanWrite: pc.Permissions.CanWrite(),
	}, nil
}

func (s *PaperServer) DeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*proto.Empty, error) {
	if len(req.PaperIds) == 0 {
		return &proto.Empty{}, nil
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Delete all specified papers
		if err := tx.Where("id IN ?", req.PaperIds).Delete(&models.Paper{}).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Delete all non-locked questions for these papers
		if err := tx.Where("paper_id IN ? AND locked = ?", req.PaperIds, false).
			Delete(&models.Question{}).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Delete all non-locked categories for these papers
		if err := tx.Where("paper_id IN ? AND locked = ?", req.PaperIds, false).
			Delete(&models.QuestionCategory{}).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Delete all permissions for these papers
		if err := tx.Where("paper_id IN ?", req.PaperIds).
			Delete(&models.PaperPermission{}).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
