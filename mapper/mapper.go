/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package mapper

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/otmc-sw/rest/debugger"
)

type MapperFunc[S any, D any] func(S) D

var (
	registry   = map[string]interface{}{}
	registryMu sync.RWMutex
)

func Register[S any, D any](fn MapperFunc[S, D]) {
	registryMu.Lock()
	defer registryMu.Unlock()
	key := typeKey[S]() + "->" + typeKey[D]()
	registry[key] = fn
	debugger.Mapper("Register[%s](%s)", typeKey[D](), typeKey[S]())
}

func Map[D any, S any](src S) D {
	start := time.Now()
	key := typeKey[S]() + "->" + typeKey[D]()
	debugger.Mapper("Map[%s](%s) start", typeKey[D](), typeKey[S]())

	registryMu.RLock()
	fn, ok := registry[key]
	registryMu.RUnlock()

	if ok {
		if m, ok := fn.(func(S) D); ok {
			result := m(src)
			debugger.MapperWithDuration(typeKey[S](), typeKey[D](), time.Since(start))
			return result
		}
	}
	result := Auto[D](src)
	debugger.MapperWithDuration(typeKey[S](), typeKey[D](), time.Since(start))
	return result
}

func MapSlice[D any, S any](src []S) []D {
	debugger.Mapper("MapSlice[%s](%s) count=%d", typeKey[D](), typeKey[S](), len(src))
	dst := make([]D, 0, len(src))
	for _, s := range src {
		dst = append(dst, Map[D](s))
	}
	return dst
}

func Auto[D any, S any](src S) D {
	debugger.Mapper("Auto[%s](%s)", typeKey[D](), typeKey[S]())
	var dst D

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(&dst).Elem()

	for srcVal.Kind() == reflect.Ptr || srcVal.Kind() == reflect.Interface {
		if srcVal.IsNil() {
			return dst
		}
		srcVal = srcVal.Elem()
	}

	if srcVal.Kind() == reflect.Slice && dstVal.Kind() == reflect.Slice {
		dstElemType := dstVal.Type().Elem()
		result := reflect.MakeSlice(dstVal.Type(), srcVal.Len(), srcVal.Len())
		for i := 0; i < srcVal.Len(); i++ {
			elem := reflect.New(dstElemType).Interface()
			copyStructFields(srcVal.Index(i).Interface(), elem)
			result.Index(i).Set(reflect.ValueOf(elem).Elem())
		}
		dstVal.Set(result)
		return dst
	}

	var srcPtr interface{}
	if srcVal.CanAddr() {
		srcPtr = srcVal.Addr().Interface()
	} else {
		srcPtr = &src
	}
	copyStructFields(srcPtr, &dst)
	return dst
}

func typeKey[T any]() string {
	var t T
	return fmt.Sprintf("%T", t)
}
