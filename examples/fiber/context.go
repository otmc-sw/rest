/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package main

import (
	"bytes"
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
)

type FiberContext struct {
	*fiber.Ctx
}

func (c FiberContext) Context() context.Context { return c.Ctx.Context() }

func (c FiberContext) Param(key string) string { return c.Ctx.Params(key) }

func (c FiberContext) Query(key string) string { return c.Ctx.Query(key) }

func (c FiberContext) QueryAll(key string) []string {
	v := c.Ctx.Query(key)
	if v == "" {
		return nil
	}
	return []string{v}
}

func (c FiberContext) Header(key string) string { return c.Ctx.Get(key) }

func (c FiberContext) Cookie(key string) string { return c.Ctx.Cookies(key) }

func (c FiberContext) Body() io.Reader { return bytes.NewReader(c.Ctx.Body()) }

func (c FiberContext) Bind(v interface{}) error { return c.Ctx.BodyParser(v) }

func (c FiberContext) JSON(code int, body interface{}) error {
	return c.Ctx.Status(code).JSON(body)
}

func (c FiberContext) Status(code int) { c.Ctx.Status(code) }

func (c FiberContext) SetHeader(key, value string) { c.Ctx.Set(key, value) }

func (c FiberContext) Method() string { return c.Ctx.Method() }

func (c FiberContext) Path() string { return c.Ctx.Path() }

func (c FiberContext) String() (string, error) { return string(c.Ctx.Body()), nil }

func (c FiberContext) Bytes() ([]byte, error) { return c.Ctx.Body(), nil }