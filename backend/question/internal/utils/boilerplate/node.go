package boilerplate

import (
	"strings"

	"pariksha/question/internal/structs"
)

// generateNodeBoilerplate generates Node.js boilerplate code
func generateNodeBoilerplate(inputs []structs.InputDefinition) string {
	params := make([]string, len(inputs))
	for i, input := range inputs {
		params[i] = input.VariableName
	}
	return "function solve(" + strings.Join(params, ", ") + ") {\n\n}"
}
