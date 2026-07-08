# OTMC REST

> A modern, lightweight, extensible REST toolkit for Go.

---

# Vision

OTMC REST is **not a web framework**.

It is a reusable toolkit that helps developers build clean, consistent, and maintainable REST APIs while remaining framework agnostic.

Supported frameworks:

- Fiber
- Gin
- Echo
- Chi
- net/http
- Custom adapters

The goal is to eliminate repetitive code found in almost every REST project:

- request parsing
- validation
- response formatting
- error handling
- DTO mapping
- nullable conversions
- pagination
- filtering
- middleware helpers

without replacing existing frameworks.

---

# Design Principles

## Simple

Easy to learn.

```
return rest.OK(c).
    Data(user)
```

---

## Fluent API

Everything should support method chaining.

```
return rest.
    Created(c).
    Data(document).
    Map[DocumentResponse]()
```

---

## Framework Agnostic

Core library must not depend on Fiber, Gin or Echo.

Adapters are separated.

```
rest-core

        ↑

rest-fiber
rest-gin
rest-echo
rest-http
```

---

## Generic First

Use Go Generics whenever possible.

```
request.Bind[CreateUserRequest](ctx)
```

```
mapper.Map[UserResponse](user)
```

---

## Zero Reflection (when possible)

Reflection should only be used where absolutely necessary.

Mapping should support:

- Generic
- Manual Mapper
- Reflection Mapper (optional)

---

## Minimal

Only solve common REST problems.

Do not become another web framework.

---

# Architecture

```
REST API

                    │

             Request Adapter

                    │

              Request Builder

                    │

               Validator

                    │

               Service Layer

                    │

               Repository

                    │

                 Mapper

                    │

            Response Builder

                    │

             HTTP Response
```

---

# Package Structure

```
rest/

    request/

    response/

    mapper/

    validator/

    nullable/

    convert/

    errors/

    pagination/

    filter/

    jsonx/

    middleware/

    adapters/

        fiber/

        gin/

        echo/

        http/

internal/
```

---

# Module Responsibilities

---

## request

Responsible for parsing requests.

### Features

- Path Param
- Query
- Header
- Cookie
- Body
- Form
- Multipart
- Validation integration

Example

```
id, err := request.ParamInt64(ctx, "id")

req, err := request.Bind[CreateUserRequest](ctx)
```

---

## response

Standard REST response builder.

Example

```
return response.
    OK(ctx).
    Data(user)
```

```
return response.
    Created(ctx).
    Data(document)
```

```
return response.
    NoContent(ctx)
```

---

## mapper

Convert Entity ↔ DTO.

Example

```
mapper.Register(
    func(User) UserResponse
)
```

```
mapper.Map[UserResponse](user)
```

```
mapper.MapSlice[UserResponse](users)
```

---

## validator

Validation helpers.

Example

```
validator.

    Required(req.Name).

    Max(req.Name,100).

    Validate()
```

---

## nullable

Convert pointers into sql.NullXXX.

Example

```
nullable.String(req.Name)

nullable.Int64(req.ParentID)

nullable.Time(req.CreatedAt)
```

Default value

```
nullable.

    String(req.Status).

    Default("draft")
```

---

## convert

Type conversion helpers.

Example

```
convert.Int64(str)

convert.String(nullString)

convert.Time(nullTime)

convert.Bool(nullBool)
```

---

## errors

Standard REST errors.

Example

```
return errors.

    BadRequest().

    Summary("Validation failed").

    Detail(err).

    Send(ctx)
```

---

## jsonx

JSON helper utilities.

Example

```
jsonx.SQL(raw)

jsonx.Marshal(data)

jsonx.Unmarshal(raw)
```

---

## pagination

Pagination builder.

```
page := pagination.

    New(ctx).

    DefaultSize(20)

page.Offset()

page.Limit()
```

---

## filter

Query filtering.

```
GET

/users

?page=1

&size=20

&sort=name

&order=asc

&keyword=abc
```

---

## middleware

Reusable middleware.

Examples

- Request ID
- Logger
- Recovery
- CORS
- Timeout
- Rate Limit

---

# Adapter Layer

Core library knows nothing about Fiber.

```
rest-core

↓

Context interface

↓

Fiber Adapter
```

Example

```
ctx := fiberadapter.Wrap(c)
```

Then

```
request.Bind(ctx)

response.OK(ctx)
```

Same API for every framework.

---

# Mapping System

Three mapping strategies.

## Manual

```
mapper.Register(func(User) UserResponse {})
```

Recommended.

---

## Generic

```
mapper.Map[UserResponse](user)
```

---

## Reflection

Optional.

```
mapper.Auto(user)
```

---

# Standard Response

Success

```
{
    "success": true,
    "data": {}
}
```

Error

```
{
    "success": false,
    "error": {}
}
```

---

# Error Builder

```
return response.

    Error(ctx).

    BadRequest().

    Summary("Validation failed").

    Detail(err)
```

---

# Request Flow

```
HTTP Request

↓

Bind

↓

Validate

↓

Business

↓

Map

↓

Response
```

---

# Example Handler

```
func (h *Handler) CreateDocument(c *fiber.Ctx) error {

    id, err := request.ParamInt64(c, "id")
    if err != nil {
        return err
    }

    req, err := request.Bind[CreateDocumentRequest](c)
    if err != nil {
        return err
    }

    doc, err := h.documentService.Create(
        c.Context(),
        id,
        req,
    )
    if err != nil {
        return err
    }

    return response.

        Created(c).

        Data(doc).

        Map[DocumentResponse]()
}
```

---

# Long-Term Roadmap

## v1.0

- Request
- Response
- Errors
- Nullable
- Convert
- JSON
- Validator

---

## v1.1

- Mapper
- Pagination
- Filter
- Generic Mapping

---

## v1.2

- Fiber Adapter
- Gin Adapter
- Echo Adapter
- net/http Adapter

---

## v1.3

- Middleware
- Request ID
- Recovery
- Logger integration

---

## v2.0

- OpenAPI generation
- Swagger integration
- DTO code generation
- Mapper code generation
- Validation tags
- Plugin system

---

# Project Goals

- Reduce boilerplate by more than 80%
- Consistent API responses
- Framework independent
- Generic-first design
- High performance
- Easy to extend
- Clean Architecture friendly
- Production ready