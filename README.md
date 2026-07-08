# 🚀 OTMC REST

A modern, lightweight, extensible REST toolkit for Go — **100% framework independent**.

## 👁️ Vision

OTMC REST is **not a web framework** and it does **not depend on one**. It is a reusable
toolkit that helps developers build clean, consistent, and maintainable REST APIs while
remaining framework agnostic.

The core library only knows about generic REST concepts. To talk to a transport layer
(Fiber, Gin, Echo, Chi, net/http, …) you provide your own implementation of the
[`context.Context`](context/context.go) interface — a tiny adapter. The core never imports
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

## 📦 Installation

```bash
go get github.com/otmc-sw/rest
```

The module has **zero external dependencies**.

## ⚡ Quick Start

### Define a Context adapter (once, per framework)

```go
import restcontext "github.com/otmc-sw/rest/context"

// Suppose you use Fiber. Implement the Context interface for *fiber.Ctx:
type FiberContext struct{ *fiber.Ctx }

func (c FiberContext) Context() context.Context { return c.Ctx.Context() }
func (c FiberContext) Param(k string) string     { return c.Ctx.Params(k) }
func (c FiberContext) Query(k string) string     { return c.Ctx.Query(k) }
func (c FiberContext) QueryAll(k string) []string { return []string{c.Ctx.Query(k)} }
func (c FiberContext) Header(k string) string    { return c.Ctx.Get(k) }
func (c FiberContext) Cookie(k string) string    { return c.Ctx.Cookies(k) }
func (c FiberContext) Body() io.Reader           { return bytes.NewReader(c.Ctx.Body()) }
func (c FiberContext) Bind(v interface{}) error  { return c.Ctx.BodyParser(v) }
func (c FiberContext) JSON(code int, body interface{}) error {
    return c.Ctx.Status(code).JSON(body)
}
func (c FiberContext) Status(code int)            { c.Ctx.Status(code) }
func (c FiberContext) SetHeader(k, v string)      { c.Ctx.Set(k, v) }
func (c FiberContext) Method() string             { return c.Ctx.Method() }
func (c FiberContext) Path() string               { return c.Ctx.Path() }
func (c FiberContext) String() (string, error)    { return string(c.Ctx.Body()), nil }
func (c FiberContext) Bytes() ([]byte, error)     { return c.Ctx.Body(), nil }
```

### Use the framework-agnostic API

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func CreateUser(c *fiber.Ctx) error {
    ctx := FiberContext{Ctx: c} // implements restcontext.Context

    // Parse & bind request
    var req CreateUserRequest
    if err := request.Bind(ctx, &req); err != nil {
        return errors.
            BadRequest().
            Summary("Invalid request body").
            Detail(err).
            Send(ctx)
    }

    // Validate
    if err := validator.New().
        Required(req.Name).
        Max(req.Name, 100).
        Email(req.Email).
        Validate(); err != nil {
        return errors.
            BadRequest().
            Summary("Validation failed").
            Detail(err).
            Send(ctx)
    }

    // Business logic here...

    // Send response with fluent API
    return response.
        Created[User](ctx).
        Data(user).
        Send()
}
```

## 📦 Packages

### 📥 Request

Request parsing helpers depend only on the `Context` interface.

```go
import "github.com/otmc-sw/rest/request"

request.Param(ctx, "id")
request.ParamInt64(ctx, "id")
request.Query(ctx, "page")
request.QueryInt64OrDefault(ctx, "page", 1)
request.QueryBool(ctx, "active")
request.Header(ctx, "Authorization")
request.GetBearerToken(ctx)
request.Bind(ctx, &req)
request.JSON(ctx, &req)
```

### 📤 Response

Standard REST response builder with fluent API and generic type safety.

```go
import "github.com/otmc-sw/rest/response"

response.OK[User](ctx).Data(user).Send()
response.Created[Document](ctx).Data(document).Send()
response.Accepted[Task](ctx).Data(task).Send()
response.NoContent[any](ctx).Send()
```

### ❌ Errors

Standard REST error types with detailed information, sent through the context.

```go
import "github.com/otmc-sw/rest/errors"

errors.New().
    BadRequest().
    Summary("Validation failed").
    Detail("Email is required").
    Send(ctx)
```

### 🔀 Mapper

Convert Entity ↔ DTO with generics (no reflection needed for registered types).

```go
import "github.com/otmc-sw/rest/mapper"

