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

// Send sends a success response using the response builder
func Send(c *fiber.Ctx, builder *response.Builder) error {
	c.Set("Content-Type", "application/json")
	c.Response().Header.Set("X-Content-Type-Options", "nosniff")
	return c.Status(builder.StatusCode()).JSON(builder.Build())
}

// SendError sends an error response using the error builder
func SendError(c *fiber.Ctx, errBuilder *errors.Builder) error {
	c.Set("Content-Type", "application/json")
	c.Response().Header.Set("X-Content-Type-Options", "nosniff")
	err := errBuilder.Build()
	return c.Status(err.Error.Code).JSON(err)
}

// OK sends a 200 OK response
func OK(c *fiber.Ctx, data interface{}) error {
	return Send(c, response.OK().Data(data))
}

// Created sends a 201 Created response
func Created(c *fiber.Ctx, data interface{}) error {
	return Send(c, response.Created().Data(data))
}

// Accepted sends a 202 Accepted response
func Accepted(c *fiber.Ctx, data interface{}) error {
	return Send(c, response.Accepted().Data(data))
}

// NoContent sends a 204 No Content response
func NoContent(c *fiber.Ctx) error {
	return Send(c, response.NoContent())
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *fiber.Ctx, summary string, detail ...interface{}) error {
	builder := errors.New().BadRequest().Summary(summary)
	if len(detail) > 0 {
		builder.Detail(detail[0])
	}
	return SendError(c, builder)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().Unauthorized().Summary(message))
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().Forbidden().Summary(message))
}

// NotFound sends a 404 Not Found error
func NotFound(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().NotFound().Summary(message))
}

// Conflict sends a 409 Conflict error
func Conflict(c *fiber.Ctx, message string) error {
	return SendError(c, errors.New().Conflict().Summary(message))
}

// InternalError sends a 500 Internal Server Error
func InternalError(c *fiber.Ctx, message string, err error) error {
	builder := errors.New().InternalError().Summary(message)
	if err != nil {
		builder.Detail(err)
	}
	return SendError(c, builder)
}
