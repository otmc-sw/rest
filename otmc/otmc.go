/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
// Package otmc is a struct-driven application framework.
//
// Developers define Go structs and an otmc.Config; OTMC builds the
// database schema, repositories and REST endpoints automatically.
//
//	app := otmc.New(otmc.Config{Name: "My App", Port: 5007, Database: ...})
//	app.Register(&models.Role{}, &models.User{})
//	app.Run()
package otmc

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/otmc-sw/rest/otmc/httpapi"
	"github.com/otmc-sw/rest/otmc/schema"
	"github.com/otmc-sw/rest/otmc/store"
)

// Config configures the application.
type Config struct {
	Name      string
	Port      int    // default 5007
	APIPrefix string // default "/api"
	Debug     bool

	Database DatabaseConfig
}

// DatabaseConfig describes the database connection.
type DatabaseConfig struct {
	Driver string // "sqlite" (default) | "postgres" | "mysql"
	DSN    string // e.g. "./data/app.db", or postgres/mysql DSN
}

// App is a running OTMC application.
type App struct {
	cfg    Config
	models []any
	fiber  *fiber.App
	store  *store.Manager
	schema []*schema.Model
}

// New creates an application with the given configuration.
func New(cfg Config) *App {
	if cfg.Port <= 0 {
		cfg.Port = 5007
	}
	if cfg.APIPrefix == "" {
		cfg.APIPrefix = "/api"
	}
	if cfg.Name == "" {
		cfg.Name = "OTMC Application"
	}
	return &App{cfg: cfg}
}

// Register adds models (pointers to structs) whose tables, repositories and
// REST endpoints will be generated automatically.
func (a *App) Register(models ...any) *App {
	a.models = append(a.models, models...)
	return a
}

// Fiber returns the underlying Fiber app (available after Run starts building,
// mainly for tests and advanced customization).
func (a *App) Fiber() *fiber.App { return a.fiber }

// Store returns the database manager (available after Run starts).
func (a *App) Store() *store.Manager { return a.store }

// Run builds the schema, migrates the database, registers all routes and
// serves until SIGINT/SIGTERM. It blocks.
func (a *App) Run() error {
	if err := a.build(); err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", a.cfg.Port)
	go func() {
		fmt.Printf("OTMC %s listening on http://localhost%s\n", a.cfg.Name, addr)
		if err := a.fiber.Listen(addr); err != nil {
			fmt.Fprintf(os.Stderr, "otmc: server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Println("otmc: shutting down...")
	return a.fiber.ShutdownWithContext(ctx)
}

// build performs schema parsing, DB migration and route registration.
// Exposed internally so tests can exercise it without blocking on Listen.
func (a *App) build() error {
	for _, m := range a.models {
		model, err := schema.Parse(m)
		if err != nil {
			return err
		}
		a.schema = append(a.schema, model)
	}

	mgr, err := store.Open(store.Config{
		Driver: a.cfg.Database.Driver,
		DSN:    a.cfg.Database.DSN,
	})
	if err != nil {
		return err
	}
	a.store = mgr
	if err := mgr.Register(a.schema...); err != nil {
		return err
	}

	a.fiber = fiber.New(fiber.Config{
		AppName:      a.cfg.Name,
		DisableStartupMessage: true,
	})
	a.fiber.Use(recover.New())
	a.fiber.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool { return true },
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
	}))

	a.fiber.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": a.cfg.Name})
	})

	httpapi.RegisterRoutes(a.fiber, a.cfg.APIPrefix, mgr.Repositories())
	return nil
}
