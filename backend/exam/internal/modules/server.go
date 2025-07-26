package modules

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/controllers"
)

type examServer struct {
	proto.UnimplementedExamServer

	examCtrl        *controllers.Exam
	answerCtrl      *controllers.Answer
	questionCtrl    *controllers.Question
	categoryCtrl    *controllers.Category
	participantCtrl *controllers.Participant
	evaluationCtrl  *controllers.Evaluation
	resultCtrl      *controllers.Result
}

// __________________________EXAM HANDLERS__________________________
func (s *examServer) GetUserExams(ctx context.Context, req *emptypb.Empty) (*proto.ExamList, error) {
	return s.examCtrl.HandleGetUserExams(ctx, req)
}

func (s *examServer) CreateExam(ctx context.Context, req *proto.CreateExamRequest) (*proto.ExamResponse, error) {
	return s.examCtrl.HandleCreateExam(ctx, req)
}

func (s *examServer) UpdateExam(ctx context.Context, req *proto.UpdateExamRequest) (*proto.ExamResponse, error) {
	return s.examCtrl.HandleUpdateExam(ctx, req)
}

func (s *examServer) StartExam(ctx context.Context, req *proto.StartExamRequest) (*emptypb.Empty, error) {
	return s.examCtrl.HandleStartExam(ctx, req)
}

func (s *examServer) EndExam(ctx context.Context, req *proto.EndExamRequest) (*emptypb.Empty, error) {
	return s.examCtrl.HandleEndExam(ctx, req)
}

func (s *examServer) GetExam(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResponse, error) {
	return s.examCtrl.HandleGetExam(ctx, req)
}

func (s *examServer) GetExamPermission(ctx context.Context, req *proto.ExamRequest) (*proto.ExamPermissionResponse, error) {
	return s.examCtrl.HandleGetExamPermission(ctx, req)
}

func (s *examServer) DeleteExams(ctx context.Context, req *proto.DeleteExamsRequest) (*emptypb.Empty, error) {
	return s.examCtrl.HandleDeleteExams(ctx, req)
}

// ______________________PARTICIPANT HANDLERS_______________________
func (s *examServer) GetExamParticipants(ctx context.Context, req *proto.ExamRequest) (*proto.ParticipantList, error) {
	return s.participantCtrl.HandleGetExamParticipants(ctx, req)
}

func (s *examServer) AddExamParticipant(ctx context.Context, req *proto.AddParticipantRequest) (*proto.ParticipantResponse, error) {
	return s.participantCtrl.HandleAddExamParticipant(ctx, req)
}

func (s *examServer) RemoveExamParticipant(ctx context.Context, req *proto.RemoveParticipantRequest) (*emptypb.Empty, error) {
	return s.participantCtrl.HandleRemoveExamParticipant(ctx, req)
}

func (s *examServer) GetExamParticipant(ctx context.Context, req *proto.GetExamParticipantRequest) (*proto.GetExamParticipantResponse, error) {
	return s.participantCtrl.HandleGetExamParticipant(ctx, req)
}

func (s *examServer) GetParticipantById(ctx context.Context, req *proto.ParticipantRequest) (*proto.ParticipantResponse, error) {
	return s.participantCtrl.HandleGetParticipantById(ctx, req)
}

// _________________________ANSWER HANDLERS_________________________
func (s *examServer) GetParticipantAnswers(ctx context.Context, req *proto.ParticipantRequest) (*proto.AnswerList, error) {
	return s.answerCtrl.HandleGetParticipantAnswers(ctx, req)
}

func (s *examServer) GetAnswerForExam(ctx context.Context, req *proto.GetAnswerRequest) (*proto.AnswerMinimalResponse, error) {
	return s.answerCtrl.HandleGetAnswerForExam(ctx, req)
}

func (s *examServer) UpsertAnswer(ctx context.Context, req *proto.UpsertAnswersRequest) (*proto.UpsertAnswersResponse, error) {
	return s.answerCtrl.HandleUpsertAnswer(ctx, req)
}

// ________________________QUESTION HANDLERS________________________
func (s *examServer) GetExamQuestions(ctx context.Context, req *proto.ExamRequest) (*proto.ExamQuestionsResponse, error) {
	return s.questionCtrl.HandleGetExamQuestions(ctx, req)
}

func (s *examServer) GetExamQuestion(ctx context.Context, req *proto.ExamQuestionRequest) (*proto.ExamQuestionResponse, error) {
	return s.questionCtrl.HandleGetExamQuestion(ctx, req)
}

// ________________________CATEGORY HANDLERS________________________
func (s *examServer) GetExamCategories(ctx context.Context, req *proto.ExamRequest) (*proto.ExamCategoriesResponse, error) {
	return s.categoryCtrl.HandleGetExamCategories(ctx, req)
}

// _______________________EVALUATION HANDLERS_______________________
func (s *examServer) GetAnswerForEvaluation(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.AnswerMinimalResponse, error) {
	return s.evaluationCtrl.HandleGetAnswerForEvaluation(ctx, req)
}

func (s *examServer) GetAnswerEvaluationData(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	return s.evaluationCtrl.HandleGetAnswerEvaluationData(ctx, req)
}

func (s *examServer) UpdateAnswerForEvaluation(ctx context.Context, req *proto.UpdateAnswerRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	return s.evaluationCtrl.HandleUpdateAnswerForEvaluation(ctx, req)
}

func (s *examServer) MarkParticipantAsEvaluated(ctx context.Context, req *proto.ParticipantRequest) (*proto.EvaluationStatusResponse, error) {
	return s.evaluationCtrl.HandleMarkParticipantAsEvaluated(ctx, req)
}

// _________________________RESULT HANDLERS_________________________
func (s *examServer) GetExamResults(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResultsResponse, error) {
	return s.resultCtrl.HandleGetExamResults(ctx, req)
}
