package mocks

import (
	"context"

	"pariksha/common/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type QuestionClient struct{}

func (m *QuestionClient) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest, opts ...grpc.CallOption) (*proto.CreateQuestionResponse, error) {
	return &proto.CreateQuestionResponse{}, nil
}

func (m *QuestionClient) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest, opts ...grpc.CallOption) (*proto.UpdateQuestionResponse, error) {
	return &proto.UpdateQuestionResponse{}, nil
}

func (m *QuestionClient) GetQuestionsMetaByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*proto.QuestionsMetaResponse, error) {
	return &proto.QuestionsMetaResponse{Meta: []*proto.QuestionMeta{}}, nil
}

func (m *QuestionClient) GetQuestionIdsByHashes(ctx context.Context, req *proto.QuestionHashesRequest, opts ...grpc.CallOption) (*proto.GetQuestionIdsByHashesResponse, error) {
	return &proto.GetQuestionIdsByHashesResponse{Ids: []int64{}}, nil
}

func (m *QuestionClient) GetQuestionsMetaByHashes(ctx context.Context, req *proto.QuestionHashesRequest, opts ...grpc.CallOption) (*proto.QuestionsMetaResponse, error) {
	return &proto.QuestionsMetaResponse{Meta: []*proto.QuestionMeta{}}, nil
}

func (m *QuestionClient) GetQuestionsByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*proto.GetQuestionsResponse, error) {
	return &proto.GetQuestionsResponse{Questions: []*proto.QuestionResponse{{}}}, nil
}

func (m *QuestionClient) GetQuestionsByHashes(ctx context.Context, req *proto.QuestionHashesRequest, opts ...grpc.CallOption) (*proto.GetQuestionsResponse, error) {
	return &proto.GetQuestionsResponse{Questions: []*proto.QuestionResponse{{}}}, nil
}

func (m *QuestionClient) DecQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest, opts ...grpc.CallOption) (*proto.BoilerplateResponse, error) {
	return &proto.BoilerplateResponse{}, nil
}

func (m *QuestionClient) UpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) CreateCategory(ctx context.Context, req *proto.CreateCategoryRequest, opts ...grpc.CallOption) (*proto.CategoryResponse, error) {
	return &proto.CategoryResponse{}, nil
}

func (m *QuestionClient) GetCategoriesByIds(ctx context.Context, req *proto.CategoryIdsRequest, opts ...grpc.CallOption) (*proto.GetCategoriesResponse, error) {
	return &proto.GetCategoriesResponse{Categories: []*proto.CategoryResponse{}}, nil
}

func (m *QuestionClient) UpdateCategoryName(ctx context.Context, req *proto.UpdateCategoryRequest, opts ...grpc.CallOption) (*proto.UpdateCategoryResponse, error) {
	return &proto.UpdateCategoryResponse{}, nil
}

func (m *QuestionClient) GetQuestionHashesByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*proto.GetQuestionHashesByIdsResponse, error) {
	return &proto.GetQuestionHashesByIdsResponse{Hashes: []string{}}, nil
}

func (m *QuestionClient) IncQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) IncQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) DecQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) GetCodingQuestionInputDefinitions(ctx context.Context, req *proto.GetCodingQuestionInputDefinitionsRequest, opts ...grpc.CallOption) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	return &proto.GetCodingQuestionInputDefinitionsResponse{}, nil
}
