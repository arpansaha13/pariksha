package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// getJSONFields extracts a map of JSON field names from a struct type using reflection.
// It returns a set (map with `true` values) of JSON field names.
func getJSONFields(v interface{}) (map[string]bool, error) {
	fields := make(map[string]bool)
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil, errors.New("expected pointer to struct")
	}
	typ := val.Elem().Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name := field.Name
		if tag != "" {
			name = tag
			// Handle tag options (e.g., `json:"name,omitempty"`)
			if commaIdx := indexComma(tag); commaIdx >= 0 {
				name = tag[:commaIdx]
			}
		}
		fields[name] = true
	}
	return fields, nil
}

// indexComma finds the first occurrence of a comma in a string.
// Returns the index of the comma or -1 if not found.
func indexComma(tag string) int {
	for i, r := range tag {
		if r == ',' {
			return i
		}
	}
	return -1
}

// StrictUnmarshal performs JSON unmarshaling with strict field checking.
// It returns an error if the JSON contains fields not defined in the target struct.
func StrictUnmarshal(rawMessage []byte, v interface{}) error {
	allowedFields, err := getJSONFields(v)
	if err != nil {
		return err
	}

	// Decode into map to find all fields
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(rawMessage, &rawFields); err != nil {
		return err
	}

	// Check for unknown fields
	for key := range rawFields {
		if !allowedFields[key] {
			return fmt.Errorf("unknown field: %s", key)
		}
	}

	// Unmarshal into the actual struct
	return json.Unmarshal(rawMessage, v)
}
