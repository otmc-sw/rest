

## 📦 Installation

```bash
go get github.com/otmc-sw/rest@latest
```


## 🚀 Usage

tạo context file cho Fiber
```go
package handlers_v2

import (
	"bytes"
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	sqlc "otmc/app/db/sqlc"
)

type FiberContext struct {
	*fiber.Ctx
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

tạo handlers ví dụ cho users
```go
package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	rest "github.com/otmc-sw/rest"
	db "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

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


khởi tạo Fiber app ví dụ:
```go
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/otmc-sw/logger"

	db "github.com/otmc-sw/rest/examples/fiber/db"
	handlers "otmc/app/handlers"
)

var (
	FLAG_PROD  = false
	FLAG_DEBUG = false
	PORT       = 3000
	DIR_RUN, _ = os.Getwd()

	app      *fiber.App
	database *db.DataBase
)

func PrintBanner() {
	fmt.Println("========================================")
	fmt.Println("  OTMC REST Example Server")
	fmt.Println("========================================")
}

func CreateDataDirectories() {
	dirs := []string{
		filepath.Join(DIR_RUN, "data", "db"),
		filepath.Join(DIR_RUN, "data", "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Crit("❌ Create directory failed: %s %v", dir, err)
			os.Exit(1)
		}
		logger.Info("📁 Directory ready: %s", dir)
	}
}

func Initializer() {
	var err error

	PrintBanner()

	logFile := filepath.Join(DIR_RUN, "data", "logs", "app.log")
	logger.Configure(
		logger.WithFile(logFile),
	)

	if FLAG_DEBUG {
		logger.SetLevel(logger.DebugLevel)
	}

	logger.Info("✅ Logger initialized successfully")

	logger.Info("✨ Initializing application...")
	CreateDataDirectories()

	logger.Info("📚 Initializing database connection ...")
	database, err = db.New()
	if err != nil {
		logger.Error("❌ Failed to connect to database: %v", err)
		os.Exit(1)
	}

	logger.Info("✅ Database connection established.")

	logger.Info("📦 Setting up handlers...")
	handlers.New(database.Queries)
	logger.Info("✅ Handlers configured.")

	app = fiber.New(fiber.Config{
		AppName:     "OTMC REST Example Server",
		IdleTimeout: 30 * time.Second,
	})

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool { return true },
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Request(c.Method(), c.Path(), c.Response().StatusCode(), time.Since(start), c.IP())
		return err
	})

	logger.Info("📁 Run directory: %s", DIR_RUN)
}

func Runner() {
	logger.Info("🌿 Running application ...")
	api := app.Group("/api")

	logger.Info("🌐 Registering APIs ...")

	api.Post("/users", handlers.CreateUser)
	api.Get("/users", handlers.GetAllUsers)
	api.Get("/users/:id", handlers.GetUser)
	api.Patch("/users/:id", handlers.UpdateUser)
	api.Delete("/users/:id", handlers.DeleteUser)

	api.Get("/test", handlers.TestResponse)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", PORT)
		logger.Info("🚀 Server starting at http://localhost:%d", PORT)
		if err := app.Listen(addr); err != nil {
			logger.Error("❌ Server failed: %v", err)
		}
	}()

	<-quit
	logger.Info("🛑 Shutdown signal received.")
}

func Finisher() {
	if database != nil {
		database.Close()
	}
}

func ParserArguments() {
	flag.BoolVar(&FLAG_PROD, "P", false, "Run in production mode")
	flag.BoolVar(&FLAG_DEBUG, "d", false, "Enable debug mode")
	flag.IntVar(&PORT, "p", PORT, "Port")
	flag.Parse()
}

func main() {
	ParserArguments()
	Initializer()
	Runner()
	Finisher()
}
```
