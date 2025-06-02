package structs

import "pariksha/common/pkg/structs"

// TestCases should not be added during creation
type CodingQuestionOmitTestCases struct {
	Title            string                    `json:"title"`
	Statement        string                    `json:"statement"`
	InputDefinitions []structs.InputDefinition `json:"input_definitions"`
	OutputDefinition structs.OutputDefinition  `json:"output_definition"`
}
