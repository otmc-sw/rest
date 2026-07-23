/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	db "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

type ProfileResponse struct {
	Bio string `json:"bio,omitempty"`
}

type UserRequest struct {
	Username        *string          `json:"username"`
	FullName        *string          `json:"full_name,omitempty"`
	Email           *string          `json:"email"`
	Enabled         *bool            `json:"enabled"`
	TestInt         *int64           `json:"test_int"`
	Content         *json.RawMessage `json:"content,omitempty"`
	TestStringArray *[]string        `json:"test_string_array,omitempty"`
	TestIntArray    *[]int           `json:"test_int_array,omitempty"`
	TestMap         *map[string]int  `json:"test_map,omitempty"`
	TestJson        *json.RawMessage `json:"test_json,omitempty"`
	Profile         *ProfileRequest  `json:"profile"`
}

type ProfileRequest struct {
	Bio string `json:"bio,omitempty"`
}

type UserResponse struct {
	ID              int64           `json:"id"`
	Username        string          `json:"username"`
	FullName        string          `json:"full_name,omitempty"`
	Email           string          `json:"email"`
	Enabled         bool            `json:"enabled"`
	TestInt         int64           `json:"test_int"`
	AppendField     string          `json:"append_field"`
	Content         json.RawMessage `json:"content,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	TestStringArray []string        `json:"test_string_array,omitempty"`
	TestIntArray    []int           `json:"test_int_array,omitempty"`
	TestMap         map[string]int  `json:"test_map,omitempty"`
	TestJson        json.RawMessage `json:"test_json,omitempty"`
	Profile         ProfileResponse `json:"profile"`
}

func ValidateUser(r UserRequest) error {
	return rest.Validator().
		Required(r.Username).
		Email(r.Email).
		Process()
}

func generatPostUsername(id int64) string {
	return fmt.Sprintf("default_post_user_%d", id)
}

func init() {
	rest.Configure(func(c *rest.Config) {
		c.Post().SetFieldFunc("Username", func(res any) any {
			id := rest.GetFieldInt64(res, "ID")
			return generatPostUsername(id)
		})
	})
}

func CustomFields() map[string]any {
	return map[string]any{
		"Username": "custom_fields_username",
	}
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
		SetField("AppendField", "append_value").
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
			return database.UpdateUser(ctx.Context(), params)
		}).
		Respond()
}

func PatchUser(c *fiber.Ctx) error {
	return rest.
		Patch[UserRequest, db.UpdateUserParams, db.User, UserResponse](FiberContext{Ctx: c}).
		Bind().
		Exec(func(ctx rest.Context, req UserRequest, params db.UpdateUserParams, id any) (any, error) {
			return database.UpdateUser(ctx.Context(), params)
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

func SetFields(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, struct{}, db.User, UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetUser(ctx.Context(), id.(int64))
		}).
		SetFields(CustomFields()).
		Respond()
}

func SetField(c *fiber.Ctx) error {
	return rest.
		Get[struct{}, struct{}, db.User, UserResponse](FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetUser(ctx.Context(), id.(int64))
		}).
		SetField("Username", "set_field_value").
		Respond()
}

func TestResponse(c *fiber.Ctx) error {
	data := map[string]any{
		"status":    "success",
		"test_data": "Hello World",
	}
	return rest.OK(FiberContext{Ctx: c}).Data(data).Message("OK").Send()
}

func DownloadReport(c *fiber.Ctx) error {
	var file rest.File

	return rest.
		Download(FiberContext{Ctx: c}).
		Source("/tmp/report.pdf").
		Bind(&file).
		After(func(ctx rest.Context, f *rest.File) error {
			return nil
		}).
		Respond()
}

func PreviewImage(c *fiber.Ctx) error {
	return rest.
		Download(FiberContext{Ctx: c}).
		Source("/tmp/image.png").
		Respond()
}

func SendLogo(c *fiber.Ctx) error {
	return rest.
		Download(FiberContext{Ctx: c}).
		Source("./assets/logo.svg").
		Respond()
}

func Login(c *fiber.Ctx) error {
	return rest.
		Raw(FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context) error {
			return ctx.Redirect(
				"/dashboard",
			)
		}).
		Respond()
}

func Home(c *fiber.Ctx) error {
	return rest.
		Raw(FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context) error {
			return ctx.HTML(`
				<h1>Hello</h1>
			`)
		}).
		Respond()
}

func StreamData(c *fiber.Ctx) error {
	return rest.
		Raw(FiberContext{Ctx: c}).
		Exec(func(ctx rest.Context) error {
			return ctx.Stream(nil)
		}).
		Respond()
}

func UploadAvatar(c *fiber.Ctx) error {
	var file rest.File

	return rest.
		Upload(FiberContext{Ctx: c}).
		Destination("./uploads/avatars").
		Bind(&file).
		Before(func(ctx rest.Context, f *rest.File) error {
			if f.ContentType != "image/jpeg" && f.ContentType != "image/png" {
				return rest.BadRequest("Invalid file type", fmt.Errorf("only JPEG and PNG images are allowed"))
			}
			if f.Size > 5*1024*1024 {
				return rest.BadRequest("File too large", fmt.Errorf("maximum file size is 5MB"))
			}
			return nil
		}).
		After(func(ctx rest.Context, f *rest.File) error {
			return nil
		}).
		Respond()
}
