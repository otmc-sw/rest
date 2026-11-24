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
	"github.com/otmc-sw/logger"

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
	Banner    string // optional startup banner

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
func (a *App) Runner() error {
	if a.cfg.Banner != "" {
		fmt.Println(a.cfg.Banner)
	}
	if a.cfg.Debug {
		logger.SetLevel(logger.DebugLevel)
		logger.Debug("🐛 Debug mode enabled.")
	}
	logger.Info("✨ Initializing %s...", a.cfg.Name)
	if err := a.Initializer(); err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", a.cfg.Port)
	go func() {
		logger.Info("🚀 Server starting at http://localhost%s", addr)
		if err := a.fiber.Listen(addr); err != nil {
			logger.Error("❌ otmc: server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutdown signal received.")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("⏳ Shutting down server...")
	return a.fiber.ShutdownWithContext(ctx)
}

// build performs schema parsing, DB migration and route registration.
// Exposed internally so tests can exercise it without blocking on Listen.
func (a *App) Initializer() error {
	if len(a.models) > 0 {
		logger.Info("📝 Parsing %d model(s)...", len(a.models))
		for _, m := range a.models {
			model, err := schema.Parse(m)
			if err != nil {
				return err
			}
			a.schema = append(a.schema, model)
			logger.Info("✅ Model parsed: %s (table: %s)", model.Name, model.Table)
		}
	} else {
		logger.Info("📋 No models registered.")
	}

	driver := a.cfg.Database.Driver
	if driver == "" {
		driver = "sqlite"
	}
	logger.Info("💾 Connecting to %s database...", driver)
	mgr, err := store.Open(store.Config{
		Driver: a.cfg.Database.Driver,
		DSN:    a.cfg.Database.DSN,
	})
	if err != nil {
		return err
	}
	logger.Info("✅ Database connection established.")
	a.store = mgr

	for _, model := range a.schema {
		logger.Info("📚  Migrating table %s...", model.Table)
		if err := mgr.Register(model); err != nil {
			return err
		}
		logger.Info("✅ Table %s ready.", model.Table)
	}

	logger.Info("🧩 Initializing web server...")
	a.fiber = fiber.New(fiber.Config{
		AppName:               a.cfg.Name,
		DisableStartupMessage: true,
	})
	a.fiber.Use(recover.New())
	a.fiber.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Request(c.Method(), c.OriginalURL(), c.Response().StatusCode(), time.Since(start), c.IP())
		return err
	})
	a.fiber.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool { return true },
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
	}))

	a.fiber.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": a.cfg.Name})
	})

	logger.Info("🌐 Registering API routes under %s...", a.cfg.APIPrefix)
	httpapi.RegisterRoutes(a.fiber, a.cfg.APIPrefix, mgr.Repositories())
	logger.Info("✅ API routes registered.")
	return nil
}
