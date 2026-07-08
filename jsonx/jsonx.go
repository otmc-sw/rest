/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package jsonx

import (
	"database/sql"
	"encoding/json"
	"strings"
)

// Marshal marshals data to JSON without HTML escaping
func Marshal(data interface{}) ([]byte, error) {
	var sb strings.Builder
	encoder := json.NewEncoder(&sb)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return []byte(strings.TrimRight(sb.String(), "\n")), nil
}

// MarshalToString marshals data to JSON string without HTML escaping
func MarshalToString(data interface{}) (string, error) {
	b, err := Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Unmarshal unmarshals JSON data
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// UnmarshalString unmarshals JSON string
func UnmarshalString(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// SQL converts raw JSON to sql.NullString
func SQL(raw []byte) sql.NullString {
	if len(raw) == 0 {
		return sql.NullString{Valid: false}
	}
	if json.Valid(raw) {
		return sql.NullString{
			String: string(raw),
			Valid:  true,
		}
	}
	return sql.NullString{Valid: false}
}

// SQLString converts JSON string to sql.NullString
func SQLString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	if json.Valid([]byte(s)) {
		return sql.NullString{
			String: s,
			Valid:  true,
		}
	}
	return sql.NullString{Valid: false}
}

// Valid checks if data is valid JSON
func Valid(data []byte) bool {
	return json.Valid(data)
}

// ValidString checks if string is valid JSON
func ValidString(s string) bool {
	return json.Valid([]byte(s))
}

// ParseJSON parses JSON string to interface{}
func ParseJSON(s string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ParseJSONOrNull parses JSON string to interface{} or returns nil if invalid
func ParseJSONOrNull(s string) interface{} {
	if s == "" {
		return nil
	}
	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return s
	}
	return result
}
