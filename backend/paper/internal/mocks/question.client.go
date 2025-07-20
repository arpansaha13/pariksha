package mocks

import (
	"context"

	"pariksha/common/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type QuestionClient struct{}

// _________________________QUESTION OPERATIONS________________________

func (m *QuestionClient) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest, opts ...grpc.CallOption) (*proto.CreateQuestionResponse, error) {
	return &proto.CreateQuestionResponse{}, nil
}

func (m *QuestionClient) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest, opts ...grpc.CallOption) (*proto.UpdateQuestionResponse, error) {
	for _, q := range questionMap {
		if q.Hash == req.Hash {
			return &proto.UpdateQuestionResponse{
				Id:   q.Id,
				Hash: q.Hash,
			}, nil
		}
	}
	// Return empty response if no match found
	return &proto.UpdateQuestionResponse{}, nil
}

func (m *QuestionClient) GetQuestionsByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*proto.GetQuestionsResponse, error) {
	var questions []*proto.QuestionResponse
	for _, id := range req.Ids {
		if q, exists := questionMap[id]; exists {
			questions = append(questions, q)
		}
	}
	return &proto.GetQuestionsResponse{Questions: questions}, nil
}

func (m *QuestionClient) GetQuestionsMetaByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*proto.QuestionsMetaResponse, error) {
	var meta []*proto.QuestionMeta
	for _, id := range req.Ids {
		if q, exists := questionMap[id]; exists {
			meta = append(meta, &proto.QuestionMeta{
				Id:          q.Id,
				Hash:        q.Hash,
				RawQuestion: q.RawQuestion,
				Type:        q.Type,
			})
		}
	}
	return &proto.QuestionsMetaResponse{Meta: meta}, nil
}

func (m *QuestionClient) GetQuestionHashesByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*proto.GetQuestionHashesByIdsResponse, error) {
	var hashes []string
	for _, id := range req.Ids {
		if q, exists := questionMap[id]; exists {
			hashes = append(hashes, q.Hash)
		}
	}
	return &proto.GetQuestionHashesByIdsResponse{Hashes: hashes}, nil
}

func (m *QuestionClient) GetQuestionsByHashes(ctx context.Context, req *proto.QuestionHashesRequest, opts ...grpc.CallOption) (*proto.GetQuestionsResponse, error) {
	var questions []*proto.QuestionResponse
	for _, hash := range req.Hashes {
		for _, q := range questionMap {
			if q.Hash == hash {
				questions = append(questions, q)
				break
			}
		}
	}
	return &proto.GetQuestionsResponse{Questions: questions}, nil
}

func (m *QuestionClient) GetQuestionsMetaByHashes(ctx context.Context, req *proto.QuestionHashesRequest, opts ...grpc.CallOption) (*proto.QuestionsMetaResponse, error) {
	var meta []*proto.QuestionMeta
	for _, hash := range req.Hashes {
		for _, q := range questionMap {
			if q.Hash == hash {
				meta = append(meta, &proto.QuestionMeta{
					Id:          q.Id,
					Hash:        q.Hash,
					RawQuestion: q.RawQuestion,
					Type:        q.Type,
				})
				break
			}
		}
	}
	return &proto.QuestionsMetaResponse{Meta: meta}, nil
}

func (m *QuestionClient) GetQuestionIdsByHashes(ctx context.Context, req *proto.QuestionHashesRequest, opts ...grpc.CallOption) (*proto.GetQuestionIdsByHashesResponse, error) {
	var ids []int64
	for _, hash := range req.Hashes {
		for id, q := range questionMap {
			if q.Hash == hash {
				ids = append(ids, id)
				break
			}
		}
	}
	return &proto.GetQuestionIdsByHashesResponse{Ids: ids}, nil
}

func (m *QuestionClient) DecQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) DecQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) IncQuestionPaperIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) IncQuestionExamIndegreeByIds(ctx context.Context, req *proto.QuestionIdsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *QuestionClient) GetCodingQuestionInputDefinitions(ctx context.Context, req *proto.GetCodingQuestionInputDefinitionsRequest, opts ...grpc.CallOption) (*proto.GetCodingQuestionInputDefinitionsResponse, error) {
	return &proto.GetCodingQuestionInputDefinitionsResponse{}, nil
}

// _______________________BOILERPLATE OPERATIONS_______________________

func (m *QuestionClient) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest, opts ...grpc.CallOption) (*proto.BoilerplateResponse, error) {
	return &proto.BoilerplateResponse{}, nil
}

// ________________________TESTCASE OPERATIONS_________________________

func (m *QuestionClient) UpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// _________________________CATEGORY OPERATIONS________________________

// CreateCategory will return a static response with question_id=6.
// Make sure question_id=6 is not already created before running the test.
func (m *QuestionClient) CreateCategory(ctx context.Context, req *proto.CreateCategoryRequest, opts ...grpc.CallOption) (*proto.CategoryResponse, error) {
	return &proto.CategoryResponse{
		Id:   6,
		Name: categoryMap[6].Name,
	}, nil
}

func (m *QuestionClient) GetCategoriesByIds(ctx context.Context, req *proto.CategoryIdsRequest, opts ...grpc.CallOption) (*proto.GetCategoriesResponse, error) {
	var categories []*proto.CategoryResponse
	for _, id := range req.Ids {
		if category, exists := categoryMap[id]; exists {
			categories = append(categories, category)
		}
	}
	return &proto.GetCategoriesResponse{Categories: categories}, nil
}

func (m *QuestionClient) UpdateCategoryName(ctx context.Context, req *proto.UpdateCategoryRequest, opts ...grpc.CallOption) (*proto.UpdateCategoryResponse, error) {
	return &proto.UpdateCategoryResponse{
		Id: req.Id,
	}, nil
}
