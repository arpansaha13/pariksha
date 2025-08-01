package interservice

import (
	"context"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
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

func (ic *Question) GetInputDefinitions(questionHash string) ([]*proto.InputDefinition, error) {
	resp, err := ic.client.GetCodingQuestionInputDefinitions(ic.ctx, &proto.GetCodingQuestionInputDefinitionsRequest{
		QuestionHash: questionHash,
	})
	if err != nil {
		return nil, err
	}

	return resp.InputDefinitions, nil
}
