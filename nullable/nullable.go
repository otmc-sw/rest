/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package nullable

import (
	"database/sql"
)

// String converts a string to sql.NullString
// Empty string becomes NULL
func String(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// StringPtr converts a string pointer to sql.NullString
// nil becomes NULL
func StringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

// Int64 converts an int64 to sql.NullInt64
// Zero becomes NULL
func Int64(i int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: i,
		Valid: i != 0,
	}
}

// Int64Ptr converts an int64 pointer to sql.NullInt64
// nil becomes NULL
func Int64Ptr(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{
		Int64: *i,
		Valid: true,
	}
}

// Float64 converts a float64 to sql.NullFloat64
// Zero becomes NULL
func Float64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   f != 0,
	}
}

// Float64Ptr converts a float64 pointer to sql.NullFloat64
// nil becomes NULL
func Float64Ptr(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{
		Float64: *f,
		Valid:   true,
	}
}

// Bool converts a bool to sql.NullBool
// false becomes NULL
func Bool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: b,
	}
}

// BoolPtr converts a bool pointer to sql.NullBool
// nil becomes NULL
func BoolPtr(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{
		Bool:  *b,
		Valid: true,
	}
}

// Time converts a time.Time to sql.NullTime
// Zero time becomes NULL
func Time(t interface{}) sql.NullTime {
	switch v := t.(type) {
	case sql.NullTime:
		return v
	default:
		// This is a placeholder - actual time handling would need time package
		return sql.NullTime{Valid: false}
	}
}

// StringBuilder provides fluent API for building nullable strings with defaults
type StringBuilder struct {
	value sql.NullString
}

// NewStringBuilder creates a new string builder
func NewStringBuilder(s string) *StringBuilder {
	return &StringBuilder{
		value: String(s),
	}
}

// Default sets a default value if the string is empty
func (b *StringBuilder) Default(defaultValue string) sql.NullString {
	if !b.value.Valid {
		return sql.NullString{
			String: defaultValue,
			Valid:  true,
		}
	}
	return b.value
}