mapper.Register(func(u User) UserResponse {
    return UserResponse{ID: u.ID, Name: u.Name}
})

res := mapper.Map[UserResponse](user)
list := mapper.MapSlice[UserResponse](users)
```

### ✅ Validator

Fluent validation helpers — independent from HTTP.

```go
import "github.com/otmc-sw/rest/validator"

validator.New().
    Required(req.Name).
    Min(req.Name, 3).
    Max(req.Name, 100).
    Email(req.Email).
    Validate()
```

### 🔀 Nullable

Convert values to lightweight nullable types. No SQL driver dependency.

```go
import "github.com/otmc-sw/rest/nullable"

nullable.String(req.Name)
nullable.StringPtr(req.Description)
nullable.Int64(req.ParentID)
nullable.Float64(req.Price)
nullable.Bool(req.IsActive)
nullable.Time(req.CreatedAt)

nullable.NewStringBuilder(req.Status).Default("draft")
```

### 🔄 Convert

Type conversion helpers.

```go
import "github.com/otmc-sw/rest/convert"

convert.Int64(str)
convert.String(nullString)
convert.Time(nullTime)
convert.Bool(nullBool)
convert.Int64FromNull(nullInt64)
convert.Float64FromNull(nullFloat64)
```

### 📋 JSONx

JSON helper utilities.

```go
import "github.com/otmc-sw/rest/jsonx"

jsonx.Marshal(data)
jsonx.Unmarshal(raw, &v)
jsonx.SQL(raw)
jsonx.Valid(raw)
jsonx.ParseJSONOrNull(s)
```

### 📄 Pagination

```go
import "github.com/otmc-sw/rest/pagination"

page := pagination.New(ctx).DefaultSize(20).Page()
offset := page.Offset()
limit := page.Limit()
meta := pagination.NewMeta(page, totalCount)
```

### 🔍 Filter

```go
import "github.com/otmc-sw/rest/filter"

f := filter.New(ctx).Build()
// f.Keyword, f.Sort, f.Order, f.Page, f.Size
```

### 🛡️ Middleware

Framework-agnostic middleware helpers operating on `context.Context`.

```go
import "github.com/otmc-sw/rest/middleware"

mw := middleware.RequestID("X-Request-ID")
mw := middleware.Logger()
mw := middleware.Recover()
mw := middleware.Timeout(30 * time.Second)
```

### 🚦 Pipeline

The REST pipeline orchestrates the full request lifecycle using only the
`Context` interface.

```go
import "github.com/otmc-sw/rest/pipeline"

func UpdateDocument(ctx restcontext.Context) error {
    return pipeline.
        Update[UpdateRequest, DocumentResponse](ctx).
        Param("id").
        Bind().
        Validate(func(r UpdateRequest) error {
            return validator.New().Required(r.Title).Validate()
        }).
        HandleWithID(service.Update).
        Respond()
}
```

### 🎯 Context

The minimal framework-agnostic context interface that every package depends on.

```go
import "github.com/otmc-sw/rest/context"

type Context interface {
    Context() context.Context

    Param(key string) string
    Query(key string) string
    QueryAll(key string) []string
    Header(key string) string
    Cookie(key string) string

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
return response.Created[Document](ctx).Data(document).Send()
```

### 🔗 Fluent API

Everything supports method chaining with generic type safety.

```go
return response.
    Created[DocumentResponse](ctx).
    Data(document).
    Send()
```

### 🌐 Framework Agnostic

The core library does **not** depend on Fiber, Gin, Echo, Chi or net/http. There is
no `adapters/` package inside this module — adapters live in your application or in a
separate module you control.

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
request.Bind(ctx, &req)
mapper.Map[UserResponse](user)
response.OK[User](ctx).Data(user).Send()
```

### 🚫 No Reflection Required

Reflection is only used as a fallback in `mapper.Auto` when no explicit mapping is
registered. Prefer `mapper.Register` with generics.

### 🎯 Minimal

Only solve common REST problems. Don't become another web framework.

## 📜 License

* Apache License 2.0
* Copyright (c) 2026 OTMC Softwares.

## ✨ Contributors

* 🌿 Nguyen Van Trung
* 🌿 Nguyen Thi Hoai
* 🌿 OTMC Contributors