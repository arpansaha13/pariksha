package interservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

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

func attachRequestIDToCtx(ctx context.Context) context.Context {
	if reqID, ok := logging.GetRequestIDFromContext(ctx); ok && reqID != "" {
		return metadata.AppendToOutgoingContext(ctx, "request_id", reqID)
	}
	return ctx
}

func (ic *Question) GetQuestionsByIDs(ctx context.Context, typedQuestionIDs []types.QuestionID) ([]*proto.QuestionResponse, error) {
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionsByIds(ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions, nil
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

func (ic *Question) GetQuestionHashesByIds(ctx context.Context, typedQuestionIDs []types.QuestionID) ([]string, error) {
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionHashesByIds(ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.Hashes, nil
}

func (ic *Question) GetQuestionByHash(ctx context.Context, questionHash string) (*proto.QuestionResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	resp, err := ic.client.GetQuestionsByHashes(ctx, &proto.QuestionHashesRequest{
		Hashes: []string{questionHash},
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions[0], nil
}

func (ic *Question) GetQuestionTypesByIds(ctx context.Context, typedQuestionIDs []types.QuestionID) ([]proto.QuestionType, error) {
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

	questionTypes := make([]proto.QuestionType, len(typedQuestionIDs))
	for i, q := range resp.Meta {
		questionTypes[i] = q.Type
	}

	return questionTypes, nil
}

func (ic *Question) GetBoilerplate(ctx context.Context, req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	ctx = attachRequestIDToCtx(ctx)
	return ic.client.GetBoilerplate(ctx, req)
}

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
