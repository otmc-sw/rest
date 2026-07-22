/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package config

import (
	"sync"
)

type OperationConfig struct {
	fields map[string]any
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
