/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	db "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

type UserRequest struct {
	Username *string          `json:"username"`
	FullName *string          `json:"full_name,omitempty"`
	Email    *string          `json:"email"`
	Content  *json.RawMessage `json:"content,omitempty"`
}

type UserResponse struct {
	ID       int64       `json:"id"`
	Username string      `json:"username"`
	FullName string      `json:"full_name,omitempty"`
	Email    string      `json:"email"`
	Content  interface{} `json:"content,omitempty"`
}

func ValidateUser(r UserRequest) error {
	return rest.Validator().
		Required(r.Username).
		Email(r.Email).
		Process()
}

func CreateUser(c *fiber.Ctx) error {
	return rest.
		Create[UserRequest, db.CreateUserParams, db.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Validate(ValidateUser).
		Exec(func(ctx rest.Context, req UserRequest, params db.CreateUserParams, id any) (any, error) {
			return database.CreateUser(ctx.Context(), params)
		}).
		Respond()
}

func GetUser(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, struct{}, db.User, UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetUser(ctx.Context(), id.(int64))
		}).
		Respond()
}

func GetAllUsers(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, struct{}, []db.User, []UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetAllUsers(ctx.Context())
		}).
		Respond()
}

func UpdateUser(c *fiber.Ctx) error {
	return rest.
		Update[UserRequest, db.UpdateUserParams, db.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Exec(func(ctx rest.Context, req UserRequest, params db.UpdateUserParams, id any) (any, error) {
			return nil, database.UpdateUser(ctx.Context(), params)
		}).
		Respond()
}

func PatchUser(c *fiber.Ctx) error {
	return rest.
		Patch[UserRequest, db.UpdateUserParams, db.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Exec(func(ctx rest.Context, req UserRequest, params db.UpdateUserParams, id any) (any, error) {
			return nil, database.UpdateUser(ctx.Context(), params)
		}).
		Respond()
}

func DeleteUser(c *fiber.Ctx) error {
	return rest.
		Delete[struct{}, struct{}, struct{}, UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return nil, database.DeleteUser(ctx.Context(), id.(int64))
		}).
		Respond()
}

func TestResponse(c *fiber.Ctx) error {
	data := map[string]any{
		"status":    "success",
		"test_data": "Hello World",
	}
	return rest.OK(FiberContext{Ctx: c}).Data(data).Message("OK").Send()
}
