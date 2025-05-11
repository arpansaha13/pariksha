package tests

import (
	"pariksha/common/pkg/models"
	"testing"

	"google.golang.org/grpc/codes"
)

// TestParticipantData represents test data for creating exam participants
type TestParticipantData struct {
	UserID int64
	Status int16
}

// BaseTestCase contains common fields used across test cases
type BaseTestCase struct {
	name         string
	metadata     map[string]string
	expectedCode codes.Code
	userID       int64
}

// ParticipantTestCase represents a test case for participant-related operations
type ParticipantTestCase[T any] struct {
	BaseTestCase
	setup    func(t *testing.T) *models.ExamParticipant
	validate func(t *testing.T, resp T)
}

// ExamTestCase represents a test case for exam-related operations
type ExamTestCase[T any] struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Exam, int64)
	validate func(t *testing.T, resp T)
}

// ParticipantRequestTestCase represents a test case for operations requiring both participant and request
type ParticipantRequestTestCase[T any, R any] struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.ExamParticipant, R)
	validate func(t *testing.T, resp T)
}

// ExamParticipantRequestTestCase represents a test case for exam participant operations with a request
type ExamParticipantRequestTestCase[T any, R any] struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Exam
	request  R
	validate func(t *testing.T, examID int64, resp T)
}

// ExamParticipantPairTestCase represents a test case returning both exam and participant
type ExamParticipantPairTestCase struct {
	BaseTestCase
	setup    func(t *testing.T) (*models.Exam, *models.ExamParticipant)
	validate func(t *testing.T, examID, participantID int64)
}

// ExamRequestTestCase represents a test case with a request
type ExamRequestTestCase[T any, R any] struct {
	BaseTestCase
	request  R
	validate func(t *testing.T, resp T)
}

// ExamStartTestCase represents a test case for starting an exam
type ExamStartTestCase[T any] struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Exam
	duration int32
	validate func(t *testing.T, exam T)
}

// ExamSetupTestCase represents a basic test case returning an exam
type ExamSetupTestCase[T any] struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp T)
}
