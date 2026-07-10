/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package jsonx

import (
	"database/sql"
	"encoding/json"
	"strings"
)

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

func MarshalToString(data interface{}) (string, error) {
	b, err := Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func UnmarshalString(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

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

func Valid(data []byte) bool {
	return json.Valid(data)
}

func ValidString(s string) bool {
	return json.Valid([]byte(s))
}

func ParseJSON(s string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

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
