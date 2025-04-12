package utils

// Int32 returns a pointer to the given int32 value
func Int32(v int32) *int32 {
	return &v
}

// String returns a pointer to the given string value
func String(s string) *string {
	return &s
}
