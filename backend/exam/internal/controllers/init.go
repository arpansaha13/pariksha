package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/repositories"
	"pariksha/exam/internal/services"
)

type ExamServer struct {
	proto.UnimplementedExamServer
}

var (
	examCtrl        *Exam
	answerCtrl      *Answer
	questionCtrl    *Question
	categoryCtrl    *Category
	participantCtrl *Participant
	evaluationCtrl  *Evaluation
	resultCtrl      *Result
)

// Init sets up all handler dependencies.
// Must be called before using any handlers.
func Init() {
	// Initialize repositories
	examRepo := repositories.NewExam(db.DB)
	answerRepo := repositories.NewAnswer(db.DB)
	questionRepo := repositories.NewQuestion(db.DB)
	categoryRepo := repositories.NewCategory(db.DB)
	participantRepo := repositories.NewParticipant(db.DB)
	permissionRepo := repositories.NewPermission(db.DB)

	// Initialize services
	examSvc := services.NewExam(examRepo, participantRepo, permissionRepo)
	answerSvc := services.NewAnswer(answerRepo, questionRepo, participantRepo)
	categorySvc := services.NewCategory(categoryRepo)
	questionSvc := services.NewQuestion(questionRepo)
	participantSvc := services.NewParticipant(participantRepo, permissionRepo)
	evaluationSvc := services.NewEvaluation(answerRepo, participantRepo)
	resultSvc := services.NewResult(participantRepo, answerRepo)

	// Initialize controllers
	examCtrl = NewExam(examSvc)
	answerCtrl = NewAnswer(answerSvc)
	categoryCtrl = NewCategory(categorySvc)
	questionCtrl = NewQuestion(questionSvc)
	participantCtrl = NewParticipant(participantSvc)
	evaluationCtrl = NewEvaluation(evaluationSvc)
	resultCtrl = NewResult(resultSvc)
}

// __________________________EXAM HANDLERS__________________________
func (s *ExamServer) GetUserExams(ctx context.Context, req *emptypb.Empty) (*proto.ExamList, error) {
	return examCtrl.HandleGetUserExams(ctx, req)
}

func (s *ExamServer) CreateExam(ctx context.Context, req *proto.CreateExamRequest) (*proto.ExamResponse, error) {
	return examCtrl.HandleCreateExam(ctx, req)
}

func (s *ExamServer) UpdateExam(ctx context.Context, req *proto.UpdateExamRequest) (*proto.ExamResponse, error) {
	return examCtrl.HandleUpdateExam(ctx, req)
}

func (s *ExamServer) StartExam(ctx context.Context, req *proto.StartExamRequest) (*emptypb.Empty, error) {
	return examCtrl.HandleStartExam(ctx, req)
}

func (s *ExamServer) EndExam(ctx context.Context, req *proto.EndExamRequest) (*emptypb.Empty, error) {
	return examCtrl.HandleEndExam(ctx, req)
}

func (s *ExamServer) GetExam(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResponse, error) {
	return examCtrl.HandleGetExam(ctx, req)
}

func (s *ExamServer) GetExamPermission(ctx context.Context, req *proto.ExamRequest) (*proto.ExamPermissionResponse, error) {
	return examCtrl.HandleGetExamPermission(ctx, req)
}

func (s *ExamServer) DeleteExams(ctx context.Context, req *proto.DeleteExamsRequest) (*emptypb.Empty, error) {
	return examCtrl.HandleDeleteExams(ctx, req)
}

// ______________________PARTICIPANT HANDLERS_______________________
func (s *ExamServer) GetExamParticipants(ctx context.Context, req *proto.ExamRequest) (*proto.ParticipantList, error) {
	return participantCtrl.HandleGetExamParticipants(ctx, req)
}

func (s *ExamServer) AddExamParticipant(ctx context.Context, req *proto.AddParticipantRequest) (*proto.ParticipantResponse, error) {
	return participantCtrl.HandleAddExamParticipant(ctx, req)
}

func (s *ExamServer) RemoveExamParticipant(ctx context.Context, req *proto.RemoveParticipantRequest) (*emptypb.Empty, error) {
	return participantCtrl.HandleRemoveExamParticipant(ctx, req)
}

func (s *ExamServer) GetExamParticipant(ctx context.Context, req *proto.GetExamParticipantRequest) (*proto.GetExamParticipantResponse, error) {
	return participantCtrl.HandleGetExamParticipant(ctx, req)
}

func (s *ExamServer) GetParticipantById(ctx context.Context, req *proto.ParticipantRequest) (*proto.ParticipantResponse, error) {
	return participantCtrl.HandleGetParticipantById(ctx, req)
}

// _________________________ANSWER HANDLERS_________________________
func (s *ExamServer) GetParticipantAnswers(ctx context.Context, req *proto.ParticipantRequest) (*proto.AnswerList, error) {
	return answerCtrl.HandleGetParticipantAnswers(ctx, req)
}

func (s *ExamServer) GetAnswerForExam(ctx context.Context, req *proto.GetAnswerRequest) (*proto.AnswerMinimalResponse, error) {
	return answerCtrl.HandleGetAnswerForExam(ctx, req)
}

func (s *ExamServer) UpsertAnswer(ctx context.Context, req *proto.UpsertAnswersRequest) (*proto.UpsertAnswersResponse, error) {
	return answerCtrl.HandleUpsertAnswer(ctx, req)
}

// ________________________QUESTION HANDLERS________________________
func (s *ExamServer) GetExamQuestions(ctx context.Context, req *proto.ExamRequest) (*proto.ExamQuestionsResponse, error) {
	return questionCtrl.HandleGetExamQuestions(ctx, req)
}

func (s *ExamServer) GetExamQuestion(ctx context.Context, req *proto.ExamQuestionRequest) (*proto.ExamQuestionResponse, error) {
	return questionCtrl.HandleGetExamQuestion(ctx, req)
}

// ________________________CATEGORY HANDLERS________________________
func (s *ExamServer) GetExamCategories(ctx context.Context, req *proto.ExamRequest) (*proto.ExamCategoriesResponse, error) {
	return categoryCtrl.HandleGetExamCategories(ctx, req)
}

// _______________________EVALUATION HANDLERS_______________________
func (s *ExamServer) GetAnswerForEvaluation(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.AnswerMinimalResponse, error) {
	return evaluationCtrl.HandleGetAnswerForEvaluation(ctx, req)
}

func (s *ExamServer) GetAnswerEvaluationData(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	return evaluationCtrl.HandleGetAnswerEvaluationData(ctx, req)
}

func (s *ExamServer) UpdateAnswerForEvaluation(ctx context.Context, req *proto.UpdateAnswerRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	return evaluationCtrl.HandleUpdateAnswerForEvaluation(ctx, req)
}

func (s *ExamServer) MarkParticipantAsEvaluated(ctx context.Context, req *proto.ParticipantRequest) (*proto.EvaluationStatusResponse, error) {
	return evaluationCtrl.HandleMarkParticipantAsEvaluated(ctx, req)
}

// _________________________RESULT HANDLERS_________________________
func (s *ExamServer) GetExamResults(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResultsResponse, error) {
	return resultCtrl.HandleGetExamResults(ctx, req)
}
