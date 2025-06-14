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
	"pariksha/common/pkg/utils/generate"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
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
		Preload("PaperHash").
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

		// Create paper hash
		paperHash := models.PaperHash{
			ID:   paper.ID,
			Hash: generate.HMACHash(int64(paper.ID)),
		}
		if err := tx.Create(&paperHash).Error; err != nil {
			return status.Error(codes.Internal, "failed to create paper hash")
		}
		paper.PaperHash = paperHash

		// Create default category
		defaultCategory := models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
	paperID, ok := interceptors.GetPaperIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "paper ID not found in context")
	}

	var paper models.Paper
	if err := db.DB.Preload("PaperHash").First(&paper, paperID).Error; err != nil {
		return nil, utils.HandleDBError(err, "paper not found")
	}

	return paperToProto(paper), nil
}

func (s *PaperServer) UpdatePaper(ctx context.Context, req *proto.UpdatePaperRequest) (*proto.Empty, error) {
	paperID, ok := interceptors.GetPaperIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "paper ID not found in context")
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		var paper models.Paper
		if err := tx.Preload("PaperHash").First(&paper, paperID).Error; err != nil {
			return utils.HandleDBError(err, "paper not found")
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
	perm, err := interceptors.GetPermissionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.PaperPermissionsResponse{
		CanRead:  perm.CanRead(),
		CanWrite: perm.CanWrite(),
	}, nil
}

func (s *PaperServer) DeletePapers(ctx context.Context, req *proto.DeletePapersRequest) (*proto.Empty, error) {
	paperIDs, ok := interceptors.GetPaperIDsFromContext(ctx)
	if !ok || len(paperIDs) == 0 {
		return &proto.Empty{}, nil
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Convert []types.PaperID to []int64
		ids := make([]int64, len(paperIDs))
		for i, id := range paperIDs {
			ids[i] = int64(id)
		}

		// Delete paper hashes first due to foreign key constraint
		if err := tx.Where("id IN ?", ids).Delete(&models.PaperHash{}).Error; err != nil {
			return status.Error(codes.Internal, "failed to delete paper hashes")
		}

		// Delete all specified papers
		if err := tx.Where("id IN ?", ids).Delete(&models.Paper{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Delete all non-locked questions for these papers
		if err := tx.Where("paper_id IN ? AND locked = ?", ids, false).
			Delete(&models.Question{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Delete all non-locked categories for these papers
		if err := tx.Where("paper_id IN ? AND locked = ?", ids, false).
			Delete(&models.QuestionCategory{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Delete all permissions for these papers
		if err := tx.Where("paper_id IN ?", ids).
			Delete(&models.PaperPermission{}).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) GetQuestionHashes(ctx context.Context, req *proto.GetQuestionHashesRequest) (*proto.GetQuestionHashesResponse, error) {
	if len(req.QuestionIds) == 0 {
		return &proto.GetQuestionHashesResponse{}, nil
	}

	var questionHashes []models.QuestionHash
	if err := db.DB.Where("id IN ?", req.QuestionIds).Find(&questionHashes).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Create a map for quick lookup
	hashMap := make(map[int64]string)
	for _, qh := range questionHashes {
		hashMap[int64(qh.ID)] = qh.Hash
	}

	// Create response maintaining input order
	response := make([]string, len(req.QuestionIds))
	for i, id := range req.QuestionIds {
		if hash, exists := hashMap[id]; exists {
			response[i] = hash
		}
	}

	return &proto.GetQuestionHashesResponse{
		QuestionHashes: response,
	}, nil
}

func (s *PaperServer) GetQuestionIds(ctx context.Context, req *proto.GetQuestionIdsRequest) (*proto.GetQuestionIdsResponse, error) {
	questionIDs, ok := interceptors.GetQuestionIDsFromContext(ctx)
	if !ok {
		return &proto.GetQuestionIdsResponse{}, nil
	}

	ids := make([]int64, len(questionIDs))
	for i, id := range questionIDs {
		ids[i] = int64(id)
	}

	return &proto.GetQuestionIdsResponse{
		QuestionIds: ids,
	}, nil
}
