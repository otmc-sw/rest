# 🚀 OTMC REST

A modern, lightweight, extensible REST toolkit for Go — **100% framework independent**.

## 👁️ Vision

OTMC REST is **not a web framework** and it does **not depend on one**. It is a reusable
toolkit that helps developers build clean, consistent, and maintainable REST APIs while
remaining framework agnostic.

The core library only knows about generic REST concepts. To talk to a transport layer
(Fiber, Gin, Echo, Chi, net/http, …) you provide your own implementation of the
[`Context`](context/context.go) interface — a tiny adapter. The core never imports
any web framework.

### 🌐 Bring Your Own Framework

- 🧶 Fiber
- 🍃 Gin
- 🔔 Echo
- 🥢 Chi
- 🌐 net/http
- 🔧 Any custom transport (CLI, gRPC gateway, test harness, …)

The core eliminates repetitive code found in almost every REST project:

- 📥 Request parsing
- ✅ Validation
- 📤 Response formatting
- ❌ Error handling
- 🔄 DTO mapping
- 🔀 Nullable conversions
- 📄 Pagination
- 🔍 Filtering
- 🛡️ Middleware helpers (framework-agnostic)

A runnable Fiber example lives in [`examples/fiber`](examples/fiber).

## 📦 Installation

```bash
go get github.com/otmc-sw/rest
```

The core module has **zero external dependencies** (sub-packages used by the example,
such as a database driver, are only needed by the example itself).

## ⚡ Quick Start

### Import the packages you need

Application code imports the top-level `rest` package for the pipeline, plus any
supporting sub-packages it actually uses:

```go
import (
    rest "github.com/otmc-sw/rest"
    "github.com/otmc-sw/rest/context"
    "github.com/otmc-sw/rest/nullable"
    "github.com/otmc-sw/rest/request"
    "github.com/otmc-sw/rest/response"
)
```

> The sub-packages (`request`, `response`, `validator`, `mapper`, `errors`,
> `nullable`, `converter`, `jsonx`, `pagination`, `filter`, `middleware`, `context`)
> are **public and importable**. The top-level `rest` package re-exports the most
> common helpers (`Validate`, `NewError`, `Register`, the pipeline constructors) for
> convenience, but you may also use the sub-packages directly.

### Define a Context adapter (once, per framework)

The core depends only on the `Context` interface. Here is the real adapter shipped
with the Fiber example ([`examples/fiber/handlers/context.go`](examples/fiber/handlers/context.go)):

```go
import (
    "bytes"
    "context"
    "io"

    "github.com/gofiber/fiber/v2"
)

type FiberContext struct {
    *fiber.Ctx
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
func (c FiberContext) Body() io.Reader          { return bytes.NewReader(c.Ctx.Body()) }
func (c FiberContext) Bind(v interface{}) error { return c.Ctx.BodyParser(v) }
func (c FiberContext) JSON(code int, body interface{}) error {
    return c.Ctx.Status(code).JSON(body)
}
func (c FiberContext) Status(code int)             { c.Ctx.Status(code) }
func (c FiberContext) SetHeader(key, value string) { c.Ctx.Set(key, value) }
func (c FiberContext) Method() string              { return c.Ctx.Method() }
func (c FiberContext) Path() string                { return c.Ctx.Path() }
func (c FiberContext) String() (string, error)     { return string(c.Ctx.Body()), nil }
func (c FiberContext) Bytes() ([]byte, error)      { return c.Ctx.Body(), nil }
```

### Handlers describe only the business flow

The pipeline coordinates everything: **Request → Bind → Validate → Params → Exec → Map → Respond**.
You only declare the flow and the business closure.

