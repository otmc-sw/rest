/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	db "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

type Test01Request struct {
	Name    string `json:"name"`
	TestInt int64  `json:"test_int"`
}

type Test01Response struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TestInt   int64     `json:"test_int"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateTest01(c *fiber.Ctx) error {
	return rest.
		Create[Test01Request, db.CreateTest01Params, db.Test01, Test01Response](FiberContext{Ctx: c}).
		Bind().
		Exec(func(ctx rest.Context, req Test01Request, params db.CreateTest01Params, id any) (any, error) {
			return database.CreateTest01(ctx.Context(), params)
		}).
		Respond()
}

func GetTest01(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, struct{}, db.Test01, Test01Response](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetTest01(ctx.Context(), id.(string))
		}).
		Respond()
}

func GetAllTest01s(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, struct{}, []db.Test01, []Test01Response](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetAllTest01s(ctx.Context())
		}).
		Respond()
}

func UpdateTest01(c *fiber.Ctx) error {
	return rest.
		Update[Test01Request, db.UpdateTest01Params, db.Test01, Test01Response](FiberContext{Ctx: c}).
		Bind().
		Exec(func(ctx rest.Context, req Test01Request, params db.UpdateTest01Params, id any) (any, error) {
			return database.UpdateTest01(ctx.Context(), params)
		}).
		Respond()
}

func DeleteTest01(c *fiber.Ctx) error {
	return rest.
		Delete[struct{}, struct{}, db.Test01, Test01Response](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return nil, database.DeleteTest01(ctx.Context(), id.(string))
		}).
		Respond()
}
