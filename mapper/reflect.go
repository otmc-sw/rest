/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package mapper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func copyStructFields(src, dst interface{}) {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	for srcVal.Kind() == reflect.Ptr || srcVal.Kind() == reflect.Interface {
		if srcVal.IsNil() {
			return
		}
		srcVal = srcVal.Elem()
	}
	for dstVal.Kind() == reflect.Ptr || dstVal.Kind() == reflect.Interface {
		if dstVal.IsNil() {
			return
		}
		dstVal = dstVal.Elem()
	}

	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return
	}

	dstType := dstVal.Type()
	dstFields := map[string]int{}
	for i := 0; i < dstType.NumField(); i++ {
		dstFields[dstType.Field(i).Name] = i
	}

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Type().Field(i)
		if idx, ok := dstFields[srcField.Name]; ok {
			dstField := dstVal.Field(idx)
			srcFieldVal := srcVal.Field(i)
			if dstField.CanSet() {
				srcIsNull := isNullType(srcFieldVal.Type())
				dstIsNull := isNullType(dstField.Type())
				if srcFieldVal.Type().AssignableTo(dstField.Type()) && (!srcIsNull || dstIsNull) {
					dstField.Set(srcFieldVal)
				} else if err := convertAndSet(dstField, srcFieldVal); err == nil {
				}
			}
		}
	}
}

func convertAndSet(dstField, srcField reflect.Value) error {
	srcType := srcField.Type()
	dstType := dstField.Type()

	if srcType.Kind() == reflect.Ptr {
		if srcField.IsNil() {
			if dstType.Kind() == reflect.String {
				dstField.SetString("")
				return nil
			}
			if dstType == nullStringType {
				dstField.Set(reflect.ValueOf(sql.NullString{Valid: false}))
				return nil
			}
			return nil
		}
		elem := srcField.Elem()
		elemType := elem.Type()

		if elemType == jsonRawMessageType {
			raw := elem.Interface().(json.RawMessage)
			if dstType.Kind() == reflect.String {
				dstField.SetString(string(raw))
				return nil
			}
			if dstType == nullStringType {
				s := string(raw)
				dstField.Set(reflect.ValueOf(sql.NullString{String: s, Valid: true}))
				return nil
			}
		}

		if elemType.Kind() == reflect.String && dstType.Kind() == reflect.String {
			dstField.SetString(elem.String())
			return nil
		}

		if elemType.Kind() == reflect.String && dstType == nullStringType {
			s := elem.String()
			dstField.Set(reflect.ValueOf(sql.NullString{String: s, Valid: s != ""}))
			return nil
		}

		srcField = elem
		srcType = elemType
	}

	if srcType.Kind() == reflect.Int64 && dstType.Kind() == reflect.String {
		dstField.SetString(strconv.FormatInt(srcField.Int(), 10))
		return nil
	}

	if srcType.Kind() == reflect.Int && dstType.Kind() == reflect.String {
		dstField.SetString(strconv.Itoa(int(srcField.Int())))
		return nil
	}

	if srcType.Kind() == reflect.String && dstType.Kind() == reflect.Int64 {
		val, err := strconv.ParseInt(srcField.String(), 10, 64)
		if err != nil {
			return err
		}
		dstField.SetInt(val)
		return nil
	}

	if srcType.Kind() == reflect.String && dstType.Kind() == reflect.Int {
		val, err := strconv.Atoi(srcField.String())
		if err != nil {
			return err
		}
		dstField.SetInt(int64(val))
		return nil
	}

	if err := convertNullTypes(dstField, srcField); err == nil {
		return nil
	}

	return fmt.Errorf("cannot convert %v to %v", srcType, dstType)
}

