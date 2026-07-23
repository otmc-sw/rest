/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package config

import (
	"reflect"
	"sync"
)

type OperationConfig struct {
	fields     map[string]any
	fieldFuncs map[string]func(any) any
}

func (oc *OperationConfig) SetField(key string, value any) *OperationConfig {
	if oc.fields == nil {
		oc.fields = make(map[string]any)
	}
	oc.fields[key] = value
	return oc
}

func (oc *OperationConfig) SetFields(fields map[string]any) *OperationConfig {
	if oc.fields == nil {
		oc.fields = make(map[string]any)
	}
	for k, v := range fields {
		oc.fields[k] = v
	}
	return oc
}

func (oc *OperationConfig) AppendField(key string, value any) *OperationConfig {
	if oc.fields == nil {
		oc.fields = make(map[string]any)
	}
	oc.fields[key] = value
	return oc
}

func (oc *OperationConfig) AppendFields(fields map[string]any) *OperationConfig {
	if oc.fields == nil {
		oc.fields = make(map[string]any)
	}
	for k, v := range fields {
		oc.fields[k] = v
	}
	return oc
}

func (oc *OperationConfig) GetFields() map[string]any {
	return oc.fields
}

func (oc *OperationConfig) GetFieldFuncs() map[string]func(any) any {
	return oc.fieldFuncs
}

func (oc *OperationConfig) SetFieldFunc(key string, fn func(any) any) *OperationConfig {
	if oc.fieldFuncs == nil {
		oc.fieldFuncs = make(map[string]func(any) any)
	}
	oc.fieldFuncs[key] = fn
	return oc
}

func (oc *OperationConfig) SetFieldsFuncs(fields map[string]func(any) any) *OperationConfig {
	if oc.fieldFuncs == nil {
		oc.fieldFuncs = make(map[string]func(any) any)
	}
	for k, v := range fields {
		oc.fieldFuncs[k] = v
	}
	return oc
}

func (oc *OperationConfig) AppendFieldFunc(key string, fn func(any) any) *OperationConfig {
	if oc.fieldFuncs == nil {
		oc.fieldFuncs = make(map[string]func(any) any)
	}
	oc.fieldFuncs[key] = fn
	return oc
}

func (oc *OperationConfig) AppendFieldsFuncs(fields map[string]func(any) any) *OperationConfig {
	if oc.fieldFuncs == nil {
		oc.fieldFuncs = make(map[string]func(any) any)
	}
	for k, v := range fields {
		oc.fieldFuncs[k] = v
	}
	return oc
}

type Config struct {
	post   *OperationConfig
	get    *OperationConfig
	update *OperationConfig
	patch  *OperationConfig
	delete *OperationConfig
}

func (c *Config) Post() *OperationConfig {
	if c.post == nil {
		c.post = &OperationConfig{}
	}
	return c.post
}

func (c *Config) Get() *OperationConfig {
	if c.get == nil {
		c.get = &OperationConfig{}
	}
	return c.get
}

func (c *Config) Update() *OperationConfig {
	if c.update == nil {
		c.update = &OperationConfig{}
	}
	return c.update
}

func (c *Config) Patch() *OperationConfig {
	if c.patch == nil {
		c.patch = &OperationConfig{}
	}
	return c.patch
}

func (c *Config) Delete() *OperationConfig {
	if c.delete == nil {
		c.delete = &OperationConfig{}
	}
	return c.delete
}

var (
	globalConfig     *Config
	globalConfigOnce sync.Once
	globalConfigMu   sync.RWMutex
)

func Configure(fn func(*Config)) {
	globalConfigOnce.Do(func() {
		globalConfig = &Config{}
	})
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	fn(globalConfig)
}

func GetGlobalConfig() *Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

func GetField(obj any, fieldName string) any {
	if obj == nil {
		return nil
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return field.Interface()
}

func GetFieldInt64(obj any, fieldName string) int64 {
	if v := GetField(obj, fieldName); v != nil {
		if i, ok := v.(int64); ok {
			return i
		}
		if i, ok := v.(int); ok {
			return int64(i)
		}
	}
	return 0
}

func GetFieldString(obj any, fieldName string) string {
	if v := GetField(obj, fieldName); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
