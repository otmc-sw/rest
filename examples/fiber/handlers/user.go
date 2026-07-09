/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	sqlc "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

type User struct {
	ID       string
	Username string
	Email    string
}

type UserRequest struct {
	Username string
	Email    string
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func ValidateUser(r UserRequest) error {
	return rest.Validate().
		Required(r.Username).
		Email(r.Email).
		Validate()
}

func CreateUser(c *fiber.Ctx) error {
	return rest.
		Create[UserRequest, User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(ValidateUser).
		Exec(func(ctx rest.Context, req UserRequest) error {
			params := sqlc.CreateUserParams{
				Username: req.Email,
				Email:    req.Email,
			}
			return database.CreateUser(ctx.Context(), params)
		}).
		Respond()
}

func GetUser(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, User, UserResponse](FiberContext{Ctx: c}).
		ExecWithIDResult(func(ctx rest.Context, req struct{}, id int64) (any, error) {
			return database.GetUser(ctx.Context(), id)
		}).
		Respond()
}

func GetAllUsers(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, User, UserResponse](FiberContext{Ctx: c}).
		ExecWithIDResult(func(ctx rest.Context, req struct{}, id int64) (any, error) {
			return database.GetAllUsers(ctx.Context())
		}).
		Respond()
}

func UpdateUser(c *fiber.Ctx) error {
	return rest.
		Update[UserRequest, User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(ValidateUser).
		ExecWithID(func(ctx rest.Context, req UserRequest, id int64) error {
			params := sqlc.UpdateUserParams{
				Username: req.Email,
				Email:    req.Email,
				ID:       id,
			}
			return database.UpdateUser(ctx.Context(), params)
		}).
		Respond()
}

func DeleteUser(c *fiber.Ctx) error {
	return rest.
		Delete[UserResponse](FiberContext{Ctx: c}).
		ExecWithID(func(ctx rest.Context, req struct{}, id int64) error {
			return database.DeleteUser(ctx.Context(), id)
		}).
		Respond()
}
