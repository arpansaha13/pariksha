package constants

type ParameterType int8

const (
	// Primitive
	PARAMETER_TYPE_NUMBER  ParameterType = 1
	PARAMETER_TYPE_STRING  ParameterType = 2
	PARAMETER_TYPE_BOOLEAN ParameterType = 3

	// Composite
	PARAMETER_TYPE_ARRAY ParameterType = 4
)

const (
	QUESTION_TYPE_MCQ        string = "MCQ"
	QUESTION_TYPE_SUBJECTIVE string = "SUBJECTIVE"
	QUESTION_TYPE_CODING     string = "CODING"
)

const (
	MIN_MCQ_OPTIONS_COUNT   int8 = 2
	MAX_MCQ_OPTIONS_COUNT   int8 = 5
	MAX_CODING_INPUTS_COUNT int8 = 5
	MAX_SCORE_PER_QUESTION       = 1000
)
