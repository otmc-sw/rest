/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package mapper

import (
	"fmt"
	"reflect"
	"strconv"
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
				if srcFieldVal.Type().AssignableTo(dstField.Type()) {
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

	return fmt.Errorf("cannot convert %v to %v", srcType, dstType)
}