```go
type UserRequest struct {
    Username *string          `json:"username"`
    FullName *string          `json:"full_name,omitempty"`
    Email    *string          `json:"email"`
    Content  *json.RawMessage `json:"content,omitempty"`
}

type UserResponse struct {
    ID        int64       `json:"id"`
    Username  string      `json:"username"`
    FullName  string      `json:"full_name,omitempty"`
    Email     string      `json:"email"`
    Content   interface{} `json:"content,omitempty"`
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`
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
```

`Respond()` automatically maps the `db.User` entity to `UserResponse` (via a registered
mapper, or reflection fallback) and writes a `201 Created` JSON response.

The pipeline uses 4 generic type parameters: `Create[Req, Params, Entity, Res]`:
- `Req` - Request DTO (parsed from body)
- `Params` - Database/service parameters (auto-mapped from Req if no Params() step)
- `Entity` - Database entity type
- `Res` - Response DTO

The `Exec` closure signature is `func(ctx rest.Context, req Req, params Params, id any) (any, error)`:

- For `Create`/`Update`/`Patch` with `Bind()`, `req` is the parsed request DTO and `params` is
  auto-mapped from `req` (or built via `Params()` step). The `id` is the route parameter
  parsed to `int64`.
- For `Get`/`Delete`, `req` and `params` are empty structs and `id` is the route `id`
  automatically read from the `id` path parameter and parsed to `int64`.
- Return a non-`nil` value to send it as the response body; return `nil` (with a `nil`
  error) when there is no body (e.g. `Delete`).

### Get / Update / Patch / Delete

```go
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
```

### Register a mapper (optional, once at startup)

When the entity and response types differ, register a mapper so `Respond()` can convert
them. If no mapper is registered, the toolkit falls back to field-name reflection.

```go
func init() {
    rest.Register(func(u db.User) UserResponse {
        return UserResponse{ID: u.ID, Username: u.Username, Email: u.Email}
    })
}
```

The `Params` type is automatically mapped from `Req` using reflection if no explicit
`Params()` step is provided. Field names must match between `Req` and `Params` types.

### Wiring it up (Fiber)

This is the real entry point from [`examples/fiber/main.go`](examples/fiber/main.go):

```go
func main() {
    db, err := db.New()
    if err != nil {
        logger.Crit("Failed to connect to database: %v", err)
    }
    handlers.New(db.Queries)

    app := fiber.New()

    app.Use(func(c *fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        logger.Request(c.Method(), c.Path(), c.Response().StatusCode(), time.Since(start), c.IP())
        return err
    })

    app.Post("/users", handlers.CreateUser)
    app.Get("/users", handlers.GetAllUsers)
    app.Get("/users/:id", handlers.GetUser)
    app.Patch("/users/:id", handlers.UpdateUser)
    app.Delete("/users/:id", handlers.DeleteUser)

    app.Get("/test", handlers.TestResponse)

    logger.Info("Server started on :3000")
    log.Fatal(app.Listen(":3000"))
}
```

## 📦 Packages

### 🚦 Pipeline (top-level `rest` package)

The REST pipeline orchestrates the full request lifecycle using only the `Context`
interface. `Create`, `Get`, `Update`, `Patch`, `Delete` start a pipeline; `Bind`, `Validate`,
`Params`, `Exec`, `Respond` drive it.

```go
// Generic constructors exposed by the rest package:
rest.Create[Req, Params, Entity, Res](ctx)   // status 201
rest.Get[Req, Params, Entity, Res](ctx)      // status 200
rest.Update[Req, Params, Entity, Res](ctx)   // status 200
rest.Patch[Req, Params, Entity, Res](ctx)    // status 200
rest.Delete[Req, Params, Entity, Res](ctx)   // status 204

// Fluent steps:
pipeline.Bind()                             // parse the request body into Req
pipeline.Validate(func(Req) error)         // run validation
pipeline.Params(func(Req) Params)          // build params from request (optional)
pipeline.Exec(func(ctx, Req, Params, id any) (any, error)) // run business logic
pipeline.Respond()                          // map Entity -> Res and write JSON
```

Status codes are chosen automatically: `Create → 201`, `Get`/`Update`/`Patch → 200`,
`Delete → 204`. The value returned by `Exec` is mapped to the response DTO by the
registered mapper (or reflection fallback) inside `Respond()`.

For `Get`, `Update`, `Patch`, and `Delete` the route `id` is read automatically from the `id`
path parameter and parsed to `int64` before `Exec` is called.

The `Params` type is automatically mapped from `Req` using reflection if no explicit
`Params()` step is provided. You can also use `Params()` to build params manually.

The following types and helpers are re-exported by the top-level `rest` package:

```go
type Context = context.Context
type Handler[Req, Entity] = pipeline.Handler[Req, Entity]
type ExecHandler[Req] = pipeline.ExecHandler[Req]
type PatchHandler[Req, Params] = pipeline.PatchHandler[Req, Params]
type Pipeline[Req, Params, Entity, Res] = pipeline.Pipeline[Req, Params, Entity, Res]

