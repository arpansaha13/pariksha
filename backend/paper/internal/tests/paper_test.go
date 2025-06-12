package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/db"
)

func TestGetUserPapers(t *testing.T) {
	tests := []ListPapersTestCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get user papers",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) {
				paper1 := createTestPaper(t, userID)
				updatePaperCounts(t, &paper1, `{"mcq": 2, "subjective": 1}`)

				paper2 := createTestPaper(t, userID)
				updatePaperCounts(t, &paper2, `{"mcq": 1, "subjective": 0}`)
			},
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.EqualValues(t, 2, len(resp.Papers))

				// Validate first paper
				assert.EqualValues(t, 2, resp.Papers[0].QuestionCounts.Mcq)
				assert.EqualValues(t, 1, resp.Papers[0].QuestionCounts.Subjective)

				// Validate second paper
				assert.EqualValues(t, 1, resp.Papers[1].QuestionCounts.Mcq)
				assert.EqualValues(t, 0, resp.Papers[1].QuestionCounts.Subjective)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - No papers",
				userID:       userID,
				expectedCode: codes.OK,
			},
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.EqualValues(t, 0, len(resp.Papers))
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Invalid user ID",
				userID:       0,
				expectedCode: codes.InvalidArgument,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			if tt.setup != nil {
				tt.setup(t)
			}

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode, &proto.Empty{},
				client.GetUserPapers,
				tt.validate)
		})
	}
}

func TestCreatePaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       types.UserID
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperResponse)
	}{
		{
			name:         "Success - Create paper",
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.PaperId)
				assert.Equal(t, "Untitled Paper", resp.Title)

				// Verify default category was created
				var categories []models.QuestionCategory
				err := db.DB.Where("paper_id = ?", resp.PaperId).Find(&categories).Error
				require.NoError(t, err)
				assert.EqualValues(t, 1, len(categories))
				assert.Equal(t, "Category 1", categories[0].Name)

				// Verify paper permissions
				var permissions models.PaperPermission
				err = db.DB.Where("paper_id = ? AND user_id = ?", resp.PaperId, userID).Take(&permissions).Error
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
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.Empty{},
				client.CreatePaper,
				tt.validate)
		})
	}
}

func TestUpdatePaper(t *testing.T) {
	tests := []UpdatePaperCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update title",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				Title: ptr.String("Updated Title"),
			},
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Updated Title", updated.Title)
				assert.EqualValues(t, 60, updated.DurationMinutes)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update duration",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(90),
			},
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Test Paper", updated.Title) // Original title unchanged
				assert.EqualValues(t, 90, updated.DurationMinutes)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update both title and duration",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				Title:           ptr.String("New Title"),
				DurationMinutes: ptr.Int32(120),
			},
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "New Title", updated.Title)
				assert.EqualValues(t, 120, updated.DurationMinutes)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Paper not found",
				userID:       userID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			request: &proto.UpdatePaperRequest{
				Title: ptr.String("Updated Title"),
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Invalid user ID",
				userID:       0,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				Title: ptr.String("Updated Title"),
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Failure - Empty paper title",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				Title: ptr.String(""), // Empty title
			},
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.Take(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Test Paper", updated.Title) // Unchanged title
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Failure - Negative duration",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(-30), // Negative duration
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Failure - Zero duration",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(0), // Zero duration
			},
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.Take(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.EqualValues(t, 0, updated.DurationMinutes) // Updates to zero
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Failure - Extremely large duration",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			request: &proto.UpdatePaperRequest{
				DurationMinutes: ptr.Int32(1441), // More than 24 hours
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			tt.request.PaperId = int64(paper.ID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.UpdatePaper,
				func(t *testing.T, _ *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, paper)
					}
				})
		})
	}
}

