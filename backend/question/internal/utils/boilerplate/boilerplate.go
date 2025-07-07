package boilerplate

import (
	"pariksha/common/pkg/constants"
	"pariksha/question/internal/models"
	"pariksha/question/internal/structs"
)

// GenerateBoilerplate generates language-specific boilerplate code based on input/output definitions
func Generate(lang *models.Language, inputs []structs.InputDefinition, output structs.OutputDefinition) string {
	switch lang.Slug {
	case constants.LangNode:
		return generateNodeBoilerplate(inputs)
	default:
		return "" // Unsupported language
	}
}
