package ptr

import "encoding/json"

// Int32 returns a pointer to the given int32 value
func Int32(v int32) *int32 {
	return &v
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
