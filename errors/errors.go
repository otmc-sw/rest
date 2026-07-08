/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

// ErrorDetails represents standard REST error structure
type ErrorDetails struct {
	Code      int         `json:"code"`               // HTTP status code: 400, 404, 500, ...
	Key       string      `json:"key"`                // Error key: BAD_REQUEST, NOT_FOUND, INTERNAL_SERVER_ERROR, ...
	Type      string      `json:"type"`               // Error type: Bad Request, Not Found, Backend Server Error
	Summary   string      `json:"summary"`            // User-friendly summary
	Detail    string      `json:"detail"`             // Detailed error message
	Reason    string      `json:"reason,omitempty"`   // Failure reason (if available)
	Request   interface{} `json:"request,omitempty"`  // Request body (if available)
	Data      interface{} `json:"data,omitempty"`     // Additional error data
	File      string      `json:"file,omitempty"`     // File name where error occurred
	Line      int         `json:"line,omitempty"`     // Line number
	Function  string      `json:"function,omitempty"` // Function name
	Timestamp string      `json:"timestamp"`          // ISO timestamp
}

// Error represents a REST error response
type Error struct {
	Success bool         `json:"success"`
	Error   ErrorDetails `json:"error"`
}

// Builder provides fluent API for building errors
type Builder struct {
	details ErrorDetails
	skip     int
}

// New creates a new error builder
func New() *Builder {
	return &Builder{
		details: ErrorDetails{
			Timestamp: time.Now().Format(time.RFC3339),
		},
		skip: 2,
	}
}

// BadRequest sets error to 400 Bad Request
func (b *Builder) BadRequest() *Builder {
	b.details.Code = 400
	b.details.Key = "BAD_REQUEST"
	b.details.Type = "Bad Request"
	return b
}

// Unauthorized sets error to 401 Unauthorized
func (b *Builder) Unauthorized() *Builder {
	b.details.Code = 401
	b.details.Key = "UNAUTHORIZED"
	b.details.Type = "Unauthorized"
	return b
}

// Forbidden sets error to 403 Forbidden
func (b *Builder) Forbidden() *Builder {
	b.details.Code = 403
	b.details.Key = "FORBIDDEN"
	b.details.Type = "Forbidden"
	return b
}

// NotFound sets error to 404 Not Found
func (b *Builder) NotFound() *Builder {
	b.details.Code = 404
	b.details.Key = "NOT_FOUND"
	b.details.Type = "Not Found"
	return b
}

// Conflict sets error to 409 Conflict
func (b *Builder) Conflict() *Builder {
	b.details.Code = 409
	b.details.Key = "CONFLICT"
	b.details.Type = "Conflict"
	return b
}

// UnprocessableEntity sets error to 422 Unprocessable Entity
func (b *Builder) UnprocessableEntity() *Builder {
	b.details.Code = 422
	b.details.Key = "UNPROCESSABLE_ENTITY"
	b.details.Type = "Unprocessable Entity"
	return b
}

// InternalError sets error to 500 Internal Server Error
func (b *Builder) InternalError() *Builder {
	b.details.Code = 500
	b.details.Key = "INTERNAL_SERVER_ERROR"
	b.details.Type = "Internal Server Error"
	return b
}

// ServiceUnavailable sets error to 503 Service Unavailable
func (b *Builder) ServiceUnavailable() *Builder {
	b.details.Code = 503
	b.details.Key = "SERVICE_UNAVAILABLE"
	b.details.Type = "Service Unavailable"
	return b
}

// Code sets custom HTTP status code
func (b *Builder) Code(code int) *Builder {
	b.details.Code = code
	return b
}

// Key sets error key
func (b *Builder) Key(key string) *Builder {
	b.details.Key = key
	return b
}

// Type sets error type
func (b *Builder) Type(typ string) *Builder {
	b.details.Type = typ
	return b
}

// Summary sets error summary
func (b *Builder) Summary(summary string) *Builder {
	b.details.Summary = summary
	return b
}

// Detail sets error detail
func (b *Builder) Detail(detail interface{}) *Builder {
	b.details.Detail = fmt.Sprint(detail)
	return b
}

// Reason sets error reason
func (b *Builder) Reason(reason string) *Builder {
	b.details.Reason = reason
	return b
}

// Request sets request body
func (b *Builder) Request(req interface{}) *Builder {
	b.details.Request = req
	return b
}

// Data sets additional error data
func (b *Builder) Data(data interface{}) *Builder {
	b.details.Data = data
	return b
}

// Skip sets the number of stack frames to skip for caller info
func (b *Builder) Skip(skip int) *Builder {
	b.skip = skip
	return b
}

// Build finalizes the error and captures caller information
func (b *Builder) Build() Error {
	file, line, fn := getCallerInfo(b.skip)
	b.details.File = file
	b.details.Line = line
	b.details.Function = fn
	return Error{
		Success: false,
		Error:   b.details,
	}
}

// getCallerInfo retrieves caller information
func getCallerInfo(skip int) (file string, line int, funcName string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", 0, "unknown"
	}
	funcName = "unknown"
	if f := runtime.FuncForPC(pc); f != nil {
		parts := splitFuncName(f.Name())
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
		}
	}
	return filepath.Base(file), line, funcName
}

func splitFuncName(name string) []string {
	var parts []string
	start := 0
	for i, c := range name {
		if c == '.' {
			parts = append(parts, name[start:i])
			start = i + 1
		}
	}
	parts = append(parts, name[start:])
	return parts
}
