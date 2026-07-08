/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package convert

import (
	"database/sql"
	"strconv"
	"time"
)

// Int64 converts a string to int64
func Int64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Int64OrDefault converts a string to int64 with default value
func Int64OrDefault(s string, defaultValue int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

// String converts sql.NullString to string
func String(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// StringPtr converts sql.NullString to string pointer
func StringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// Time converts sql.NullTime to time.Time
func Time(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

// TimeString converts sql.NullTime to ISO string
func TimeString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format(time.RFC3339)
	}
	return ""
}

// Bool converts sql.NullBool to bool
func Bool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

// Int64FromNull converts sql.NullInt64 to int64
func Int64FromNull(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

// Float64FromNull converts sql.NullFloat64 to float64
func Float64FromNull(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

// ToNullString converts string to sql.NullString
func ToNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// ToNullInt64 converts int64 to sql.NullInt64
func ToNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}

// ToNullInt64Ptr converts int64 pointer to sql.NullInt64
func ToNullInt64Ptr(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{
		Int64: *i,
		Valid: true,
	}
}

// ToNullFloat64 converts float64 to sql.NullFloat64
func ToNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   true,
	}
}

// ToNullFloat64Ptr converts float64 pointer to sql.NullFloat64
func ToNullFloat64Ptr(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{
		Float64: *f,
		Valid:   true,
	}
}

// ToNullBool converts bool to sql.NullBool
func ToNullBool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: true,
	}
}

// ToNullBoolPtr converts bool pointer to sql.NullBool
func ToNullBoolPtr(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{
		Bool:  *b,
		Valid: true,
	}
}
