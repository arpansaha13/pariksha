package interservice

import (
	"context"

	"google.golang.org/grpc"

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

func (ic *Question) GetQuestionsByIDs(typedQuestionIDs []types.QuestionID) ([]*proto.QuestionResponse, error) {
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	resp, err := ic.client.GetQuestionsByIds(ic.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions, nil
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

func (ic *Question) GetQuestionHashesByIds(typedQuestionIDs []types.QuestionID) ([]string, error) {
	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	resp, err := ic.client.GetQuestionHashesByIds(ic.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.Hashes, nil
}

func (ic *Question) GetQuestionByHash(questionHash string) (*proto.QuestionResponse, error) {
	resp, err := ic.client.GetQuestionsByHashes(ic.ctx, &proto.QuestionHashesRequest{
		Hashes: []string{questionHash},
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions[0], nil
}

func (ic *Question) GetQuestionTypesByIds(typedQuestionIDs []types.QuestionID) ([]proto.QuestionType, error) {
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

	questionTypes := make([]proto.QuestionType, len(typedQuestionIDs))
	for i, q := range resp.Meta {
		questionTypes[i] = q.Type
	}

	return questionTypes, nil
}

func (ic *Question) GetBoilerplate(req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	return ic.client.GetBoilerplate(ic.ctx, req)
}

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
