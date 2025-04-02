package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/auth/internal/config/db"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.User
		makeRequest  func(user *models.User) *proto.GetUserRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, user *models.User, resp *proto.UserProfileResponse)
	}{
		{
			name: "Success - Get existing user",
			setup: func(t *testing.T) *models.User {
				user := createTestUser(t, testVerifiedEmail, true)
				user.FirstName.String = "John"
				user.FirstName.Valid = true
				user.LastName.String = "Doe"
				user.LastName.Valid = true
				db.DB.Save(&user)
				return &user
			},
			makeRequest: func(user *models.User) *proto.GetUserRequest {
				return &proto.GetUserRequest{
					UserId: int32(user.ID),
				}
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, user *models.User, resp *proto.UserProfileResponse) {
				assert.Equal(t, int32(user.ID), resp.Id)
				assert.Equal(t, user.Username, resp.Username)
				assert.Equal(t, user.Email, resp.Email)
				assert.Equal(t, user.FirstName.String, resp.FirstName)
				assert.Equal(t, user.LastName.String, resp.LastName)
			},
		},
		{
			name: "User not found",
			makeRequest: func(user *models.User) *proto.GetUserRequest {
				return &proto.GetUserRequest{
					UserId: 999,
				}
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user *models.User
			if tt.setup != nil {
				user = tt.setup(t)
			}

			req := tt.makeRequest(user)
			resp, err := client.GetUser(context.Background(), req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, user, resp)
				}
			}

			clearTables(t)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.User
		makeRequest  func(user *models.User) *proto.UpdateUserRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, resp *proto.UpdateUserResponse)
	}{
		{
			name: "Success - Update all fields",
			setup: func(t *testing.T) *models.User {
				user := createTestUser(t, testVerifiedEmail, true)
				return &user
			},
			makeRequest: func(user *models.User) *proto.UpdateUserRequest {
				return &proto.UpdateUserRequest{
					UserId:    int32(user.ID),
					Username:  strPtr("newusername"),
					FirstName: strPtr("John"),
					LastName:  strPtr("Doe"),
				}
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UpdateUserResponse) {
				assert.Equal(t, "newusername", resp.User.Username)
				assert.Equal(t, "John", resp.User.FirstName)
				assert.Equal(t, "Doe", resp.User.LastName)
				assert.Empty(t, resp.NotUpdatedFields)

				// Verify database update
				var user models.User
				err := db.DB.First(&user, resp.User.Id).Error
				assert.NoError(t, err)
				assert.Equal(t, "newusername", user.Username)
				assert.Equal(t, "John", user.FirstName.String)
				assert.Equal(t, "Doe", user.LastName.String)
			},
		},
		{
			name: "Username already taken",
			setup: func(t *testing.T) *models.User {
				user1 := createTestUser(t, "user1@example.com", true)
				createTestUser(t, "user2@example.com", true)
				return &user1
			},
			makeRequest: func(user *models.User) *proto.UpdateUserRequest {
				return &proto.UpdateUserRequest{
					UserId:   int32(user.ID),
					Username: strPtr("user2"),
				}
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UpdateUserResponse) {
				assert.Contains(t, resp.NotUpdatedFields, "username")
				assert.Equal(t, "username is already taken", resp.NotUpdatedFields["username"])
			},
		},
		{
			name: "User not found",
			makeRequest: func(user *models.User) *proto.UpdateUserRequest {
				return &proto.UpdateUserRequest{
					UserId:    999,
					Username:  strPtr("newname"),
					FirstName: strPtr("John"),
				}
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "No fields to update",
			setup: func(t *testing.T) *models.User {
				user := createTestUser(t, testVerifiedEmail, true)
				return &user
			},
			makeRequest: func(user *models.User) *proto.UpdateUserRequest {
				return &proto.UpdateUserRequest{
					UserId: int32(user.ID),
				}
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UpdateUserResponse) {
				assert.NotNil(t, resp.User)
				assert.Empty(t, resp.NotUpdatedFields)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user *models.User
			if tt.setup != nil {
				user = tt.setup(t)
			}

			req := tt.makeRequest(user)
			resp, err := client.UpdateUser(context.Background(), req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, resp)
				}
			}

			clearTables(t)
		})
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
