package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// StrictUnmarshal performs JSON unmarshaling with strict field checking using json.Decoder.
// It returns an error mentioning any unknown field encountered.
func StrictUnmarshal(rawMessage []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(rawMessage))
	decoder.DisallowUnknownFields() // Enforce strict field checking

	if err := decoder.Decode(v); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("syntax error at byte offset %d", syntaxErr.Offset)
		case errors.As(err, &unmarshalTypeErr):
			return fmt.Errorf("type error: field %q, value %v", unmarshalTypeErr.Field, unmarshalTypeErr.Value)
		default:
			return fmt.Errorf("JSON unmarshal error: %w", err)
		}
	}

	return nil
}
