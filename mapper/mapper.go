/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package mapper

import (
	"fmt"
	"sync"
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
}

func Map[D any, S any](src S) D {
	registryMu.RLock()
	defer registryMu.RUnlock()
	key := typeKey[S]() + "->" + typeKey[D]()
	if fn, ok := registry[key]; ok {
		if m, ok := fn.(func(S) D); ok {
			return m(src)
		}
	}
	return Auto[D](src)
}

func MapSlice[D any, S any](src []S) []D {
	dst := make([]D, 0, len(src))
	for _, s := range src {
		dst = append(dst, Map[D](s))
	}
	return dst
}

func Auto[D any, S any](src S) D {
	var dst D
	copyStructFields(&src, &dst)
	return dst
}

func typeKey[T any]() string {
	var t T
	return fmt.Sprintf("%T", t)
}
