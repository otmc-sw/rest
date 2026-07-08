/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/response"
)

func Send(c *fiber.Ctx, builder *response.Builder) error {
	c.Set("Content-Type", "application/json")
	c.Response().Header.Set("X-Content-Type-Options", "nosniff")
	return c.Status(builder.StatusCode()).JSON(builder.Build())
}

func SendError(c *fiber.Ctx, errBuilder *errors.Builder) error {
	c.Set("Content-Type", "application/json")
	c.Response().Header.Set("X-Content-Type-Options", "nosniff")
	err := errBuilder.Build()
	return c.Status(err.Error.Code).JSON(err)
}

func OK(c *fiber.Ctx, data interface{}) error {
	return Send(c, response.OK().Data(data))
}

func Created(c *fiber.Ctx, data interface{}) error {
	return Send(c, response.Created().Data(data))
}

func Accepted(c *fiber.Ctx, data interface{}) error {
	return Send(c, response.Accepted().Data(data))
}

func NoContent(c *fiber.Ctx) error {
	return Send(c, response.NoContent())
}

func BadRequest(c *fiber.Ctx, summary string, detail ...interface{}) error {
	builder := errors.New().BadRequest().Summary(summary)
	if len(detail) > 0 {
		builder.Detail(detail[0])
	}
	return SendError(c, builder)
}

func Unauthorized(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().Unauthorized().Summary(message))
}

func Forbidden(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().Forbidden().Summary(message))
}

func NotFound(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().NotFound().Summary(message))
}

func Conflict(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().Conflict().Summary(message))
}

func InternalError(c *fiber.Ctx, message string, err error) error {
	builder := errors.New().InternalError().Summary(message)
	if err != nil {
		builder.Detail(err)
	}
	return SendError(c, builder)
}
