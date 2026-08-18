# 🚀 REST API Migration Guide & AI Execution Plan

Tài liệu này hướng dẫn từng bước refactor toàn bộ backend API legacy sang kiến trúc `handlers_v2` sử dụng thư viện `@otmc-sw/rest`.

---

# 📋 Rules & Standards cho AI (Yêu cầu bắt buộc)

## Max 8 Lines/API
Mỗi handler/function API không được vượt quá **8 dòng code** (tính từ khai báo `func` đến `Respond()`).

## Method Chaining
Bắt buộc áp dụng Fluent API pattern của `rest`:

```text
rest.Verb[Req, Params, DB, Res](Ctx)
    .Bind()
    .Validate()
    .Exec()
    .Respond()
```

## Safety First

- Không xóa thư mục `handlers` cũ.
- Chỉ chuyển sang `handlers_v2` sau khi:
  - Đã tích hợp vào `main.go`.
  - `go build` thành công.
  - Toàn bộ test pass.

---

# 🛠️ Step-by-Step Execution Plan

## Step 1: Cài đặt Dependency & Khởi tạo handlers_v2

### Cài đặt thư viện

```bash
cd backend

go get github.com/otmc-sw/rest@latest
```

### Tạo handlers_v2

```bash
cp -r handlers handlers_v2
```

### Tạo `handlers_v2/context.go`

```go
package handlers_v2

import (
	"bytes"
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	sqlc "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

type FiberContext struct {
	*fiber.Ctx
}

func init() {
	rest.Debug(true)
}

var database *sqlc.Queries

func New(db *sqlc.Queries) {
	database = db
}

func (c FiberContext) Context() context.Context { return c.Ctx.Context() }
func (c FiberContext) Param(key string) string  { return c.Ctx.Params(key) }
func (c FiberContext) Query(key string) string  { return c.Ctx.Query(key) }

func (c FiberContext) QueryAll(key string) []string {
	v := c.Ctx.Query(key)
	if v == "" {
		return nil
	}
	return []string{v}
}

func (c FiberContext) Header(key string) string { return c.Ctx.Get(key) }
func (c FiberContext) Cookie(key string) string { return c.Ctx.Cookies(key) }

func (c FiberContext) Body() io.Reader {
	return bytes.NewReader(c.Ctx.Body())
}

func (c FiberContext) Bind(v interface{}) error {
	return c.Ctx.BodyParser(v)
}

func (c FiberContext) JSON(code int, body interface{}) error {
	return c.Ctx.Status(code).JSON(body)
}

func (c FiberContext) Status(code int) {
	c.Ctx.Status(code)
}

func (c FiberContext) SetHeader(key, value string) {
	c.Ctx.Set(key, value)
}

func (c FiberContext) Method() string {
	return c.Ctx.Method()
}

func (c FiberContext) Path() string {
	return c.Ctx.Path()
}

func (c FiberContext) String() (string, error) {
	return string(c.Ctx.Body()), nil
}

func (c FiberContext) Bytes() ([]byte, error) {
	return c.Ctx.Body(), nil
}

func (c FiberContext) SendFile(path string) error {
	return c.Ctx.SendFile(path)
}

func (c FiberContext) Download(path, name string) error {
	return c.Ctx.Download(path, name)
}

func (c FiberContext) HTML(html string) error {
	return c.Ctx.Type("html").SendString(html)
}

func (c FiberContext) Text(text string) error {
	return c.Ctx.Type("text").SendString(text)
}

func (c FiberContext) Redirect(location string) error {
	return c.Ctx.Redirect(location)
}

func (c FiberContext) Stream(reader io.Reader) error {
	return c.Ctx.SendStream(reader)
}

func (c FiberContext) FormFile(key string) (interface{}, error) {
	return c.Ctx.FormFile(key)
}
```

---

## Step 2: Cập nhật SQL Queries (sqlc)

### Update `.sql`

- Chuẩn hóa SQL theo `sqlc`
- Dùng:

```sql
COALESCE(sqlc.narg(...), ...)
```

để hỗ trợ Partial Update.

- `CREATE` và `UPDATE` phải sử dụng:

```sql
RETURNING *
```

### Generate SQLC

```bash
sqlc generate
```

### Ví dụ `users.sql`

```sql
-- name: CreateUser :one
INSERT INTO users (...)
VALUES (...)
RETURNING *;

-- name: GetUser :one
SELECT ...
FROM users
WHERE id = ?;

-- name: GetAllUsers :many
SELECT ...
FROM users;

-- name: UpdateUser :one
UPDATE users
SET
    username = COALESCE(sqlc.narg(username), username),
    ...
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;
```

---

## Step 3: Refactor Controllers sang handlers_v2

### Entity Handler

```go
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

```

Các API khác (`GetUser`, `GetAllUsers`, `UpdateUser`, `DeleteUser`) refactor theo cùng pattern.

---

### File Handlers

```go
func DownloadFile(c *fiber.Ctx) error {
	idInt, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	return rest.
		Download(FiberContext{Ctx: c}).
		Source(fmt.Sprintf("./data/files/download_%d.txt", idInt)).
		Respond()
}
```

```go
func UploadFile(c *fiber.Ctx) error {
	var file rest.File

	parentIdInt, _ := strconv.ParseInt(c.Query("parent_id"), 10, 64)

	return rest.
		Upload(FiberContext{Ctx: c}).
		Destination(fmt.Sprintf("./data/files/dir_%d/%s", parentIdInt, file.Name)).
		Bind(&file).
		Respond()
}
```

```go
func ReadFileContent(c *fiber.Ctx) error {
	idInt, _ := strconv.ParseInt(c.Params("id"), 10, 64)

	return rest.
		ReadFileContent(FiberContext{Ctx: c}).
		Source(fmt.Sprintf("./data/files/download_%d.txt", idInt)).
		Respond()
}
```

---

## Step 4: Update `main.go`

```go
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"your-project/db"
	handlers "your-project/handlers_v2"
)

func main() {
	app := fiber.New()

	queries := db.InitSQLC()
	handlers.New(queries)

	api := app.Group("/api/v2")

	api.Post("/users", handlers.CreateUser)
	api.Get("/users", handlers.GetAllUsers)
	api.Get("/users/:id", handlers.GetUser)
	api.Put("/users/:id", handlers.UpdateUser)
	api.Delete("/users/:id", handlers.DeleteUser)

	api.Get("/files/:id/download", handlers.DownloadFile)
	api.Post("/files/upload", handlers.UploadFile)
	api.Get("/files/:id/content", handlers.ReadFileContent)

	log.Fatal(app.Listen(":3000"))
}
```

---

# ✅ Validation Checklist

Sau khi hoàn thành, AI phải xác nhận:

- [ ] `go build ./...` chạy thành công.
- [ ] Tất cả handler trong `handlers_v2/*.go` ≤ **8 dòng code**.
- [ ] Không còn `BodyParser()` hoặc parse JSON thủ công trong handlers.
- [ ] Không còn validate thủ công, sử dụng `rest.Validator()`.
- [ ] SQL đã dùng `RETURNING *`.
- [ ] SQL Update đã dùng `COALESCE(sqlc.narg(...))`.
- [ ] `sqlc generate` chạy thành công.
- [ ] API response của v2 tương thích với v1.
- [ ] Integration Test hoặc Postman Collection pass 100%.