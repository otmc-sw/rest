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

type FiberAdapter struct {
	*fiber.Ctx
}

func Wrap(c *fiber.Ctx) restcontext.Context {
	return &FiberAdapter{Ctx: c}
}

func (a *FiberAdapter) GetContext() context.Context {
	return a.Ctx.Context()
}

func (a *FiberAdapter) Param(key string) string {
	return a.Ctx.Params(key)
}

func (a *FiberAdapter) Query(key string) string {
	return a.Ctx.Query(key)
}

func (a *FiberAdapter) QueryAll(key string) []string {
	value := a.Ctx.Query(key)
	if value == "" {
		return []string{}
	}
	return []string{value}
}

func (a *FiberAdapter) Header(key string) string {
	return a.Ctx.Get(key)
}

func (a *FiberAdapter) Cookie(key string) string {
	return a.Ctx.Cookies(key)
}

func (a *FiberAdapter) Body() io.Reader {
	return bytes.NewReader(a.Ctx.Body())
}

func (a *FiberAdapter) Bind(v interface{}) error {
	return a.Ctx.BodyParser(v)
}

func (a *FiberAdapter) Method() string {
	return a.Ctx.Method()
}

func (a *FiberAdapter) Path() string {
	return a.Ctx.Path()
}

func (a *FiberAdapter) String() (string, error) {
	return string(a.Ctx.Body()), nil
}

func (a *FiberAdapter) Bytes() ([]byte, error) {
	return a.Ctx.Body(), nil
}
