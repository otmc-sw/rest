/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/otmc-sw/logger"
	"github.com/otmc-sw/rest/examples/fiber/handlers"
)

func main() {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Request(c.Method(), c.Path(), c.Response().StatusCode(), time.Since(start), c.IP())
		return err
	})

	app.Post("/users", func(c *fiber.Ctx) error {
		return handlers.CreateUser(FiberContext{Ctx: c})
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		return handlers.GetUser(FiberContext{Ctx: c})
	})

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateUser(FiberContext{Ctx: c})
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteUser(FiberContext{Ctx: c})
	})

	log.Fatal(app.Listen(":3000"))
}
