/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package store

import "fmt"

// ValidationError is returned when a required field is missing or invalid.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("field %q: %s", e.Field, e.Msg)
	}
	return fmt.Sprintf("field %q is required", e.Field)
}

// UniqueError is returned when a unique constraint is violated.
type UniqueError struct {
	Field string
}

func (e *UniqueError) Error() string {
	return fmt.Sprintf("field %q must be unique", e.Field)
}
