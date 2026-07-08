/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang Logger.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package nullable

import "time"

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

func String(s string) NullString {
	return NullString{
		String: s,
		Valid:  s != "",
	}
}

func StringPtr(s *string) NullString {
	if s == nil {
		return NullString{Valid: false}
	}
	return NullString{
		String: *s,
		Valid:  true,
	}
}

func Int64(i int64) NullInt64 {
	return NullInt64{
		Int64: i,
		Valid: i != 0,
	}
}

func Int64Ptr(i *int64) NullInt64 {
	if i == nil {
		return NullInt64{Valid: false}
	}
	return NullInt64{
		Int64: *i,
		Valid: true,
	}
}

func Float64(f float64) NullFloat64 {
	return NullFloat64{
		Float64: f,
		Valid:   f != 0,
	}
}

func Float64Ptr(f *float64) NullFloat64 {
	if f == nil {
		return NullFloat64{Valid: false}
	}
	return NullFloat64{
		Float64: *f,
		Valid:   true,
	}
}

func Bool(b bool) NullBool {
	return NullBool{
		Bool:  b,
		Valid: b,
	}
}

func BoolPtr(b *bool) NullBool {
	if b == nil {
		return NullBool{Valid: false}
	}
	return NullBool{
		Bool:  *b,
		Valid: true,
	}
}

func Time(t time.Time) NullTime {
	return NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

func TimePtr(t *time.Time) NullTime {
	if t == nil {
		return NullTime{Valid: false}
	}
	return NullTime{
		Time:  *t,
		Valid: !t.IsZero(),
	}
}

type StringBuilder struct {
	value NullString
}

func NewStringBuilder(s string) *StringBuilder {
	return &StringBuilder{
		value: String(s),
	}
}

func (b *StringBuilder) Default(defaultValue string) NullString {
	if !b.value.Valid {
		return NullString{
			String: defaultValue,
			Valid:  true,
		}
	}
	return b.value
}
