# 🚀 OTMC — Struct-Driven REST Framework

OTMC builds database schema, repositories and REST endpoints automatically from
Go structs. Developers only define models and a `otmc.Config` — everything else
(table, migration, validation, routes) is generated.

## 📦 Installation

```bash
go get github.com/otmc-sw/rest@latest
```

## 📖 Usage

### 1. Define models

```go
package models

// Role is migrated to a "roles" table and exposed at /api/roles.
type Role struct {
	ID          uint64 `json:"id" otmc:"primary;auto"`
	Name        string `json:"name" otmc:"required;unique"`
	Description string `json:"description"`
}

// User is migrated to a "users" table and exposed at /api/users.
type User struct {
	ID     uint64 `json:"id" otmc:"primary;auto"`
	Name   string `json:"name" otmc:"required"`
	Email  string `json:"email" otmc:"required;unique"`
	RoleID uint64 `json:"role_id"`
	Role   *Role  `json:"role,omitempty"`
}
```

`otmc` struct tag directives:

| Directive   | Effect                                              |
|-------------|-----------------------------------------------------|
| `primary`   | Marks the field as primary key                      |
| `auto`      | Auto-increment primary key                          |
| `required`  | Rejects empty values (422)                          |
| `unique`    | App-level uniqueness check (409 on conflict)        |
| `index`     | Adds a database index                               |
| `hidden`    | Never exposed in API responses                      |
| `readonly`  | Accepted on read, ignored on create/update          |
| `size=N`    | Max length for string columns                       |
| `-` / `ignore` | Excludes the field entirely                      |

### 2. Start the app

```go
app := otmc.New(otmc.Config{
	Name: "My Application",
	Port: 6000,
	APIPrefix: "/api", // default "/api"
	Debug: false,
	Database: otmc.DatabaseConfig{
		Driver: "sqlite", // "sqlite" | "postgres" | "mysql"
		DSN:    "./data/app.db",
	},
})

app.Register(&models.Role{}, &models.User{})
app.Run() // blocks until SIGINT/SIGTERM
```

Base URL: `http://localhost:6000/api`

---

## 📡 Auto-generated endpoints

For every registered model `X` with route `xs`, OTMC generates:

| Method   | Path              | Description                     |
|----------|-------------------|---------------------------------|
| `GET`    | `/api/xs`         | List (paginated, filter, sort)  |
| `GET`    | `/api/xs/search`  | Same as list (search alias)     |
| `GET`    | `/api/xs/:id`     | Get one record                  |
| `POST`   | `/api/xs`         | Create a record                 |
| `PUT`    | `/api/xs/:id`     | Full update (replace)           |
| `PATCH`  | `/api/xs/:id`     | Partial update                  |
| `DELETE` | `/api/xs/:id`     | Delete one record               |
| `DELETE` | `/api/xs?ids=…`   | Delete many records             |

### Response envelope

Success responses use a consistent envelope:

```json
{
  "success": true,
  "data": { ... },
  "message": "optional message"
}
```

Error responses:

```json
{
  "success": false,
  "error": {
    "code": 404,
    "key": "NOT_FOUND",
    "type": "Not Found",
    "summary": "record not found",
    "detail": "...",
    "timestamp": "2026-11-23T10:00:00+07:00"
  }
}
```

---

## 🧪 Examples — `Role` (`/api/roles`)

Model:

```go
type Role struct {
	ID          uint64 `json:"id" otmc:"primary;auto"`
	Name        string `json:"name" otmc:"required;unique"`
	Description string `json:"description"`
}
```

### 1. Create — `POST /api/roles`

Request:

```http
POST /api/roles
Content-Type: application/json

{
  "name": "admin",
  "description": "Administrator role"
}
```

