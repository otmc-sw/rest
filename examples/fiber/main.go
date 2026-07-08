/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package main

import (
	"log"

	rest "github.com/otmc-sw/rest"
	"github.com/gofiber/fiber/v2"
)

func init() {
	// Register the entity -> response DTO mapper once at startup.
	rest.Register(func(u User) UserResponse {
		return UserResponse{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		}
	})
}

func main() {
	app := fiber.New()

	app.Post("/users", func(c *fiber.Ctx) error {
		return CreateUser(FiberContext{Ctx: c})
	})

	log.Fatal(app.Listen(":3000"))
}