func convertNullTypes(dstField, srcField reflect.Value) error {
	srcType := srcField.Type()
	dstType := dstField.Type()

	srcIsNull := srcType == nullStringType || srcType == nullInt64Type ||
		srcType == nullFloat64Type || srcType == nullBoolType || srcType == nullTimeType
	dstIsNull := dstType == nullStringType || dstType == nullInt64Type ||
		dstType == nullFloat64Type || dstType == nullBoolType || dstType == nullTimeType

	if !srcIsNull && !dstIsNull {
		return fmt.Errorf("neither field is a sql.Null type")
	}

	switch {
	case srcType == nullStringType && dstType.Kind() == reflect.String:
		ns := srcField.Interface().(sql.NullString)
		if ns.Valid {
			dstField.SetString(ns.String)
		} else {
			dstField.SetString("")
		}
	case dstType == nullStringType && srcType.Kind() == reflect.String:
		dstField.Set(reflect.ValueOf(sql.NullString{String: srcField.String(), Valid: srcField.String() != ""}))
	case srcType == nullInt64Type && (dstType.Kind() == reflect.Int64 || dstType.Kind() == reflect.Int):
		ni := srcField.Interface().(sql.NullInt64)
		if dstType.Kind() == reflect.Int64 {
			if ni.Valid {
				dstField.SetInt(ni.Int64)
			} else {
				dstField.SetInt(0)
			}
		} else {
			if ni.Valid {
				dstField.SetInt(int64(ni.Int64))
			} else {
				dstField.SetInt(0)
			}
		}
	case dstType == nullInt64Type && (srcType.Kind() == reflect.Int64 || srcType.Kind() == reflect.Int):
		v := srcField.Int()
		dstField.Set(reflect.ValueOf(sql.NullInt64{Int64: v, Valid: v != 0}))
	case srcType == nullFloat64Type && dstType.Kind() == reflect.Float64:
		nf := srcField.Interface().(sql.NullFloat64)
		if nf.Valid {
			dstField.SetFloat(nf.Float64)
		} else {
			dstField.SetFloat(0)
		}
	case dstType == nullFloat64Type && srcType.Kind() == reflect.Float64:
		dstField.Set(reflect.ValueOf(sql.NullFloat64{Float64: srcField.Float(), Valid: true}))
	case srcType == nullBoolType && dstType.Kind() == reflect.Bool:
		nb := srcField.Interface().(sql.NullBool)
		if nb.Valid {
			dstField.SetBool(nb.Bool)
		} else {
			dstField.SetBool(false)
		}
	case dstType == nullBoolType && srcType.Kind() == reflect.Bool:
		dstField.Set(reflect.ValueOf(sql.NullBool{Bool: srcField.Bool(), Valid: true}))
	case srcType == nullTimeType && dstType == timeTimeType:
		nt := srcField.Interface().(sql.NullTime)
		if nt.Valid {
			dstField.Set(reflect.ValueOf(nt.Time))
		} else {
			dstField.Set(reflect.ValueOf(time.Time{}))
		}
	case dstType == nullTimeType && srcType == timeTimeType:
		dstField.Set(reflect.ValueOf(sql.NullTime{Time: srcField.Interface().(time.Time), Valid: true}))
	case srcType == nullStringType && dstType == interfaceType:
		ns := srcField.Interface().(sql.NullString)
		if ns.Valid {
			var v interface{}
			if err := json.Unmarshal([]byte(ns.String), &v); err == nil {
				dstField.Set(reflect.ValueOf(v))
			} else {
				dstField.Set(reflect.ValueOf(ns.String))
			}
		} else {
			dstField.Set(reflect.ValueOf(""))
		}
	case dstType == nullStringType && srcType == interfaceType:
		if s, ok := srcField.Interface().(string); ok {
			dstField.Set(reflect.ValueOf(sql.NullString{String: s, Valid: s != ""}))
		} else {
			dstField.Set(reflect.ValueOf(sql.NullString{Valid: false}))
		}
	default:
		return fmt.Errorf("unsupported sql.Null conversion")
	}
	return nil
}

var (
	nullStringType     = reflect.TypeOf(sql.NullString{})
	nullInt64Type      = reflect.TypeOf(sql.NullInt64{})
	nullFloat64Type    = reflect.TypeOf(sql.NullFloat64{})
	nullBoolType       = reflect.TypeOf(sql.NullBool{})
	nullTimeType       = reflect.TypeOf(sql.NullTime{})
	timeTimeType       = reflect.TypeOf(time.Time{})
	interfaceType      = reflect.TypeOf((*interface{})(nil)).Elem()
	jsonRawMessageType = reflect.TypeOf(json.RawMessage{})
)

func isNullType(t reflect.Type) bool {
	return t == nullStringType || t == nullInt64Type ||
		t == nullFloat64Type || t == nullBoolType || t == nullTimeType
}
