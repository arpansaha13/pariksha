package ptr

import "encoding/json"

// Int32 returns a pointer to the given int or int32 value.
func Int32[T int32 | int](v T) *int32 {
	val := int32(v) // Convert int to int32 if necessary
	return &val
}

// Int64 returns a pointer to the given int64 value
func Int64(v int64) *int64 {
	return &v
}

// String returns a pointer to the given string value
func String(s string) *string {
	return &s
}

// JsonRawMessage returns a pointer to the given byte array
func JsonRawMessage(b []byte) *json.RawMessage {
	r := json.RawMessage(b)
	return &r
}
