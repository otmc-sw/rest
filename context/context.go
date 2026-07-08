/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package context

import (
	"context"
	"io"
)

// Context defines the interface for framework-agnostic context
// This allows the core library to work with any web framework
type Context interface {
	// GetContext returns the underlying context.Context
	GetContext() context.Context

	// Param returns a path parameter by key
	Param(key string) string

	// Query returns a query parameter by key
	Query(key string) string

	// QueryAll returns all query parameters by key
	QueryAll(key string) []string

	// Header returns a header value by key
	Header(key string) string

	// Cookie returns a cookie value by key
	Cookie(key string) string

	// Body returns the request body
	Body() io.Reader

	// Bind binds the request body to a struct
	Bind(v interface{}) error

	// Method returns the HTTP method
	Method() string

	// Path returns the request path
	Path() string

	// String returns the request body as string
	String() (string, error)

	// Bytes returns the request body as bytes
	Bytes() ([]byte, error)
}
