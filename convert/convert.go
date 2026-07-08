/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package convert

import (
	"strconv"
	"time"

	"github.com/otmc-sw/rest/nullable"
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

func String(ns nullable.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func StringPtr(ns nullable.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func Time(nt nullable.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func TimeString(nt nullable.NullTime) string {
	if nt.Valid {
		return nt.Time.Format(time.RFC3339)
	}
	return ""
}

func Bool(nb nullable.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

func Int64FromNull(ni nullable.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

func Float64FromNull(nf nullable.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func ToNullString(s string) nullable.NullString {
	return nullable.String(s)
}

func ToNullInt64(i int64) nullable.NullInt64 {
	return nullable.Int64(i)
}

func ToNullInt64Ptr(i *int64) nullable.NullInt64 {
	return nullable.Int64Ptr(i)
}

func ToNullFloat64(f float64) nullable.NullFloat64 {
	return nullable.Float64(f)
}

func ToNullFloat64Ptr(f *float64) nullable.NullFloat64 {
	return nullable.Float64Ptr(f)
}

func ToNullBool(b bool) nullable.NullBool {
	return nullable.Bool(b)
}

func ToNullBoolPtr(b *bool) nullable.NullBool {
	return nullable.BoolPtr(b)
}
