package interservice

import (
	"context"

	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/config/env"
)

// FetchInputDefinitions returns question details for the given question IDs.
// Returns error if any of the question IDs don't have corresponding questions.
var FetchInputDefinitions = fetchInputDefinitions

// fetchInputDefinitions retrieves the InputDefinitions array
// from the JSONB Question field for a given question ID.
func fetchInputDefinitions(questionHash string) ([]proto.InputDefinition, error) {
	ensurePaperService()

	// Create metadata with engine token
	md := metadata.New(map[string]string{
		constants.X_EXAM_API_TOKEN: env.ENGINE_API_TOKEN,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := pSvc.client.GetCodingQuestionInputDefinitions(ctx, &proto.GetCodingQuestionInputDefinitionsRequest{
		QuestionHash: questionHash,
	})
	if err != nil {
		return nil, err
	}

	inputDefs := make([]proto.InputDefinition, len(resp.InputDefinitions))
	for i, def := range resp.InputDefinitions {
		inputDefs[i] = proto.InputDefinition{
			VariableName: def.VariableName,
			Type:         def.Type,
			Items:        def.Items,
		}
	}
	return inputDefs, nil
}
