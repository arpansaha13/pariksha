package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/paper/internal/config/db"
)

func TestGetUserPapers(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperList)
	}{
		{
			name: "Success - Get user papers",
			setup: func(t *testing.T) {
				paper1 := createTestPaper(t, userID)
				err := db.DB.Model(&paper1).Update("question_counts", `{"mcq": 2, "short": 1, "long": 0}`).Error
				require.NoError(t, err)

				paper2 := createTestPaper(t, userID)
				err = db.DB.Model(&paper2).Update("question_counts", `{"mcq": 1, "short": 0, "long": 1}`).Error
				require.NoError(t, err)
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.Equal(t, 2, len(resp.Papers))

				// Validate first paper
				assert.Equal(t, int32(2), resp.Papers[0].QuestionCounts.Mcq)
				assert.Equal(t, int32(1), resp.Papers[0].QuestionCounts.Short)
				assert.Equal(t, int32(0), resp.Papers[0].QuestionCounts.Long)

				// Validate second paper
				assert.Equal(t, int32(1), resp.Papers[1].QuestionCounts.Mcq)
				assert.Equal(t, int32(0), resp.Papers[1].QuestionCounts.Short)
				assert.Equal(t, int32(1), resp.Papers[1].QuestionCounts.Long)
			},
		},
		{
			name:         "Success - No papers",
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.Equal(t, 0, len(resp.Papers))
			},
		},
		{
			name:         "Invalid user ID",
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			if tt.setup != nil {
				tt.setup(t)
			}

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetUserPapers(ctx, &proto.Empty{})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestCreatePaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperResponse)
	}{
		{
			name:         "Success - Create paper",
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "Untitled Paper", resp.Title)

				// Verify default category was created
				var categories []models.QuestionCategory
				err := db.DB.Where("paper_id = ?", resp.Id).Find(&categories).Error
				require.NoError(t, err)
				assert.Equal(t, 1, len(categories))
				assert.Equal(t, "Category 1", categories[0].Name)

				// Verify paper permissions
				var permissions models.PaperPermissions
				err = db.DB.Where("paper_id = ? AND user_id = ?", resp.Id, userID).Take(&permissions).Error
				require.NoError(t, err)
				assert.True(t, permissions.CanWrite(), "User should have write access to the paper")
			},
		},
		{
			name:         "Invalid user ID",
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			if tt.setup != nil {
				tt.setup(t)
			}

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.CreatePaper(ctx, &proto.Empty{})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestUpdatePaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int64
		request      *proto.UpdatePaperRequest
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper)
	}{
		{
			name: "Success - Update title",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				Title: ptr.String("Updated Title"),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Updated Title", updated.Title)
				assert.Equal(t, 60, updated.DurationMinutes) // Default duration unchanged
			},
		},
		{
			name: "Success - Update duration",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(90),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Test Paper", updated.Title) // Original title unchanged
				assert.Equal(t, 90, updated.DurationMinutes)
			},
		},
		{
			name: "Success - Update both title and duration",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				Title:           ptr.String("New Title"),
				DurationMinutes: ptr.Int32(120),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "New Title", updated.Title)
				assert.Equal(t, 120, updated.DurationMinutes)
			},
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				Title: ptr.String("Updated Title"),
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Invalid user ID",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: 0,
			request: &proto.UpdatePaperRequest{
				Title: ptr.String("Updated Title"),
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Empty paper title",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				Title: ptr.String(""), // Empty title
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.Take(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Test Paper", updated.Title) // Unchanged title
			},
		},
		{
			name: "Failure - Negative duration",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(-30), // Negative duration
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Failure - Zero duration",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(0), // Zero duration
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.Take(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, 0, updated.DurationMinutes) // Updates to zero
			},
		},
		{
			name: "Failure - Extremely large duration",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID: userID,
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(1441), // More than 24 hours
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			tt.request.PaperId = paper.ID
			_, err := client.UpdatePaper(ctx, tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, paper)
			}
		})
	}
}

func TestGetPaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperResponse)
	}{
		{
			name: "Success - Get owned paper",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 2, "short": 1, "long": 1}`).Error
				require.NoError(t, err)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "Test Paper", resp.Title)
				assert.Equal(t, int32(60), resp.DurationMinutes)

				// Validate question counts
				assert.NotNil(t, resp.QuestionCounts)
				assert.Equal(t, int32(2), resp.QuestionCounts.Mcq)
				assert.Equal(t, int32(1), resp.QuestionCounts.Short)
				assert.Equal(t, int32(1), resp.QuestionCounts.Long)

				// Verify paper permissions
				var permissions models.PaperPermissions
				err := db.DB.Where("paper_id = ? AND user_id = ?", resp.Id, userID).Take(&permissions).Error
				require.NoError(t, err)
				assert.True(t, permissions.CanWrite(), "User should have write access to the paper")
			},
		},
		{
			name: "Success - Get shared paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2
				paper := createTestPaper(t, 2)

				// Create read-only permissions for test user
				permissions := models.PaperPermissions{
					UserID:  userID,
					PaperID: paper.ID,
				}
				permissions.SetRead()
				err := db.DB.Create(&permissions).Error
				require.NoError(t, err)

				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)

				// Verify paper permissions
				var permissions models.PaperPermissions
				err := db.DB.Where("paper_id = ? AND user_id = ?", resp.Id, userID).Take(&permissions).Error
				require.NoError(t, err)
				assert.True(t, permissions.CanRead(), "User should have read access to the paper")
				assert.False(t, permissions.CanWrite(), "User should not have write access to the paper")
			},
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "No access to paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2 with no sharing
				paper := createTestPaper(t, 2)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Invalid user ID",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetPaper(ctx, &proto.PaperRequest{
				PaperId: paper.ID,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestCheckPaperAccess(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int64
		expectedCode codes.Code
	}{
		{
			name: "Success - Owner access",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "Has shared access to paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2
				paper := createTestPaper(t, 2)

				// Create read-only permissions for test user
				permissions := models.PaperPermissions{
					PaperID: paper.ID,
					UserID:  userID,
				}
				permissions.SetRead()
				err := db.DB.Create(&permissions).Error
				require.NoError(t, err)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
		},
		{
			name: "No access to paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2 with no sharing
				paper := createTestPaper(t, 2)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Invalid user ID",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.CheckPaperAccess(ctx, &proto.PaperRequest{
				PaperId: paper.ID,
			})

			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}
