# 🎨 Gin Integration

## 📦 Installation

```bash
go get github.com/otmc-sw/rest@latest
```

## 📖 Usage

### 1. Create Gin Context

Create a context adapter for Gin:

```go
package handlers_v2

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	rest "github.com/otmc-sw/rest"
	sqlc "otmc/app/db/sqlc"
)

type GinContext struct {
	*gin.Context
}

var database *sqlc.Queries

func New(db *sqlc.Queries) {
	database = db
}

func (c GinContext) Context() context.Context { return c.Request.Context() }
func (c GinContext) Param(key string) string  { return c.Param(key) }
func (c GinContext) Query(key string) string  { return c.Query(key) }
func (c GinContext) QueryAll(key string) []string {
	v := c.Query(key)
	if v == "" {
		return nil
	}
	return []string{v}
}
func (c GinContext) Header(key string) string { return c.GetHeader(key) }
func (c GinContext) Cookie(key string) string  { return c.Cookie(key) }
func (c GinContext) Body() io.Reader          { return c.Request.Body }
func (c GinContext) Bind(v interface{}) error { return c.ShouldBindJSON(v) }
func (c GinContext) JSON(code int, body interface{}) error {
	c.JSON(code, body)
	return nil
}
func (c GinContext) Status(code int)             { c.Status(code) }
func (c GinContext) SetHeader(key, value string) { c.Header(key, value) }
func (c GinContext) Method() string              { return c.Request.Method }
func (c GinContext) Path() string                { return c.Request.URL.Path }
func (c GinContext) String() (string, error) {
	body, _ := io.ReadAll(c.Request.Body)
	return string(body), nil
}
func (c GinContext) Bytes() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}
```

### 2. Create Handlers

Example handlers for users:

```go
package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	rest "github.com/otmc-sw/rest"
	db "github.com/otmc-sw/rest/examples/gin/db/sqlc"
)

type UserRequest struct {
	Username *string `json:"username"`
	FullName *string `json:"full_name,omitempty"`
	Email    *string `json:"email"`
	Content  *string `json:"content,omitempty"`
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

func CreateUser(c *gin.Context) error {
	return rest.
		Create[UserRequest, db.CreateUserParams, db.User, UserResponse](GinContext{Context: c}).
		Bind().
		Validate(ValidateUser).
		Exec(func(ctx rest.Context, req UserRequest, params db.CreateUserParams, id any) (any, error) {
			return database.CreateUser(ctx.Context(), params)
		}).
		Respond()
}

func GetUser(c *gin.Context) error {
	return rest.
		Get[struct{}, struct{}, db.User, UserResponse](GinContext{Context: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetUser(ctx.Context(), id.(int64))
		}).
		Respond()
}

func GetAllUsers(c *gin.Context) error {
	return rest.
		Get[struct{}, struct{}, []db.User, []UserResponse](GinContext{Context: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return database.GetAllUsers(ctx.Context())
		}).
		Respond()
}

func UpdateUser(c *gin.Context) error {
	return rest.
		Update[UserRequest, db.UpdateUserParams, db.User, UserResponse](GinContext{Context: c}).
		Bind().
		Exec(func(ctx rest.Context, req UserRequest, params db.UpdateUserParams, id any) (any, error) {
			return database.UpdateUser(ctx.Context(), params)
		}).
		Respond()
}

func DeleteUser(c *gin.Context) error {
	return rest.
		Delete[struct{}, struct{}, struct{}, UserResponse](GinContext{Context: c}).
		Exec(func(ctx rest.Context, req struct{}, params struct{}, id any) (any, error) {
			return nil, database.DeleteUser(ctx.Context(), id.(int64))
		}).
		Respond()
}
```

### 3. Initialize Gin App

Example application setup:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/otmc-sw/logger"

	db "github.com/otmc-sw/rest/examples/gin/db"
	handlers "otmc/app/handlers"
)

//go:embed configs/banner.txt
var APP_BANNER string

var (
	FLAG_PROD  = false
	FLAG_DEBUG = false
	PORT       = 3000
	DIR_RUN, _ = os.Getwd()

	app      *gin.Engine
	database *db.DataBase
)

func PrintBanner() {
	fmt.Print(APP_BANNER)
}

func ParserArguments() {
	flag.BoolVar(&FLAG_PROD, "P", false, "Run in production mode")
	flag.BoolVar(&FLAG_DEBUG, "d", false, "Enable debug mode")
	flag.IntVar(&PORT, "p", PORT, "Port")
	flag.Parse()
}

func SetupLogger() {
	logFile := filepath.Join(DIR_RUN, "data", "logs", "app.log")
	logger.Configure(
		logger.WithFile(logFile),
	)

	if FLAG_DEBUG {
		logger.SetLevel(logger.DebugLevel)
	}

	logger.Info("✅ Logger initialized successfully.")
}

func InitializeDatabase() {
	var err error

	logger.Info("✨ Initializing database connection...")
	database, err = db.New()
	if err != nil {
		logger.Error("❌ Failed to connect to database: %v", err)
		os.Exit(1)
	}

	logger.Info("✅ Database connection established.")
}

func SetupGin() {
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()

	app.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func SetupHandlers() {
	logger.Info("📚 Setting up handlers...")
	handlers.New(database.Queries)
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
	PrintBanner()
	ParserArguments()
	SetupLogger()
	CreateDataDirectories()
	InitializeDatabase()
	SetupGin()
	SetupHandlers()
}

func Runner() {
	logger.Info("🌿 Running application ...")
	api := app.Group("/api")

	logger.Info("📚 Registering APIs...")
	api.POST("/users", handlers.CreateUser)
	api.GET("/users", handlers.GetAllUsers)
	api.GET("/users/:id", handlers.GetUser)
	api.PATCH("/users/:id", handlers.UpdateUser)
	api.DELETE("/users/:id", handlers.DeleteUser)

	api.GET("/test", handlers.TestResponse)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", PORT)
		logger.Info("🚀 Server starting at http://localhost:%d", PORT)
		if err := app.Run(addr); err != nil {
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

func main() {
	Initializer()
	Runner()
	Finisher()
}

```