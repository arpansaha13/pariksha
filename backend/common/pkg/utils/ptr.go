package utils

// Ptr returns a pointer to the provided value of any type.
func Ptr[T any](v T) *T {
	return &v
}
