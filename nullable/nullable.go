/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package nullable

import (
	"database/sql"
	"time"
)

type NullString struct {
	String string
	Valid  bool
}

type NullInt64 struct {
	Int64 int64
	Valid bool
}

type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

type NullBool struct {
	Bool  bool
	Valid bool
}

type NullTime struct {
	Time  time.Time
	Valid bool
}

func String(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

func StringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

func Int64(i int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: i,
		Valid: i != 0,
	}
}

func Int64Ptr(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{
		Int64: *i,
		Valid: true,
	}
}

func Float64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   f != 0,
	}
}

func Float64Ptr(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{
		Float64: *f,
		Valid:   true,
	}
}

func Bool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: b,
	}
}

func BoolPtr(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{
		Bool:  *b,
		Valid: true,
	}
}

func Time(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

func TimePtr(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: !t.IsZero(),
	}
}

type StringBuilder struct {
	value sql.NullString
}

func NewStringBuilder(s string) *StringBuilder {
	return &StringBuilder{
		value: String(s),
	}
}

func (b *StringBuilder) Default(defaultValue string) sql.NullString {
	if !b.value.Valid {
		return sql.NullString{
			String: defaultValue,
			Valid:  true,
		}
	}
	return b.value
}
