package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
)

type PaperServer struct {
	proto.UnimplementedPaperServiceServer
}

func getUserID(ctx context.Context) (int32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing user id")
	}

	userID, err := strconv.Atoi(userIDs[0])
	if err != nil || userID == 0 {
		return 0, status.Error(codes.InvalidArgument, "invalid user id")
	}

	return int32(userID), nil
}

// Helper function to verify paper access
func verifyPaperAccess(tx *gorm.DB, paperID interface{}, userID int32, ownershipType string) error {
	if tx == nil {
		tx = db.DB
	}

	var actualPaperID int
	switch v := paperID.(type) {
	case sql.NullInt64:
		if !v.Valid {
			return status.Error(codes.InvalidArgument, "invalid paper id")
		}
		actualPaperID = int(v.Int64)
	case int:
		actualPaperID = v
	default:
		return status.Error(codes.InvalidArgument, "invalid paper id type")
	}

	var condition string
	var args []any

	args = append(args, actualPaperID, userID)
	condition = "po.paper_id = ? AND po.user_id = ?"

	if ownershipType != "" {
		condition += " AND po.type = ?"
		args = append(args, ownershipType)
	}

	var exists bool
	err := tx.Raw(`SELECT EXISTS (
			SELECT 1 FROM paper_ownerships po
			WHERE `+condition+`)`, args...).
		Scan(&exists).Error

	if err != nil {
		return status.Error(codes.Internal, "failed to check paper access")
	}

	if !exists {
		return status.Error(codes.PermissionDenied, "no permission to perform this action")
	}

	return nil
}

// Helper function to convert Paper model to proto response
func paperToProto(paper models.Paper) *proto.PaperResponse {
	var questionCounts proto.QuestionCount
	json.Unmarshal(paper.QuestionCounts, &questionCounts)

	return &proto.PaperResponse{
		Id:              int32(paper.ID),
		Title:           paper.Title,
		MaxScore:        int32(paper.MaxScore),
		DurationMinutes: int32(paper.DurationMinutes),
		QuestionCounts:  &questionCounts,
		Ownership: &proto.PaperOwnership{
			Id:   int32(paper.PaperOwnership.ID),
			Path: paper.PaperOwnership.Path,
			Type: paper.PaperOwnership.Type,
		},
	}
}

func (s *PaperServer) GetUserPapers(ctx context.Context, _ *proto.Empty) (*proto.PaperList, error) {
	userID, err := getUserID(ctx)
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
	userID, err := getUserID(ctx)
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
			UserID:  int(userID),
			PaperID: paper.ID,
			Type:    constants.PAPER_OWNERSHIP_TYPE_OWNER,
		}

		if err := tx.Create(&paperOwnership).Error; err != nil {
			return err
		}

		// Create default category for questions in this paper
		defaultCategory := models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
	userID, err := getUserID(ctx)
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
	userID, err := getUserID(ctx)
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
