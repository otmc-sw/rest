/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	"github.com/otmc-sw/rest/examples/fiber/services"
)

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func CreateUser(c *fiber.Ctx) error {
	return rest.
		Create[services.CreateUserRequest, services.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(services.Validates()).
		Handle(services.CreateUserHandler()).
		Respond()
}

func GetUser(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, services.User, UserResponse](FiberContext{Ctx: c}).
		Handle(services.GetUserHandler()).
		Respond()
}

func UpdateUser(c *fiber.Ctx) error {
	return rest.
		Update[services.UpdateUserRequest, services.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(services.ValidatesUpdate()).
		Handle(services.UpdateUserHandler()).
		Respond()
}

func DeleteUser(c *fiber.Ctx) error {
	return rest.
		Delete[UserResponse](FiberContext{Ctx: c}).
		Handle(services.DeleteUserHandler()).
		Respond()
}

