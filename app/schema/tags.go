/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package schema

import (
	"fmt"
	"reflect"
	"strings"
)

func errInvalidModel(t reflect.Type) error {
	if t == nil {
		return fmt.Errorf("otmc: model must be a pointer to a struct, got nil")
	}
	return fmt.Errorf("otmc: model must be a pointer to a struct, got %s", t.String())
}

// parseJSONTag splits a `json:"name,omitempty"` tag into name and options.
func parseJSONTag(tag string) (name string, opts map[string]bool) {
	if tag == "" {
		return "", nil
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	opts = map[string]bool{}
	for _, p := range parts[1:] {
		opts[p] = true
	}
	return name, opts
}

// parseDirectives parses `otmc:"required;unique;size=255"` style tags.
func parseDirectives(tag string) map[string]string {
	out := map[string]string{}
	if tag == "" {
		return out
	}
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if k, v, ok := strings.Cut(part, "="); ok {
			out[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
		} else {
			out[strings.ToLower(part)] = "true"
		}
	}
	return out
}
