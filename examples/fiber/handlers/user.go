/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
)

type User struct {
	ID    string
	Name  string
	Email string
}

type UserRequest struct {
	Name  string
	Email string
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ValidateUser(r UserRequest) error {
	return rest.Validate().
		Required(r.Name).
		Email(r.Email).
		Validate()
}

func CreateUser(c *fiber.Ctx) error {
	return rest.
		Create[UserRequest, User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(ValidateUser).
		Handle(func(ctx rest.Context, req UserRequest) (User, error) {
			user := User{
				ID:    fmt.Sprintf("usr_%s", req.Email),
				Name:  req.Name,
				Email: req.Email,
			}
			return user, nil
		}).
		Respond()
}

func GetUser(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, User, UserResponse](FiberContext{Ctx: c}).
		Handle(func(ctx rest.Context, req struct{}) (User, error) {
			id := ctx.Param("id")
			user := User{
				ID:    id,
				Name:  "John Doe",
				Email: "john@example.com",
			}
			return user, nil
		}).
		Respond()
}

func UpdateUser(c *fiber.Ctx) error {
	return rest.
		Update[UserRequest, User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(ValidateUser).
		Handle(func(ctx rest.Context, req UserRequest) (User, error) {
			id := ctx.Param("id")
			user := User{
				ID:    id,
				Name:  req.Name,
				Email: req.Email,
			}
			return user, nil
		}).
		Respond()
}

func DeleteUser(c *fiber.Ctx) error {
	return rest.
		Delete[UserResponse](FiberContext{Ctx: c}).
		Handle(func(ctx rest.Context, req struct{}) (struct{}, error) {
			ctx.Param("id")
			return struct{}{}, nil
		}).
		Respond()
}