Response `201 Created`:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "admin",
    "description": "Administrator role"
  }
}
```

### 2. List — `GET /api/roles`

Query params: `page`, `size` (max 100, default 20), `keyword`/`q`
(LIKE on all string fields), `sort` + `order=asc|desc`, and any
field name as an exact-match filter (`?name=admin`, `?role_id=1`).

```http
GET /api/roles?page=1&size=10&keyword=admin&sort=name&order=asc
```

Response `200 OK`:

```json
{
  "success": true,
  "data": {
    "items": [
      { "id": 1, "name": "admin", "description": "Administrator role" }
    ],
    "total": 1,
    "page": 1,
    "size": 10,
    "total_pages": 1
  }
}
```

Filter example: `GET /api/roles?name=admin` — any non-reserved query key
is treated as an exact-match field filter.

### 3. Get — `GET /api/roles/:id`

```http
GET /api/roles/1
```

Response `200 OK`:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "admin",
    "description": "Administrator role"
  }
}
```

Not found → `404`:

```json
{
  "success": false,
  "error": {
    "code": 404,
    "key": "NOT_FOUND",
    "type": "Not Found",
    "summary": "record not found",
    "detail": "record not found",
    "timestamp": "2026-11-23T10:00:00+07:00"
  }
}
```

### 4. Update (full replace) — `PUT /api/roles/:id`

```http
PUT /api/roles/1
Content-Type: application/json

{
  "name": "superadmin",
  "description": "Super administrator role"
}
```

Response `200 OK` — returns the updated record:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "superadmin",
    "description": "Super administrator role"
  }
}
```

### 5. Patch (partial update) — `PATCH /api/roles/:id`

Only fields present in the body are updated:

```http
PATCH /api/roles/1
Content-Type: application/json

{
  "description": "Patched description"
}
```

Response `200 OK`:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "superadmin",
    "description": "Patched description"
  }
}
```

### 6. Delete — `DELETE /api/roles/:id`

```http
DELETE /api/roles/1
```

Response `200 OK`:

```json
{
  "success": true,
  "message": "deleted"
}
```

### 7. Delete many — `DELETE /api/roles?ids=1,2,3`

```http
DELETE /api/roles?ids=1,2
```

Response `200 OK`:

```json
{
  "success": true,
  "data": { "deleted": 2 }
}
```

`ids` is required — missing it returns `400 Bad Request`.

---

## ⚠️ Error responses

| Status | `key`                | When                                        |
|--------|----------------------|---------------------------------------------|
| `400`  | `BAD_REQUEST`        | Invalid id, missing `ids`, bad body         |
| `404`  | `NOT_FOUND`          | Record does not exist                       |
| `409`  | `CONFLICT`           | `unique` field already taken                |
| `422`  | `UNPROCESSABLE_ENTITY` | `required` field missing/empty            |
| `500`  | `INTERNAL_SERVER_ERROR` | Unexpected server error                  |

Examples:

Duplicate `name` (unique) → `409 Conflict`:

```http
POST /api/roles
Content-Type: application/json

{ "name": "admin", "description": "Duplicate" }
```

```json
{
  "success": false,
  "error": {
    "code": 409,
    "key": "CONFLICT",
    "type": "Conflict",
    "summary": "conflict",
    "detail": "field \"name\" must be unique",
    "timestamp": "2026-11-23T10:00:00+07:00"
  }
}
```

Missing `name` (required) → `422 Unprocessable Entity`:

```http
POST /api/roles
Content-Type: application/json

{ "description": "Missing required name" }
```

```json
{
  "success": false,
  "error": {
    "code": 422,
    "key": "UNPROCESSABLE_ENTITY",
    "type": "Unprocessable Entity",
    "summary": "validation failed",
    "detail": "field \"name\" is required",
    "timestamp": "2026-11-23T10:00:00+07:00"
  }
}
```

Malformed JSON body → `500` (decoded as invalid request body) — always send
`Content-Type: application/json` with a valid JSON object.

---

## 🔧 Other endpoints

- `GET /healthz` — liveness probe:

```json
{ "status": "ok", "app": "My Application" }
```

---

## 🧪 Testing with REST Client (`.http` files)

The repo ships runnable collections in `tests/rest/`:

- `tests/rest/Roles.http` — full CRUD flow for `/api/roles`
- `tests/rest/Users.http` — user payloads incl. JSON/array fields

Run the example app first:

```powershell
go run ./examples/otmc
```

Then execute requests from your IDE's REST Client (port `6000`).
