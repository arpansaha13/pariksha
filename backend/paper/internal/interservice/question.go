package interservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
)

type Question struct {
	conn   *grpc.ClientConn
	client proto.QuestionClient
}

func NewQuestion(conn *grpc.ClientConn, client proto.QuestionClient) *Question {
	return &Question{
		conn:   conn,
		client: client,
	}
}

func (ic *Question) Close() {
	if ic.conn != nil {
		ic.conn.Close()
	}
}

// helper to attach request id metadata if present
func attachRequestIDToCtx(ctx context.Context) context.Context {
	if reqID, ok := logging.GetRequestIDFromContext(ctx); ok && reqID != "" {
		return metadata.AppendToOutgoingContext(ctx, "request_id", reqID)
	}
	return ctx
}

// Question operations

func (ic *Question) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	return ic.client.CreateQuestion(ctx, req)
}

func (ic *Question) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	return ic.client.UpdateQuestion(ctx, req)
}

func (ic *Question) GetQuestionsMetaByIDs(ctx context.Context, typedQuestionIDs []types.QuestionID) ([]*proto.QuestionMeta, error) {
	if len(typedQuestionIDs) == 0 {
		return []*proto.QuestionMeta{}, nil
	}
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionsMetaByIds(ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}
	return resp.Meta, nil
}

func (ic *Question) GetQuestionIDsByHashes(ctx context.Context, questionHashes []string) ([]types.QuestionID, error) {
	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionIdsByHashes(ctx, &proto.QuestionHashesRequest{
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
func (ic *Question) GetQuestionMetaByHash(ctx context.Context, questionHash string) (*proto.QuestionMeta, error) {
	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionsMetaByHashes(ctx, &proto.QuestionHashesRequest{
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

func (ic *Question) GetQuestionByID(ctx context.Context, typedQuestionID types.QuestionID) (*proto.QuestionResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionsByIds(ctx, &proto.QuestionIdsRequest{
		Ids: []int64{int64(typedQuestionID)},
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions[0], err
}

func (ic *Question) GetQuestionByHash(ctx context.Context, questionHash string) (*proto.QuestionResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionsByHashes(ctx, &proto.QuestionHashesRequest{
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

func (ic *Question) DecQuestionPaperIndegreeByIds(ctx context.Context, typedQuestionIDs []types.QuestionID) error {
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, q := range typedQuestionIDs {
		questionIDs[i] = int64(q)
	}

	ctx = attachRequestIDToCtx(ctx)
	_, err := ic.client.DecQuestionPaperIndegreeByIds(ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	return err
}

func (ic *Question) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	return ic.client.GetBoilerplate(ctx, req)
}

func (ic *Question) UpsertTestCases(ctx context.Context, req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	ctx = attachRequestIDToCtx(ctx)
	return ic.client.UpsertTestCases(ctx, req)
}

// CreateCategory creates a new category by calling the gRPC CreateCategory method.
func (ic *Question) CreateCategory(ctx context.Context, name string) (*proto.CategoryResponse, error) { // Assuming CreateCategory takes an empty request.
	ctx = attachRequestIDToCtx(ctx)
	return ic.client.CreateCategory(ctx, &proto.CreateCategoryRequest{
		Name: name,
	})
}

// GetCategoriesByIDs fetches categories by their IDs.
func (ic *Question) GetCategoriesByIDs(ctx context.Context, typedCategoryIDs []types.CategoryID) ([]*proto.CategoryResponse, error) {
	catIds := make([]int64, len(typedCategoryIDs))
	for i, id := range typedCategoryIDs {
		catIds[i] = int64(id)
	}

	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetCategoriesByIds(ctx, &proto.CategoryIdsRequest{
		Ids: catIds,
	})
	if err != nil {
		return nil, err
	}

	return resp.Categories, nil
}

// UpdateCategoryName updates the name of a category.
func (ic *Question) UpdateCategoryName(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.UpdateCategoryResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.UpdateCategoryName(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
