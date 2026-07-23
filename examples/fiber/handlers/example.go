/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
)

func TestResponse(c *fiber.Ctx) error {
	data := map[string]any{
		"status":    "success",
		"test_data": "Hello World",
	}
	return rest.OK(FiberContext{Ctx: c}).Data(data).Message("OK").Send()
}
