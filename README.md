# OTMC REST

A modern, lightweight, extensible REST toolkit for Go.

## Vision

OTMC REST is **not a web framework**. It is a reusable toolkit that helps developers build clean, consistent, and maintainable REST APIs while remaining framework agnostic.

### Supported Frameworks

- Fiber
- Gin (planned)
- Echo (planned)
- Chi (planned)
- net/http (planned)
- Custom adapters

The goal is to eliminate repetitive code found in almost every REST project:

- Request parsing
- Validation
- Response formatting
- Error handling
- DTO mapping
- Nullable conversions
- Pagination (planned)
- Filtering (planned)
- Middleware helpers (planned)

without replacing existing frameworks.

## Installation

```bash
go get github.com/otmc-sw/rest
```

## Quick Start

### Using with Fiber

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/otmc-sw/rest/adapters/fiber"
    "github.com/otmc-sw/rest/request"
)

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func CreateUser(c *fiber.Ctx) error {
    // Parse request
    var req CreateUserRequest
    if err := request.Bind(fiber.Wrap(c), &req); err != nil {
        return fiber.BadRequest(c, "Invalid request body", err)
    }

    // Business logic here...

    // Send response
    return fiber.OK(c, user)
}
```

## Packages

### Response

Standard REST response builder with fluent API.

```go
import "github.com/otmc-sw/rest/response"

// Success responses
response.OK().Data(user)
response.Created().Data(document)
response.Accepted().Data(task)
response.NoContent()

// Error responses
response.Error().BadRequest().Summary("Validation failed").Detail(err)
response.Error().NotFound().Summary("User not found")
response.Error().InternalError().Summary("Database error").Detail(err)
```

### Errors

Standard REST error types with detailed information.

```go
import "github.com/otmc-sw/rest/errors"

errors.New().
    BadRequest().
    Summary("Validation failed").
    Detail("Email is required").
    Request(req).
    Build()
```

Error structure includes:
- `code`: HTTP status code
- `key`: Error key (e.g., BAD_REQUEST, NOT_FOUND)
- `type`: Error type (e.g., Bad Request, Not Found)
- `summary`: User-friendly summary
- `detail`: Detailed error message
- `reason`: Failure reason
- `request`: Request body (if available)
- `data`: Additional error data
- `file`, `line`, `function`: Source location
- `timestamp`: ISO timestamp

### Nullable

Convert pointers to sql.NullXXX types.

```go
import "github.com/otmc-sw/rest/nullable"

nullable.String(req.Name)
nullable.StringPtr(req.Description)
nullable.Int64(req.ParentID)
nullable.Float64(req.Price)
nullable.Bool(req.IsActive)

// With default value
nullable.NewStringBuilder(req.Status).Default("draft")
```

### Convert

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

### JSONx

JSON helper utilities.

```go
import "github.com/otmc-sw/rest/jsonx"

jsonx.Marshal(data)
jsonx.Unmarshal(raw, &v)
jsonx.SQL(raw)
jsonx.Valid(raw)
jsonx.ParseJSONOrNull(s)
```

### Validator

Fluent validation helpers.

```go
import "github.com/otmc-sw/rest/validator"

validator.New().
    Required(req.Name).
    Min(req.Name, 3).
    Max(req.Name, 100).
    Email(req.Email).
    Validate()

// Password validation
validator.New().
    Required(req.Password).
    Min(req.Password, 8).
    HasUpperCase(req.Password).
    HasLowerCase(req.Password).
    HasDigit(req.Password).
    HasSpecialChar(req.Password).
    Validate()
```

### Request

Request parsing helpers.

```go
import "github.com/otmc-sw/rest/request"

// Path parameters
request.Param(ctx, "id")
request.ParamInt64(ctx, "id")

// Query parameters
request.Query(ctx, "page")
request.QueryInt64(ctx, "page")
request.QueryInt64OrDefault(ctx, "page", 1)
request.QueryBool(ctx, "active")

// Headers
request.Header(ctx, "Authorization")
request.GetBearerToken(ctx)
request.GetContentType(ctx)
request.GetUserAgent(ctx)

// Body
request.Bind(ctx, &req)
request.String(ctx)
request.Bytes(ctx)
request.JSON(ctx, &req)
```

### Context

Framework-agnostic context interface.

```go
import "github.com/otmc-sw/rest/context"

// The context interface is implemented by adapters
// to provide a unified API across frameworks
type Context interface {
    GetContext() context.Context
    Param(key string) string
    Query(key string) string
    QueryAll(key string) []string
    Header(key string) string
    Cookie(key string) string
    Body() io.Reader
    Bind(v interface{}) error
    Method() string
    Path() string
    String() (string, error)
    Bytes() ([]byte, error)
}
```

### Adapters

Framework-specific adapters implement the context interface.

#### Fiber Adapter

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/otmc-sw/rest/adapters/fiber"
    "github.com/otmc-sw/rest/request"
)

func Handler(c *fiber.Ctx) error {
    // Wrap Fiber context
    ctx := fiber.Wrap(c)

    // Use framework-agnostic request helpers
    id, err := request.ParamInt64(ctx, "id")
    if err != nil {
        return fiber.BadRequest(c, "Invalid ID", err)
    }

    // Send response using adapter helpers
    return fiber.OK(c, data)
}
```

Fiber adapter also provides convenience response helpers:

```go
fiber.OK(c, data)
fiber.Created(c, data)
fiber.Accepted(c, data)
fiber.NoContent(c)
fiber.BadRequest(c, "summary", detail)
fiber.Unauthorized(c, "message")
fiber.Forbidden(c, "message")
fiber.NotFound(c, "message")
fiber.Conflict(c, "message")
fiber.InternalError(c, "message", err)
```

## Design Principles

### Simple

Easy to learn and use.

```go
return fiber.OK(c).Data(user)
```

### Fluent API

Everything supports method chaining.

```go
return response.
    Created(c).
    Data(document).
    Map[DocumentResponse]()
```

### Framework Agnostic

Core library doesn't depend on Fiber, Gin, or Echo. Adapters are separated.

```
rest-core
    ↓
Context interface
    ↓
Fiber Adapter
```

### Generic First

Use Go Generics whenever possible.

```go
request.Bind[CreateUserRequest](ctx)
mapper.Map[UserResponse](user)
```

### Zero Reflection

Reflection only used where absolutely necessary. Mapping supports:
- Generic
- Manual Mapper
- Reflection Mapper (optional, planned)

### Minimal

Only solve common REST problems. Don't become another web framework.

## Example Handler

```go
func (h *Handler) CreateDocument(c *fiber.Ctx) error {
    // Parse path parameter
    id, err := request.ParamInt64(fiber.Wrap(c), "id")
    if err != nil {
        return fiber.BadRequest(c, "Invalid ID", err)
    }

    // Parse request body
    var req CreateDocumentRequest
    if err := request.Bind(fiber.Wrap(c), &req); err != nil {
        return fiber.BadRequest(c, "Invalid request body", err)
    }

    // Validate
    if err := validator.New().
        Required(req.Title).
        Min(req.Title, 3).
        Max(req.Title, 100).
        Validate(); err != nil {
        return fiber.BadRequest(c, "Validation failed", err)
    }

    // Business logic
    doc, err := h.documentService.Create(c.Context(), id, req)
    if err != nil {
        return fiber.InternalError(c, "Failed to create document", err)
    }

    // Send response
    return fiber.Created(c, doc)
}
```

## License

OTMC License

## Contributing

Contributions are welcome! Please see the NOTICE file for details.
