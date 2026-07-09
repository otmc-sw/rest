/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package debugger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Component string

const (
	ComponentMapper    Component = "mapper"
	ComponentPipeline  Component = "pipeline"
	ComponentValidator Component = "validator"
	ComponentErrors    Component = "errors"
	ComponentRequest   Component = "request"
	ComponentResponse  Component = "response"
	ComponentFilter    Component = "filter"
	ComponentConverter Component = "converter"
	ComponentAll       Component = "all"
)

var (
	mu             sync.RWMutex
	enabled        bool
	componentFlags = map[Component]bool{}
	logger         = log.New(os.Stdout, "[REST] ", log.Lmicroseconds)
)

func Enable() {
	mu.Lock()
	defer mu.Unlock()
	enabled = true
}

func Disable() {
	mu.Lock()
	defer mu.Unlock()
	enabled = false
}

func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

func EnableComponent(component string) {
	mu.Lock()
	defer mu.Unlock()
	switch Component(component) {
	case ComponentAll:
		enabled = true
		for c := range componentFlags {
			componentFlags[c] = true
		}
	default:
		componentFlags[Component(component)] = true
	}
}

func DisableComponent(component string) {
	mu.Lock()
	defer mu.Unlock()
	if Component(component) == ComponentAll {
		enabled = false
		componentFlags = map[Component]bool{}
	} else {
		delete(componentFlags, Component(component))
	}
}

func IsComponentEnabled(component Component) bool {
	mu.RLock()
	defer mu.RUnlock()
	if enabled {
		return true
	}
	return componentFlags[component]
}

func WithEnvFromString(value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		EnableComponent(part)
	}
}

func WithEnv() {
	value := os.Getenv("REST_DEBUG")
	WithEnvFromString(value)
}

func logDebug(component Component, format string, args ...interface{}) {
	if !IsComponentEnabled(component) && component != ComponentAll {
		return
	}
	prefix := fmt.Sprintf("[%-8s] ", component)
	msg := fmt.Sprintf(format, args...)
	logger.Output(3, prefix+msg)
}

func Mapper(format string, args ...interface{}) {
	logDebug(ComponentMapper, format, args...)
}

func MapperWithDuration(srcType, dstType string, duration time.Duration) {
	Mapper("Map[%s](%s) took %v", dstType, srcType, duration)
}

func Pipeline(format string, args ...interface{}) {
	logDebug(ComponentPipeline, format, args...)
}

func PipelineStep(step string, format string, args ...interface{}) {
	logDebug(ComponentPipeline, "[%s] "+format, append([]interface{}{step}, args...)...)
}

func Validator(format string, args ...interface{}) {
	logDebug(ComponentValidator, format, args...)
}

func Errors(format string, args ...interface{}) {
	logDebug(ComponentErrors, format, args...)
}

func Request(format string, args ...interface{}) {
	logDebug(ComponentRequest, format, args...)
}

func Response(format string, args ...interface{}) {
	logDebug(ComponentResponse, format, args...)
}

func Filter(format string, args ...interface{}) {
	logDebug(ComponentFilter, format, args...)
}

func Converter(format string, args ...interface{}) {
	logDebug(ComponentConverter, format, args...)
}
