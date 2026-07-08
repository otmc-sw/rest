/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package mapper

import "reflect"

func copyStructFields(src, dst interface{}) {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	for srcVal.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return
		}
		srcVal = srcVal.Elem()
	}
	for dstVal.Kind() == reflect.Ptr {
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
			if dstField.CanSet() && srcFieldVal.Type().AssignableTo(dstField.Type()) {
				dstField.Set(srcFieldVal)
			}
		}
	}
}