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

## 📦 Installation

```bash
go get github.com/otmc-sw/rest
```

The module has **zero external dependencies**.

## ⚡ Quick Start

### One import only

Applications import a **single package**. All internal building blocks
(`request`, `response`, `validator`, `mapper`, `errors`, `nullable`, …) are
implementation details and should **never** be imported by application code.

```go
import "github.com/otmc-sw/rest"
```

### Define a Context adapter (once, per framework)

```go
import "github.com/otmc-sw/rest"

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

### Handlers describe only the business flow

The pipeline coordinates everything: **Request → Bind → Validate → Business → Map → Respond**.
You only declare the flow.

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Service owns the business logic and returns an *entity* (not a DTO).
func (s *UserService) Create(ctx rest.Context, req CreateUserRequest) (User, error) {
    // load, validate business rules, persist, return entity
}

func CreateUser(c FiberContext) error {
    return rest.
        Create[CreateUserRequest, User, UserResponse](c).
        Bind().
        Validate(func(r CreateUserRequest) error {
            return rest.Validate().Required(r.Name).Email(r.Email).Validate()
        }).
        Handle(userService.Create).
        Respond()
}
```

`Respond()` automatically maps the returned `User` entity to `UserResponse`
(via a registered mapper) and writes a `201 Created` JSON response.

### Update with a route id

```go
func UpdateDocument(c FiberContext) error {
    return rest.
        Update[UpdateDocumentRequest, Document, DocumentResponse](c).
        Param("id").
        Bind().
        HandleWithID(documentService.Update).
        Respond()
}
```

### Get / Delete

```go
func GetDocument(c FiberContext) error {
    return rest.
        Get[struct{}, Document, DocumentResponse](c).
        Param("id").
        HandleWithID(documentService.Get).
        Respond()
}

func DeleteDocument(c FiberContext) error {
    return rest.
        Delete[DocumentResponse](c).
        Param("id").
        HandleWithID(documentService.Delete).
        Respond()
}
```

### Register a mapper (once, at startup)

```go
func init() {
    rest.Register(func(u User) UserResponse {
        return UserResponse{ID: u.ID, Name: u.Name}
    })
}
```

## 📦 Packages (internal)

These packages are implementation details wired together by the top-level `rest`
package. Application code does **not** import them directly.

### 🚦 Pipeline

The REST pipeline orchestrates the full request lifecycle using only the
`Context` interface. `Create`, `Get`, `Update`, `Delete` start a pipeline;
`Bind`, `Validate`, `Handle`, `HandleWithID`, `Respond` drive it.

Status codes are chosen automatically: `Create → 201`, `Get`/`Update → 200`,
`Delete → 204`. The entity returned by the service is mapped to the response
DTO by the registered mapper inside `Respond()`.

### 🔄 Mapper

Convert Entity ↔ DTO with generics (no reflection needed for registered types).

```go
import "github.com/otmc-sw/rest"

rest.Register(func(u User) UserResponse {
    return UserResponse{ID: u.ID, Name: u.Name}
})

res := rest.Map[UserResponse](user)         // exported for advanced use
list := rest.MapSlice[UserResponse](users)
```

### 📥 Request

Request parsing helpers depend only on the `Context` interface.

```go
rest.Param(ctx, "id")
rest.ParamInt64(ctx, "id")
rest.Query(ctx, "page")
rest.QueryInt64OrDefault(ctx, "page", 1)
rest.QueryBool(ctx, "active")
rest.Header(ctx, "Authorization")
rest.GetBearerToken(ctx)
rest.Bind(ctx, &req)
rest.JSON(ctx, &req)
```

### 📤 Response

Standard REST response builder with fluent API and generic type safety.

```go
rest.OK[User](ctx).Data(user).Send()
rest.Created[Document](ctx).Data(document).Send()
rest.Accepted[Task](ctx).Data(task).Send()
rest.NoContent[any](ctx).Send()
```

### ❌ Errors

Standard REST error types with detailed information, sent through the context.

```go
rest.NewError().
    BadRequest().
    Summary("Validation failed").
    Detail("Email is required").
    Send(ctx)
```

### ✅ Validator

Fluent validation helpers — independent from HTTP.

```go
rest.Validate().
    Required(req.Name).
    Min(req.Name, 3).
    Max(req.Name, 100).
    Email(req.Email).
    Validate()
```

### 🔀 Nullable

Convert values to lightweight nullable types. No SQL driver dependency.

```go
nullable.String(req.Name)
nullable.StringPtr(req.Description)
nullable.Int64(req.ParentID)
nullable.Float64(req.Price)
nullable.Bool(req.IsActive)
nullable.Time(req.CreatedAt)

nullable.NewStringBuilder(req.Status).Default("draft")
```

### 🔄 Converter

Type conversion helpers.

```go
converter.Int64(str)
converter.String(nullString)
converter.Time(nullTime)
converter.Bool(nullBool)
converter.Int64FromNull(nullInt64)
converter.Float64FromNull(nullFloat64)
```

### 📋 JSONx

JSON helper utilities.

```go
jsonx.Marshal(data)
jsonx.Unmarshal(raw, &v)
jsonx.SQL(raw)
jsonx.Valid(raw)
jsonx.ParseJSONOrNull(s)
```

### 📄 Pagination

```go
page := pagination.New(ctx).DefaultSize(20).Page()
offset := page.Offset()
limit := page.Limit()
meta := pagination.NewMeta(page, totalCount)
```

### 🔍 Filter

```go
f := filter.New(ctx).Build()
// f.Keyword, f.Sort, f.Order, f.Page, f.Size
```

### 🛡️ Middleware

Framework-agnostic middleware helpers operating on `context.Context`.

```go
mw := middleware.RequestID("X-Request-ID")
mw := middleware.Logger()
mw := middleware.Recover()
mw := middleware.Timeout(30 * time.Second)
```

### 🎯 Context

The minimal framework-agnostic context interface that every package depends on.

```go
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
return rest.Create[Req, Entity, Res](ctx).Bind().Handle(service.Create).Respond()
```

### 🔗 Fluent API

Everything supports method chaining with generic type safety.

```go
return rest.
    Update[UpdateDocumentRequest, Document, DocumentResponse](ctx).
    Param("id").
    Bind().
    HandleWithID(documentService.Update).
    Respond()
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
rest.Bind(ctx, &req)
rest.Map[UserResponse](user)
rest.OK[User](ctx).Data(user).Send()
```

### 🚫 No Reflection Required

Reflection is only used as a fallback in `mapper.Auto` when no explicit mapping is
registered. Prefer `rest.Register` with generics.

### 🎯 Minimal

Only solve common REST problems. Don't become another web framework.

## 📜 License

* Apache License 2.0
* Copyright (c) 2026 OTMC Softwares.

## ✨ Contributors

* 🌿 Nguyen Van Trung
* 🌿 Nguyen Thi Hoai
* 🌿 OTMC Contributors