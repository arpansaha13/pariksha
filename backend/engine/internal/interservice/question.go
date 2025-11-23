package interservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
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

// GetInputDefinitions fetches input definitions for a coding question. The ctx passed
// will be used to propagate request metadata (request id) to the question service.
func (ic *Question) GetInputDefinitions(ctx context.Context, questionHash string) ([]*proto.InputDefinition, error) {
	// extract request id from ctx and append to outgoing metadata if present
	if reqID, ok := logging.GetRequestIDFromContext(ctx); ok && reqID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "request_id", reqID)
	}

	resp, err := ic.client.GetCodingQuestionInputDefinitions(ctx, &proto.GetCodingQuestionInputDefinitionsRequest{
		QuestionHash: questionHash,
	})
	if err != nil {
		return nil, err
	}

	return resp.InputDefinitions, nil
}
