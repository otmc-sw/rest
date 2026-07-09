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

	app.Post("/users", handlers.CreateUser)
	app.Get("/users/:id", handlers.GetUser)
	app.Patch("/users/:id", handlers.UpdateUser)
	app.Delete("/users/:id", handlers.DeleteUser)

	log.Fatal(app.Listen(":3000"))
}
