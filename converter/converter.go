/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package converter

import (
	"database/sql"
	"strconv"
	"time"
)

func Int64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func Int64OrDefault(s string, defaultValue int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

func String(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func StringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func Time(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func TimeString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format(time.RFC3339)
	}
	return ""
}

func Bool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

func Int64FromNull(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

func Float64FromNull(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func StringOrNull(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func StringPtrOrNull(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func BytesOrNull(b []byte) sql.NullString {
	s := string(b)
	return sql.NullString{String: s, Valid: s != ""}
}

func Int64OrNull(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: i != 0}
}

func Int64PtrOrNull(i *int64) sql.NullInt64 {
	return sql.NullInt64{Int64: *i, Valid: i != nil}
}

func Float64OrNull(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: f != 0}
}

func Float64PtrOrNull(f *float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: *f, Valid: f != nil}
}

func BoolOrNull(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func BoolPtrOrNull(b *bool) sql.NullBool {
	return sql.NullBool{Bool: *b, Valid: true}
}

func StringOrEmpty(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func StringPtrOrEmpty(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func Float64PtrOrZero(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func Int64PtrOrZero(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func StringOrDefault(s string, defaultValue string) sql.NullString {
	if s == "" {
		return sql.NullString{String: defaultValue, Valid: true}
	}
	return sql.NullString{String: s, Valid: true}
}

func Int64PtrOrDefault(i *int64, defaultValue int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Int64: defaultValue, Valid: true}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func Float64PtrOrDefault(f *float64, defaultValue float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Float64: defaultValue, Valid: true}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func BoolPtrOrDefault(b *bool, defaultValue bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Bool: defaultValue, Valid: true}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

func TimeOrDefault(t time.Time, defaultValue time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Time: defaultValue, Valid: true}
	}
	return sql.NullTime{Time: t, Valid: true}
}


