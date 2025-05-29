package constants

type InputType int8

const (
	// Primitive
	INPUT_TYPE_NUMBER  InputType = 1
	INPUT_TYPE_STRING  InputType = 2
	INPUT_TYPE_BOOLEAN InputType = 3

	// Composite
	INPUT_TYPE_ARRAY InputType = 4
)