func Create[Req, Params, Entity, Res](ctx Context) *Pipeline[Req, Params, Entity, Res]
func Get[Req, Params, Entity, Res](ctx Context) *Pipeline[Req, Params, Entity, Res]
func Update[Req, Params, Entity, Res](ctx Context) *Pipeline[Req, Params, Entity, Res]
func Patch[Req, Params, Entity, Res](ctx Context) *Pipeline[Req, Params, Entity, Res]
func Delete[Req, Params, Entity, Res](ctx Context) *Pipeline[Req, Params, Entity, Res]

func Register[Src, Dst](fn func(Src) Dst)
func Validate() *validator.Validator
func NewError() *errors.Builder

func Debug()
func DebugComponent(component string)
func DebugWithEnv()
```

### 🔄 Mapper (`github.com/otmc-sw/rest/mapper`)

Convert Entity ↔ DTO with generics (no reflection needed for registered types).

```go
import "github.com/otmc-sw/rest/mapper"

mapper.Register(func(u db.User) UserResponse {
    return UserResponse{ID: u.ID, Username: u.Username, Email: u.Email}
})

res := mapper.Map[UserResponse](user)      // exported for advanced use
list := mapper.MapSlice[UserResponse](users)
```

If no mapping is registered for a pair of types, `mapper.Auto` copies fields with
matching names via reflection. Prefer `mapper.Register` (or `rest.Register`) with
generics for full control.

### 📥 Request (`github.com/otmc-sw/rest/request`)

Request parsing helpers depend only on the `Context` interface.

```go
request.Param(ctx, "id")
request.ParamInt64(ctx, "id")
request.Query(ctx, "page")
request.QueryInt64OrDefault(ctx, "page", 1)
request.QueryInt(ctx, "page")
request.QueryBool(ctx, "active")
request.Header(ctx, "Authorization")
request.Cookie(ctx, "session")
request.Bind(ctx, &req)
request.JSON(ctx, &req)
request.GetBearerToken(ctx)
request.GetClientIP(ctx)
```

### 📤 Response (`github.com/otmc-sw/rest/response`)

Standard REST response builder with fluent API and generic type safety.

```go
import "github.com/otmc-sw/rest/response"

response.OK[User](ctx).Data(user).Send()
response.Created[Document](ctx).Data(document).Send()
response.Accepted[Task](ctx).Data(task).Send()
response.NoContent[any](ctx).Send()
response.New[User](ctx, 200).Data(user).Message("ok").Send()
```

### ❌ Errors (`github.com/otmc-sw/rest/errors`)

Standard REST error types with detailed information, sent through the context.

```go
import "github.com/otmc-sw/rest/errors"

errors.New().
    BadRequest().
    Summary("Validation failed").
    Detail("Email is required").
    Send(ctx)
```

Available status helpers: `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`,
`Conflict`, `UnprocessableEntity`, `InternalError`, `ServiceUnavailable`, or set a
custom code with `Code(429)`. The error response includes code, key, type, summary,
detail, and (when built) file/line/function for debugging.

### ✅ Validator (`github.com/otmc-sw/rest/validator`)

Fluent validation helpers — independent from HTTP. Also re-exported as `rest.Validator()`.

```go
import "github.com/otmc-sw/rest/validator"

validator.New().
    Required(req.Name).
    Min(req.Name, 3).
    Max(req.Name, 100).
    Email(req.Email).
    Validate()
```

Additional helpers: `Between`, `URL`, `Numeric`, `Alpha`, `AlphaNumeric`, `Match`,
`Equals`, `OneOf`, `MinInt`, `MaxInt`, `Positive`, `Negative`, `HasUpperCase`,
`HasLowerCase`, `HasDigit`, `HasSpecialChar`, and `Custom`.

### 🔀 Nullable (`github.com/otmc-sw/rest/nullable`)

Convert values to lightweight nullable types. No SQL driver dependency.

```go
import "github.com/otmc-sw/rest/nullable"

nullable.String(req.Name)
nullable.StringPtr(req.Description)
nullable.Int64(req.ParentID)
nullable.Int64Ptr(req.ParentID)
nullable.Float64(req.Price)
nullable.Bool(req.IsActive)
nullable.Time(req.CreatedAt)

nullable.NewStringBuilder(req.Status).Default("draft")
```

### 🔄 Converter (`github.com/otmc-sw/rest/converter`)

Type conversion helpers between primitives and nullable types.

```go
import "github.com/otmc-sw/rest/converter"

