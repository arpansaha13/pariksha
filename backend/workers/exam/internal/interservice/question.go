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
	"pariksha/workers/exam/internal/config/env"
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

var IncQuestionExamIndegreeByIds = incQuestionExamIndegreeByIds

func incQuestionExamIndegreeByIds(typedQuestionIDs []types.QuestionID) error {
	ensureQuestionService()

	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, q := range typedQuestionIDs {
		questionIDs[i] = int64(q)
	}

	_, err := qSvc.client.IncQuestionExamIndegreeByIds(qSvc.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	return err
}

var DecQuestionExamIndegreeByIds = decQuestionExamIndegreeByIds

func decQuestionExamIndegreeByIds(typedQuestionIDs []types.QuestionID) error {
	ensureQuestionService()

	questionIDs := make([]int64, len(typedQuestionIDs))
	for i, q := range typedQuestionIDs {
		questionIDs[i] = int64(q)
	}

	_, err := qSvc.client.DecQuestionExamIndegreeByIds(qSvc.ctx, &proto.QuestionIdsRequest{
		Ids: questionIDs,
	})
	return err
}
