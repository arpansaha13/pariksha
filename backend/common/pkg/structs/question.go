package structs

import (
	"pariksha/common/pkg/proto"
)

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

type SubjectiveQuestion struct {
	Statement string `json:"statement"`
}

type InputDefinition struct {
	VariableName string                 `json:"variable_name"`
	Type         proto.ParameterType    `json:"type"`
	Items        []*proto.ParameterItem `json:"items,omitempty"`
}

type OutputDefinition struct {
	Type  proto.ParameterType    `json:"type"`
	Items []*proto.ParameterItem `json:"items,omitempty"`
}

type CodingQuestion struct {
	Title            string            `json:"title"`
	Statement        string            `json:"statement"`
	InputDefinitions []InputDefinition `json:"input_definitions"`
	OutputDefinition OutputDefinition  `json:"output_definition"`
}