converter.Int64(str)
converter.Int64OrDefault(str, 0)
converter.String(nullString)
converter.StringPtr(nullString)
converter.Time(nullTime)
converter.Bool(nullBool)
converter.Int64FromNull(nullInt64)
converter.Float64FromNull(nullFloat64)
converter.ToNullString(s)
```

### 📋 JSONx (`github.com/otmc-sw/rest/jsonx`)

JSON helper utilities.

```go
import "github.com/otmc-sw/rest/jsonx"

jsonx.Marshal(data)
jsonx.MarshalToString(data)
jsonx.Unmarshal(raw, &v)
jsonx.UnmarshalString(s, &v)
jsonx.SQL(raw)            // wrap valid JSON as a nullable string (e.g. for SQL columns)
jsonx.Valid(raw)
jsonx.ParseJSONOrNull(s)
```

### 📄 Pagination (`github.com/otmc-sw/rest/pagination`)

```go
import "github.com/otmc-sw/rest/pagination"

page := pagination.New(ctx).DefaultSize(20).Page()
offset := page.Offset()
limit := page.Limit()
meta := pagination.NewMeta(page, totalCount)
```

Default page is `1`, default size `20`, maximum size `100`. Reads `page` and `size`
query parameters.

### 🔍 Filter (`github.com/otmc-sw/rest/filter`)

```go
import "github.com/otmc-sw/rest/filter"

f := filter.New(ctx).Build()
// f.Keyword, f.Sort, f.Order (filter.OrderAsc | filter.OrderDesc), f.Page, f.Size
```

Reads `keyword`, `sort`, `order`, `page`, and `size` query parameters.

### 🛡️ Middleware (`github.com/otmc-sw/rest/middleware`)

Framework-agnostic middleware helpers operating on `rest.Context`.

```go
import "github.com/otmc-sw/rest/middleware"

mw := middleware.RequestID("X-Request-ID") // defaults to "X-Request-ID"
mw := middleware.Logger()
mw := middleware.Recover()
mw := middleware.Timeout(30 * time.Second)
```

Each helper returns `func(ctx rest.Context, next func(ctx rest.Context) error) error`,
so it can be adapted to any framework's middleware signature.

### 🎯 Context (`github.com/otmc-sw/rest/context`)

The minimal framework-agnostic context interface that every package depends on.

```go
type Context interface {
    Context() context.Context

    Param(key string) string
    Query(key string) string
    QueryAll(key string) []string
    Header(key string) string
    Cookie(key string) string

    Body() io.Reader
    Bind(v interface{}) error
    JSON(status int, body interface{}) error
    Status(code int)
    SetHeader(key, value string)

    Method() string
    Path() string
    String() (string, error)
    Bytes() ([]byte, error)
}
```

## 🎨 Design Principles

### 💡 Simple

Easy to learn and use.

```go
return rest.Create[Req, Params, Entity, Res](ctx).Bind().Validate(validate).Exec(service).Respond()
```

### 🔗 Fluent API

Everything supports method chaining with generic type safety.

```go
return rest.
    Update[UpdateDocumentRequest, UpdateParams, Document, DocumentResponse](ctx).
    Bind().
    Validate(ValidateDocument).
    Exec(updateDocument).
    Respond()
```

### 🌐 Framework Agnostic

The core library does **not** depend on Fiber, Gin, Echo, Chi or net/http. There is
no `adapters/` package inside this module — adapters live in your application or in a
separate module you control (see [`examples/fiber`](examples/fiber)).

```
Fiber / Gin / Echo / net/http
            ↓
   Context Adapter (your code)
            ↓
   rest.Context (interface)
            ↓
   request → validator → service → mapper → response
```

### 🔷 Generic First

Use Go Generics whenever possible.

```go
rest.Create[UserRequest, db.User, UserResponse](ctx)
mapper.Map[UserResponse](user)
response.OK[User](ctx).Data(user).Send()
```

### 🚫 No Reflection Required

Reflection is only used as a fallback in `mapper.Auto` when no explicit mapping is
registered. Prefer `rest.Register` / `mapper.Register` with generics.

### 🎯 Minimal

Only solve common REST problems. Don't become another web framework.

## 📜 License

* Apache License 2.0
* Copyright (c) 2026 OTMC Softwares.

## ✨ Contributors

* 🌿 Nguyen Van Trung
* 🌿 Nguyen Thi Hoai
* 🌿 OTMC Contributors