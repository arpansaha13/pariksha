package interservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
)

type Question struct {
	ctx    context.Context
	conn   *grpc.ClientConn
	client proto.QuestionClient
}

func NewQuestion(conn *grpc.ClientConn, client proto.QuestionClient) *Question {
	return &Question{
		conn:   conn,
		client: client,
		ctx:    context.Background(),
	}
}

func (ic *Question) Close() {
	if ic.conn != nil {
		ic.conn.Close()
	}
}

// Question operations

func (ic *Question) CreateQuestion(req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	return ic.client.CreateQuestion(ic.ctx, req)
}

func (ic *Question) UpdateQuestion(req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	return ic.client.UpdateQuestion(ic.ctx, req)
}

func (ic *Question) GetQuestionsMetaByIDs(typedQuestionIDs []types.QuestionID) ([]*proto.QuestionMeta, error) {
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	resp, err := ic.client.GetQuestionsMetaByIds(ic.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}
	return resp.Meta, nil
}

func (ic *Question) GetQuestionIDsByHashes(questionHashes []string) ([]types.QuestionID, error) {
	resp, err := ic.client.GetQuestionIdsByHashes(ic.ctx, &proto.QuestionHashesRequest{
		Hashes: questionHashes,
	})
	if err != nil {
		return nil, err
	}

	typedQuestionIDs := make([]types.QuestionID, len(resp.Ids))
	for i, qid := range resp.Ids {
		typedQuestionIDs[i] = types.QuestionID(qid)
	}

	return typedQuestionIDs, nil
}

// GetQuestionMetaByHash fetches question metadata for a given hash.
func (ic *Question) GetQuestionMetaByHash(questionHash string) (*proto.QuestionMeta, error) {
	resp, err := ic.client.GetQuestionsMetaByHashes(ic.ctx, &proto.QuestionHashesRequest{
		Hashes: []string{questionHash},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Meta) == 0 {
		return &proto.QuestionMeta{}, err
	}

	return resp.Meta[0], nil
}

func (ic *Question) GetQuestionByID(typedQuestionID types.QuestionID) (*proto.QuestionResponse, error) {
	resp, err := ic.client.GetQuestionsByIds(ic.ctx, &proto.QuestionIdsRequest{
		Ids: []int64{int64(typedQuestionID)},
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions[0], err
}

func (ic *Question) GetQuestionByHash(questionHash string) (*proto.QuestionResponse, error) {
	resp, err := ic.client.GetQuestionsByHashes(ic.ctx, &proto.QuestionHashesRequest{
		Hashes: []string{questionHash},
	})

	if err != nil {
		return nil, err
	}
	if len(resp.Questions) == 0 {
		return nil, status.Error(codes.NotFound, "could not find question")
	}

	return resp.Questions[0], nil
}

func (ic *Question) DecQuestionPaperIndegreeByIds(typedQuestionIDs []types.QuestionID) error {
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, q := range typedQuestionIDs {
		questionIDs[i] = int64(q)
	}

	_, err := ic.client.DecQuestionPaperIndegreeByIds(ic.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	return err
}

func (ic *Question) GetBoilerplate(req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return ic.client.GetBoilerplate(ic.ctx, req)
}

func (ic *Question) UpsertTestCases(req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	return ic.client.UpsertTestCases(ic.ctx, req)
}

// CreateCategory creates a new category by calling the gRPC CreateCategory method.
func (ic *Question) CreateCategory(name string) (*proto.CategoryResponse, error) { // Assuming CreateCategory takes an empty request.
	return ic.client.CreateCategory(ic.ctx, &proto.CreateCategoryRequest{
		Name: name,
	})
}

// GetCategoriesByIDs fetches categories by their IDs.
func (ic *Question) GetCategoriesByIDs(typedCategoryIDs []types.CategoryID) ([]*proto.CategoryResponse, error) {
	catIds := make([]int64, len(typedCategoryIDs))
	for i, id := range typedCategoryIDs {
		catIds[i] = int64(id)
	}

	resp, err := ic.client.GetCategoriesByIds(ic.ctx, &proto.CategoryIdsRequest{
		Ids: catIds,
	})
	if err != nil {
		return nil, err
	}

	return resp.Categories, nil
}

// UpdateCategoryName updates the name of a category.
func (ic *Question) UpdateCategoryName(req *proto.UpdateCategoryRequest) (*proto.UpdateCategoryResponse, error) {
	resp, err := ic.client.UpdateCategoryName(ic.ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
