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

/models/user.go
/models/role.go

```go
package models

// Role is automatically migrated to a "roles" table and exposed at /api/roles.
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

// User is automatically migrated to a "users" table and exposed at /api/users.
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
| `-` / `ignore` | Excludes the field entirely                      |

### 2. Start the app
main.go

```go
/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package main

import (
	_ "embed"

	"github.com/otmc-sw/rest/examples/otmc/models"
	"github.com/otmc-sw/rest/otmc"
)

func main() {
	app := otmc.New(otmc.Config{
		Name:        "My Application",
		Port:        6000,
		Banner:      "configs/banner.txt",
		InitialData: "db/*.sql",
	})

	app.Register(
		&models.Role{},
		&models.User{},
	)

	app.Runner()
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
DELETE FROM roles;


-- Insert initial roles
INSERT OR IGNORE INTO roles (id, name, description, status, created_at, updated_at) VALUES
(1, 'admin', 'Administrator with full access', 'active', datetime('now'), datetime('now')),
(2, 'user', 'Regular user with limited access', 'active', datetime('now'), datetime('now')),
(3, 'guest', 'Guest user with read-only access', 'active', datetime('now'), datetime('now'));

```

db/users.sql
```sql
-- Initial data for users

-- Clear existing data first
DELETE FROM users;

-- Insert initial users
INSERT OR IGNORE INTO users (id, username, email, password, role_id, status, created_at, updated_at) VALUES
(1, 'admin', 'admin@example.com', 'hashed_password', 1, 'active', datetime('now'), datetime('now')),
(2, 'user', 'user@example.com', 'hashed_password', 2, 'active', datetime('now'), datetime('now')),
(3, 'guest', 'guest@example.com', 'hashed_password', 3, 'active', datetime('now'), datetime('now'));
```

### Note: Auto-generated endpoints

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


