/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	db "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

func init() {
	rest.Debug()
}

type UserRequest struct {
	Username string
	Email    string
}

type UserResponse struct {
	ID       int64  `json:"id"`
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
		Create[UserRequest, db.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(ValidateUser).
		Exec(func(ctx rest.Context, req UserRequest, id any) (any, error) {
			params := db.CreateUserParams{
				Username: req.Email,
				Email:    req.Email,
			}
			return nil, database.CreateUser(ctx.Context(), params)
		}).
		Respond()
}

func GetUser(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, db.User, UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, id any) (any, error) {
			return database.GetUser(ctx.Context(), id.(int64))
		}).
		Respond()
}

func GetAllUsers(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, []db.User, []UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, id any) (any, error) {
			return database.GetAllUsers(ctx.Context())
		}).
		Respond()
}

func UpdateUser(c *fiber.Ctx) error {
	return rest.
		Update[UserRequest, db.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(ValidateUser).
		Exec(func(ctx rest.Context, req UserRequest, id any) (any, error) {
			params := db.UpdateUserParams{
				Username: req.Email,
				Email:    req.Email,
				ID:       id.(int64),
			}
			return nil, database.UpdateUser(ctx.Context(), params)
		}).
		Respond()
}

func DeleteUser(c *fiber.Ctx) error {
	return rest.
		Delete[UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, id any) (any, error) {
			return nil, database.DeleteUser(ctx.Context(), id.(int64))
		}).
		Respond()
}