func TestGetPaper(t *testing.T) {
	tests := []GetPaperTestCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get owned paper",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				updatePaperCounts(t, &paper, `{"mcq": 2, "subjective": 1}`)
				return &paper
			},
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.PaperId)
				assert.Equal(t, "Test Paper", resp.Title)
				assert.EqualValues(t, 60, resp.DurationMinutes)

				// Validate question counts
				assert.NotNil(t, resp.QuestionCounts)
				assert.EqualValues(t, 2, resp.QuestionCounts.Mcq)
				assert.EqualValues(t, 1, resp.QuestionCounts.Subjective)

				// Verify paper permissions
				var permissions models.PaperPermission
				err := db.DB.Where("paper_id = ? AND user_id = ?", resp.PaperId, userID).Take(&permissions).Error
				require.NoError(t, err)
				assert.True(t, permissions.CanWrite(), "User should have write access to the paper")
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get shared paper",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, 2)

				permissions := models.PaperPermission{
					UserID:  userID,
					PaperID: paper.ID,
				}
				permissions.SetRead()
				err := db.DB.Create(&permissions).Error
				require.NoError(t, err)

				return &paper
			},
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.PaperId)
				verifyPaperPermissions(t, types.PaperID(resp.PaperId), userID, true, false)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Paper not found",
				userID:       userID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "No access to paper",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2 with no sharing
				paper := createTestPaper(t, 2)
				return &paper
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Invalid user ID",
				userID:       0,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.PaperRequest{PaperId: int64(paper.ID)},
				client.GetPaper,
				tt.validate)
		})
	}
}

func TestGetPaperPermissions(t *testing.T) {
	tests := []PaperPermissionsCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Paper owner has full permissions",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
			validate: func(t *testing.T, resp *proto.PaperPermissionsResponse) {
				assert.True(t, resp.CanRead, "Owner should have read permission")
				assert.True(t, resp.CanWrite, "Owner should have write permission")
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - User has read-only permissions",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by another user
				paper := createTestPaper(t, userID+1)

				// Create read-only permissions for test user
				permissions := models.PaperPermission{
					PaperID: paper.ID,
					UserID:  userID,
				}
				permissions.SetRead()
				err := db.DB.Create(&permissions).Error
				require.NoError(t, err)
				return &paper
			},
			validate: func(t *testing.T, resp *proto.PaperPermissionsResponse) {
				assert.True(t, resp.CanRead, "User should have read permission")
				assert.False(t, resp.CanWrite, "User should not have write permission")
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Paper not found",
				userID:       userID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "No permission to access paper",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID+1)
				return &paper
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Invalid user ID",
				userID:       0,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, userID)
				return &paper
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.PaperRequest{PaperId: int64(paper.ID)},
				client.GetPaperPermissions,
				tt.validate)
		})
	}
}

func TestDeletePapers(t *testing.T) {
	tests := []struct {
		BaseTestCase
		setup    func(t *testing.T) []int64
		validate func(t *testing.T)
	}{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Delete multiple owned papers",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) []int64 {
				// Create 3 papers owned by test user
				paper1 := createTestPaper(t, userID)
				paper2 := createTestPaper(t, userID)
				paper3 := createTestPaper(t, userID)
				return []int64{int64(paper1.ID), int64(paper2.ID), int64(paper3.ID)}
			},
			validate: func(t *testing.T) {
				// Verify papers are soft deleted
				var count int64
				err := db.DB.Model(&models.Paper{}).Where("deleted_at IS NULL").Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Mixed permissions - Only delete papers with WRITE access",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) []int64 {
				// Create paper owned by test user
				paper1 := createTestPaper(t, userID)

				// Create paper with another user and grant read-only access to test-user
				paper2 := createTestPaper(t, 2)
				permissions := models.PaperPermission{
					PaperID: paper2.ID,
					UserID:  userID,
				}
				permissions.SetRead()
				err := db.DB.Create(&permissions).Error
				require.NoError(t, err)

				return []int64{int64(paper1.ID), int64(paper2.ID)}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Failure - Empty paper IDs list",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) []int64 {
				return []int64{}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Failure - Non-existent papers",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) []int64 {
				return []int64{999, 1000}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paperIds := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.DeletePapersRequest{PaperIds: paperIds},
				client.DeletePapers,
				func(t *testing.T, _ *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t)
					}
				})
		})
	}
}
