package structs

import "pariksha/common/pkg/constants"

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

type SubjectiveQuestion struct {
	Statement string `json:"statement"`
}

type ParameterItem struct {
	PropertyName *string                 `json:"property_name,omitempty"`
	Type         constants.ParameterType `json:"type"`
}

type InputDefinition struct {
	VariableName string                  `json:"variable_name"`
	Type         constants.ParameterType `json:"type"`
	Items        *[]ParameterItem        `json:"items,omitempty"`
}

type OutputDefinition struct {
	Type  constants.ParameterType `json:"type"`
	Items *[]ParameterItem        `json:"items,omitempty"`
}

type CodingQuestion struct {
	Title            string            `json:"title"`
	Statement        string            `json:"statement"`
	InputDefinitions []InputDefinition `json:"input_definitions"`
	OutputDefinition OutputDefinition  `json:"output_definition"`
}
