/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package fiber

import (
	"bytes"
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
	restcontext "github.com/otmc-sw/rest/context"
)

// FiberAdapter wraps Fiber's Ctx to implement the restcontext.Context interface
type FiberAdapter struct {
	*fiber.Ctx
}

// Wrap creates a new FiberAdapter from Fiber's Ctx
func Wrap(c *fiber.Ctx) restcontext.Context {
	return &FiberAdapter{Ctx: c}
}

// GetContext returns the underlying context.Context
func (a *FiberAdapter) GetContext() context.Context {
	return a.Ctx.Context()
}

// Param returns a path parameter by key
func (a *FiberAdapter) Param(key string) string {
	return a.Ctx.Params(key)
}

// Query returns a query parameter by key
func (a *FiberAdapter) Query(key string) string {
	return a.Ctx.Query(key)
}

// QueryAll returns all query parameters by key
func (a *FiberAdapter) QueryAll(key string) []string {
	value := a.Ctx.Query(key)
	if value == "" {
		return []string{}
	}
	return []string{value}
}

// Header returns a header value by key
func (a *FiberAdapter) Header(key string) string {
	return a.Ctx.Get(key)
}

// Cookie returns a cookie value by key
func (a *FiberAdapter) Cookie(key string) string {
	return a.Ctx.Cookies(key)
}

// Body returns the request body
func (a *FiberAdapter) Body() io.Reader {
	return bytes.NewReader(a.Ctx.Body())
}

// Bind binds the request body to a struct
func (a *FiberAdapter) Bind(v interface{}) error {
	return a.Ctx.BodyParser(v)
}

// Method returns the HTTP method
func (a *FiberAdapter) Method() string {
	return a.Ctx.Method()
}

// Path returns the request path
func (a *FiberAdapter) Path() string {
	return a.Ctx.Path()
}

// String returns the request body as string
func (a *FiberAdapter) String() (string, error) {
	return string(a.Ctx.Body()), nil
}

// Bytes returns the request body as bytes
func (a *FiberAdapter) Bytes() ([]byte, error) {
	return a.Ctx.Body(), nil
}
