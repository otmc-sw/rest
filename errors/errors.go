/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/otmc-sw/rest/context"
)

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

type Error struct {
	Success bool         `json:"success"`
	Details ErrorDetails `json:"error"`
}

func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Details.Code, e.Details.Key, e.Details.Summary)
}

func (e Error) Err() error {
	return fmt.Errorf("[%d] %s: %s", e.Details.Code, e.Details.Key, e.Details.Summary)
}

type Builder struct {
	details ErrorDetails
	skip    int
}

func New() *Builder {
	return &Builder{
		details: ErrorDetails{
			Timestamp: time.Now().Format(time.RFC3339),
		},
		skip: 2,
	}
}

func (b *Builder) BadRequest() *Builder {
	b.details.Code = 400
	b.details.Key = "BAD_REQUEST"
	b.details.Type = "Bad Request"
	return b
}

func (b *Builder) Unauthorized() *Builder {
	b.details.Code = 401
	b.details.Key = "UNAUTHORIZED"
	b.details.Type = "Unauthorized"
	return b
}

func (b *Builder) Forbidden() *Builder {
	b.details.Code = 403
	b.details.Key = "FORBIDDEN"
	b.details.Type = "Forbidden"
	return b
}

func (b *Builder) NotFound() *Builder {
	b.details.Code = 404
	b.details.Key = "NOT_FOUND"
	b.details.Type = "Not Found"
	return b
}

func (b *Builder) Conflict() *Builder {
	b.details.Code = 409
	b.details.Key = "CONFLICT"
	b.details.Type = "Conflict"
	return b
}

func (b *Builder) UnprocessableEntity() *Builder {
	b.details.Code = 422
	b.details.Key = "UNPROCESSABLE_ENTITY"
	b.details.Type = "Unprocessable Entity"
	return b
}

func (b *Builder) InternalError() *Builder {
	b.details.Code = 500
	b.details.Key = "INTERNAL_SERVER_ERROR"
	b.details.Type = "Internal Server Error"
	return b
}

func (b *Builder) ServiceUnavailable() *Builder {
	b.details.Code = 503
	b.details.Key = "SERVICE_UNAVAILABLE"
	b.details.Type = "Service Unavailable"
	return b
}

func (b *Builder) Code(code int) *Builder {
	b.details.Code = code
	return b
}

func (b *Builder) Key(key string) *Builder {
	b.details.Key = key
	return b
}

func (b *Builder) Type(typ string) *Builder {
	b.details.Type = typ
	return b
}

func (b *Builder) Summary(summary string) *Builder {
	b.details.Summary = summary
	return b
}

func (b *Builder) Detail(detail interface{}) *Builder {
	b.details.Detail = fmt.Sprint(detail)
	return b
}

func (b *Builder) Reason(reason string) *Builder {
	b.details.Reason = reason
	return b
}

func (b *Builder) Request(req interface{}) *Builder {
	b.details.Request = req
	return b
}

func (b *Builder) Data(data interface{}) *Builder {
	b.details.Data = data
	return b
}

func (b *Builder) Skip(skip int) *Builder {
	b.skip += skip
	return b
}

func (b *Builder) Build() Error {
	file, line, fn := getCallerInfo(b.skip)
	b.details.File = file
	b.details.Line = line
	b.details.Function = fn
	return Error{
		Success: false,
		Details: b.details,
	}
}

func (b *Builder) Send(ctx context.Context) error {
	err := b.Build()
	if ctx == nil {
		return err.Err()
	}
	return ctx.JSON(err.Details.Code, err)
}

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
