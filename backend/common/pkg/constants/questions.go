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
