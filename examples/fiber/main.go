/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/otmc-sw/rest/examples/fiber/handlers"
)

func main() {
	app := fiber.New()

	app.Post("/users", func(c *fiber.Ctx) error {
		return handlers.CreateUser(FiberContext{Ctx: c})
	})

	log.Fatal(app.Listen(":3000"))
}
