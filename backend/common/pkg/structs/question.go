package structs

import "pariksha/common/pkg/constants"

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

type SubjectiveQuestion struct {
	Statement string `json:"statement"`
}

type InputDefinitionItem struct {
	PropertyName *string             `json:"property_name,omitempty"`
	Type         constants.InputType `json:"type"`
}

type InputDefinition struct {
	VariableName *string                `json:"variable_name,omitempty"`
	Type         constants.InputType    `json:"type"`
	Items        *[]InputDefinitionItem `json:"items,omitempty"`
}

type CodingQuestionExample struct {
	Input       string  `json:"input"`
	Output      string  `json:"output"`
	Explanation *string `json:"explanation,omitempty"`
}

type CodingQuestion struct {
	Title            string                  `json:"title"`
	Statement        string                  `json:"statement"`
	InputDefinitions []InputDefinition       `json:"input_definitions"`
	Examples         []CodingQuestionExample `json:"examples,omitempty"`
}
