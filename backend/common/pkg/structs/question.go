package structs

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
)

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

// Validate validates mcq question data
func (mcq *MCQQuestion) Validate() error {
	if strings.TrimSpace(mcq.Statement) == "" {
		return status.Error(codes.InvalidArgument, "question statement cannot be empty")
	}
	if len(mcq.Options) < int(constants.MIN_MCQ_OPTIONS_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("MCQ questions must have at least %d options", constants.MIN_MCQ_OPTIONS_COUNT))
	}
	if len(mcq.Options) > int(constants.MAX_MCQ_OPTIONS_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("MCQ questions cannot have more than %d options", constants.MAX_MCQ_OPTIONS_COUNT))
	}
	return nil
}

type SubjectiveQuestion struct {
	Statement string `json:"statement"`
}

// Validate validates subjective question data
func (subjective *SubjectiveQuestion) Validate() error {
	if strings.TrimSpace(subjective.Statement) == "" {
		return status.Error(codes.InvalidArgument, "question statement cannot be empty")
	}
	return nil
}

type InputDefinition struct {
	VariableName string                 `json:"variable_name"`
	Type         proto.ParameterType    `json:"type"`
	Items        []*proto.ParameterItem `json:"items,omitempty"`
}

func (def *InputDefinition) Validate() error {
	if strings.TrimSpace(def.VariableName) == "" {
		return status.Error(codes.InvalidArgument, "variable name is required for input definition")
	}

	switch def.Type {
	case proto.ParameterType_ARRAY:
		if def.Items == nil || len(def.Items) != 1 {
			return status.Error(codes.InvalidArgument, "array input definition must have exactly one item")
		}
		if def.Items[0].PropertyName != nil {
			return status.Error(codes.InvalidArgument, "array input definition cannot have a property name")
		}
		// Validate item type is primitive
		switch def.Items[0].Type {
		case proto.ParameterType_NUMBER, proto.ParameterType_STRING, proto.ParameterType_BOOLEAN:
			// Valid primitive type
		default:
			return status.Error(codes.InvalidArgument, "array items must have primitive types")
		}
	case proto.ParameterType_NUMBER, proto.ParameterType_STRING, proto.ParameterType_BOOLEAN:
		if def.Items != nil {
			return status.Error(codes.InvalidArgument, "primitive input definition cannot have items")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid input definition type")
	}

	return nil
}

type OutputDefinition struct {
	Type  proto.ParameterType    `json:"type"`
	Items []*proto.ParameterItem `json:"items,omitempty"`
}

func (def *OutputDefinition) Validate() error {
	switch def.Type {
	case proto.ParameterType_ARRAY:
		if def.Items == nil || len(def.Items) != 1 {
			return status.Error(codes.InvalidArgument, "array output definition must have exactly one item")
		}
		// Validate item type is primitive
		switch def.Items[0].Type {
		case proto.ParameterType_NUMBER, proto.ParameterType_STRING, proto.ParameterType_BOOLEAN:
			// Valid primitive type
		default:
			return status.Error(codes.InvalidArgument, "array output items must have primitive types")
		}
	case proto.ParameterType_NUMBER, proto.ParameterType_STRING, proto.ParameterType_BOOLEAN:
		if def.Items != nil {
			return status.Error(codes.InvalidArgument, "primitive output definition cannot have items")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid output definition type")
	}

	return nil
}

type CodingQuestion struct {
	Title            string            `json:"title"`
	Statement        string            `json:"statement"`
	InputDefinitions []InputDefinition `json:"input_definitions"`
	OutputDefinition OutputDefinition  `json:"output_definition"`
}

// Validate validates coding question data
func (coding *CodingQuestion) Validate() error {

	// CodingQuestionData validates coding question data
	if strings.TrimSpace(coding.Title) == "" {
		return status.Error(codes.InvalidArgument, "question title cannot be empty")
	}
	if strings.TrimSpace(coding.Statement) == "" {
		return status.Error(codes.InvalidArgument, "question statement cannot be empty")
	}
	if len(coding.InputDefinitions) == 0 {
		return status.Error(codes.InvalidArgument, "coding question must have input definitions")
	}
	if len(coding.InputDefinitions) > int(constants.MAX_CODING_INPUTS_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("Number of inputs cannot be more than %d", constants.MAX_CODING_INPUTS_COUNT))
	}

	if err := coding.OutputDefinition.Validate(); err != nil {
		return err
	}

	for _, def := range coding.InputDefinitions {
		if err := def.Validate(); err != nil {
			return err
		}
	}

	return nil
}
