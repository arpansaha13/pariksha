package interservice

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/config/env"
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

var GetInputDefinitions = getInputDefinitions

func getInputDefinitions(questionHash string) ([]*proto.InputDefinition, error) {
	ensureQuestionService()

	resp, err := qSvc.client.GetCodingQuestionInputDefinitions(qSvc.ctx, &proto.GetCodingQuestionInputDefinitionsRequest{
		QuestionHash: questionHash,
	})
	if err != nil {
		return nil, err
	}

	return resp.InputDefinitions, err
}
