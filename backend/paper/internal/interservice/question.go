package interservice

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/paper/internal/config/env"
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

// Question operations
var CreateQuestion = createQuestion

// CreateQuestion creates a new question by calling the gRPC CreateQuestion method.
func createQuestion(req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	ensureQuestionService()
	return qSvc.client.CreateQuestion(qSvc.ctx, req)
}

var UpdateQuestion = updateQuestion

// UpdateQuestion updates an existing question by calling the gRPC UpdateQuestion method.
func updateQuestion(req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	ensureQuestionService()
	return qSvc.client.UpdateQuestion(qSvc.ctx, req)
}

var GetQuestionsMetaByIDs = getQuestionsMetaByIDs

// GetQuestionsMetaByIDs fetches question metadata for given IDs.
func getQuestionsMetaByIDs(typedQuestionIDs []types.QuestionID) ([]*proto.QuestionMeta, error) {
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
	return resp.Meta, nil
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

	return typedQuestionIDs, err
}

var GetQuestionMetaByHash = getQuestionMetaByHash

// GetQuestionMetaByHash fetches question metadata for a given hash.
func getQuestionMetaByHash(questionHash string) (*proto.QuestionMeta, error) {
	ensureQuestionService()

	resp, err := qSvc.client.GetQuestionsMetaByHashes(qSvc.ctx, &proto.QuestionHashesRequest{
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

var GetQuestionByID = getQuestionByID

func getQuestionByID(typedQuestionID types.QuestionID) (*proto.QuestionResponse, error) {
	resp, err := qSvc.client.GetQuestionsByIds(qSvc.ctx, &proto.QuestionIdsRequest{
		Ids: []int64{int64(typedQuestionID)},
	})
	if err != nil {
		return nil, err
	}

	return resp.Questions[0], err
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

var DecQuestionPaperIndegreeByIds = decQuestionPaperIndegreeByIds

func decQuestionPaperIndegreeByIds(typedQuestionIDs []types.QuestionID) error {
	ensureQuestionService()

	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, q := range typedQuestionIDs {
		questionIDs[i] = int64(q)
	}

	_, err := qSvc.client.DecQuestionPaperIndegreeByIds(qSvc.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	return err
}

var GetBoilerplate = getBoilerplate

func getBoilerplate(req *proto.GetBoilerplateRequest) (*proto.BoilerplateResponse, error) {
	ensureQuestionService()
	return qSvc.client.GetBoilerplate(qSvc.ctx, req)
}

var UpsertTestCases = upsertTestCases

func upsertTestCases(req *proto.UpsertTestCasesRequest) (*emptypb.Empty, error) {
	ensureQuestionService()
	return qSvc.client.UpsertTestCases(qSvc.ctx, req)
}

// CreateCategory creates a new category by calling the gRPC CreateCategory method.
var CreateCategory = createCategory

func createCategory(name string) (*proto.CategoryResponse, error) {
	ensureQuestionService()
	// Assuming CreateCategory takes an empty request.
	return qSvc.client.CreateCategory(qSvc.ctx, &proto.CreateCategoryRequest{
		Name: name,
	})
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

// UpdateCategoryName updates the name of a category.
var UpdateCategoryName = updateCategoryName

func updateCategoryName(req *proto.UpdateCategoryRequest) (*proto.UpdateCategoryResponse, error) {
	ensureQuestionService()

	resp, err := qSvc.client.UpdateCategoryName(qSvc.ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
