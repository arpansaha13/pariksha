package interservice

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/exam/internal/config/env"
)

var (
	qSvc     *questionService
	qSvcOnce sync.Once
)

type questionService struct {
	client proto.QuestionClient
	conn   *grpc.ClientConn
	ctx    context.Context
}

func CloseQuestionConn() {
	if qSvc != nil && qSvc.conn != nil {
		qSvc.conn.Close()
	}
}

func ensureQuestionService() {
	qSvcOnce.Do(func() {
		qSvc = &questionService{}
		addr := fmt.Sprintf("%s:%s", env.QUESTION_SERVER_HOST, env.QUESTION_SERVER_PORT)
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to question service: %v", err)
		}

		qSvc.conn = conn
		qSvc.client = proto.NewQuestionClient(conn)
		qSvc.ctx = context.Background()
	})
}

func init() {
	ensureQuestionService()
}

var GetQuestionsByIDs = getQuestionsByIDs

// GetQuestionsByIDs fetches question metadata for given IDs.
func getQuestionsByIDs(typedQuestionIDs []types.QuestionID) ([]*proto.QuestionResponse, error) {
	ensureQuestionService()

	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	resp, err := qSvc.client.GetQuestionsByIds(qSvc.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions, nil
}

var GetQuestionIDsByHashes = getQuestionIDsByHashes

func getQuestionIDsByHashes(questionHashes []string) ([]types.QuestionID, error) {
	ensureQuestionService()

	resp, err := qSvc.client.GetQuestionIdsByHashes(qSvc.ctx, &proto.QuestionHashesRequest{
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

var GetQuestionHashesByIds = getQuestionHashesByIds

func getQuestionHashesByIds(typedQuestionIDs []types.QuestionID) ([]string, error) {
	ensureQuestionService()

	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	resp, err := qSvc.client.GetQuestionHashesByIds(qSvc.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.Hashes, nil
}

var GetQuestionByHash = getQuestionByHash

func getQuestionByHash(questionHash string) (*proto.QuestionResponse, error) {
	ensureQuestionService()

	resp, err := qSvc.client.GetQuestionsByHashes(qSvc.ctx, &proto.QuestionHashesRequest{
		Hashes: []string{questionHash},
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions[0], nil
}

var GetQuestionTypesByIds = getQuestionTypesByIds

func getQuestionTypesByIds(typedQuestionIDs []types.QuestionID) ([]proto.QuestionType, error) {
	ensureQuestionService()

	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, qid := range typedQuestionIDs {
		questionIDs[i] = int64(qid)
	}

	resp, err := qSvc.client.GetQuestionsMetaByIds(qSvc.ctx, &proto.QuestionIdsRequest{
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

var GetBoilerplate = getBoilerplate

func getBoilerplate(req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	ensureQuestionService()
	return qSvc.client.GetBoilerplate(qSvc.ctx, req)
}

// GetCategoriesByIDs fetches categories by their IDs.
var GetCategoriesByIDs = getCategoriesByIDs

func getCategoriesByIDs(typedCategoryIDs []types.CategoryID) ([]*proto.CategoryResponse, error) {
	ensureQuestionService()

	catIds := make([]int64, len(typedCategoryIDs))
	for i, id := range typedCategoryIDs {
		catIds[i] = int64(id)
	}

	resp, err := qSvc.client.GetCategoriesByIds(qSvc.ctx, &proto.CategoryIdsRequest{
		Ids: catIds,
	})
	if err != nil {
		return nil, err
	}

	return resp.Categories, nil
}
