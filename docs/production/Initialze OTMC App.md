# 🚀 OTMC — Struct-Driven REST Framework
Migration current BE structure to new structure
Application cần build theo cấu trúc thư mục:

```
app/
├── configs/     
├── core/        // core business logic
├── models/
│   ├── user.go
│   └── role.go
├── db/
│   ├── user.sql
│   └── role.sql
└── handlers/
    └── user.go
main.go
go.mod
```

OTMC builds database schema, repositories and REST endpoints automatically from
Go structs. Developers only define models and a `otmc.Config` — everything else
(table, migration, validation, routes) is generated.

## 📦 Installation

```bash
go get github.com/otmc-sw/rest@latest
```

## 📖 Usage

### 1. Define models

/models/user.go
/models/role.go

```go
package models

// Role is automatically migrated to a "role" table and exposed at /api/role.
type Role struct {
	ID          uint64     `json:"id" otmc:"primary;auto"`
	Name        string     `json:"name" otmc:"required;unique"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
```

```go
package models

// User is automatically migrated to a "user" table and exposed at /api/user.
// RoleID + Role give GORM a belongs-to relation to Role.
type User struct {
	ID        uint64     `json:"id" otmc:"primary;auto"`
	Username  string     `json:"username" otmc:"required;unique"`
	Email     string     `json:"email" otmc:"required;unique"`
	Password  string     `json:"password,omitempty" otmc:"hidden"`
	FullName  string     `json:"full_name"`
	AvatarURL string     `json:"avatar_url"`
	Status    string     `json:"status"`
	RoleID    *uint64    `json:"role_id"`
	Role      *Role      `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
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
| `hash=bcrypt` | Hashes plaintext with bcrypt before save; already-hashed values are left untouched |
| `-` / `ignore` | Excludes the field entirely                      |

### 2. Start the app
main.go

```go
package main

package main

import (
	_ "embed"

	"github.com/otmc-sw/rest/app"
	"otmc/app/handlers"
	"otmc/models"
)

func main() {
	// Initialize the application
	myapp := app.New(app.Config{
		Name:        "My Demo Application",
		Port:        6000,
		Banner:      "configs/banner.txt",
		InitialData: "db/*.sql",
	})

	// Register models and default handlers
	myapp.Register(
		&models.Role{},
		&models.User{},
	)

	// Register custom handlers
	myapp.RegisterHandlers(handlers.RegisterUserHandlers)

	// Run the application
	myapp.Runner()
}

```

Base URL: `http://localhost:6000/api`

---

## 3. Define initial data

Create SQL files in the `db/` directory to seed your database on startup. For example, `db/roles.sql`:

db/roles.sql
```sql
-- Initial data for roles

-- Clear existing data first
DELETE FROM role;

-- Insert initial roles
INSERT OR IGNORE INTO role (id, name, description, status, created_at, updated_at) VALUES
(1, 'admin', 'Administrator with full access', 'active', datetime('now'), datetime('now')),
(2, 'user', 'Regular user with limited access', 'active', datetime('now'), datetime('now')),
(3, 'guest', 'Guest user with read-only access', 'active', datetime('now'), datetime('now'));

```

db/users.sql
```sql
-- Initial data for users

-- Clear existing data first
DELETE FROM user;

-- Insert initial users
INSERT OR IGNORE INTO user (id, username, email, password, role_id, status, created_at, updated_at) VALUES
(1, 'admin', 'admin@example.com', 'hashed_password', 1, 'active', datetime('now'), datetime('now')),
(2, 'user', 'user@example.com', 'hashed_password', 2, 'active', datetime('now'), datetime('now')),
(3, 'guest', 'guest@example.com', 'hashed_password', 3, 'active', datetime('now'), datetime('now'));
```

4. Create custom handlers if needed
handlers/user.go
```go
/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"

	"github.com/otmc-sw/rest/app"
	"github.com/otmc-sw/rest/app/handlers"
	"github.com/otmc-sw/rest/app/store"
	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/response"
)

// RegisterUserHandlers registers custom user-related handlers.
func RegisterUserHandlers(app *fiber.App, prefix string, mgr *store.Manager) {
	userRepo := handlers.GetRepository(mgr, "User")
	if userRepo == nil {
		return
	}

	base := prefix + "/user"

	// Custom user routes
	app.Post(base+"/login", Login(userRepo))
	app.Post(base+"/logout", Logout)
	app.Post(base+"/:id/avatar", UploadAvatar(userRepo))
}

// Login handles user authentication.
func Login(repo *store.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := handlers.FiberContext{Ctx: c}

		// Parse request body
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return errors.New().BadRequest().Summary("invalid request body").Send(ctx)
		}

		// Simple authentication logic (in production, use proper password hashing)
		// This is a simplified example
		query := store.Query{
			Filters: map[string]string{
				"username": req.Username,
			},
		}

		result, err := repo.List(c.Context(), query)
		if err != nil {
			return errors.New().InternalError().Summary("authentication failed").Send(ctx)
		}

		if len(result.Items) == 0 {
			return errors.New().Unauthorized().Summary("invalid credentials").Send(ctx)
		}

		// In production, verify password hash here
		// For this example, we'll just return success
		user := result.Items[0]

		// Extract ID from user struct
		userID := getUserID(user)

		return response.OK[map[string]any](ctx).Data(map[string]any{
			"user":  user,
			"token": "fake-jwt-token-" + fmt.Sprint(userID),
		}).Message("login successful").Send()
	}
}

// Logout handles user logout.
func Logout(c *fiber.Ctx) error {
	ctx := handlers.FiberContext{Ctx: c}
	return response.OK[string](ctx).Message("logout successful").Send()
}

// UploadAvatar handles avatar upload for a user.
func UploadAvatar(repo *store.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := handlers.FiberContext{Ctx: c}

		// Parse user ID
		id, err := repo.ParseID(c.Params("id"))
		if err != nil {
			return errors.New().BadRequest().Summary("invalid user id").Send(ctx)
		}

		// Get the file from form
		file, err := c.FormFile("avatar")
		if err != nil {
			return errors.New().BadRequest().Summary("avatar file is required").Send(ctx)
		}

		// In production, upload to cloud storage and get URL
		// For this example, we'll just simulate it
		avatarURL := fmt.Sprintf("/uploads/avatars/%s-%s", c.Params("id"), file.Filename)

		// Update user with avatar URL - need to pass JSON bytes
		updateData := map[string]any{
			"avatar_url": avatarURL,
		}
		body, err := json.Marshal(updateData)
		if err != nil {
			return errors.New().InternalError().Summary("failed to marshal update data").Send(ctx)
		}

		updatedUser, err := repo.Patch(c.Context(), id, body)
		if err != nil {
			return errors.New().InternalError().Summary("failed to update avatar").Send(ctx)
		}

		return response.OK[any](ctx).Data(updatedUser).Message("avatar uploaded successfully").Send()
	}
}

// getUserID extracts the ID field from a user struct using reflection
func getUserID(user any) any {
	v := reflect.ValueOf(user)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		idField := v.FieldByName("ID")
		if idField.IsValid() {
			return idField.Interface()
		}
	}
	return nil
}

// HandlerFunc is the type signature for custom handlers.
// This matches app.HandlerFunc for easy registration.
type HandlerFunc = app.HandlerFunc

```

### Note: Auto-generated endpoints

For every registered model `X` with route `x` (singular snake_case), OTMC generates:

| Method   | Path              | Description                     |
|----------|-------------------|---------------------------------|
| `GET`    | `/api/x`         | List (paginated, filter, sort)  |
| `GET`    | `/api/xs/search`  | Same as list (search alias)     |
| `GET`    | `/api/xs/:id`     | Get one record                  |
| `POST`   | `/api/x`         | Create a record                 |
| `PUT`    | `/api/xs/:id`     | Full update (replace)           |
| `PATCH`  | `/api/xs/:id`     | Partial update                  |
| `PATCH`  | `/api/xs?ids=…`   | Partial update many records     |
| `DELETE` | `/api/xs/:id`     | Delete one record               |
| `DELETE` | `/api/xs?ids=…`   | Delete many records             |


