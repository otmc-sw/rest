/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
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
	handlers "github.com/otmc-sw/rest/examples/fiber/handlers"
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
		filepath.Join(DIR_RUN, "data", "files"),
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

	logger.Info("🌐 Registering APIs ...")

	app.Post("/users", handlers.CreateUser)
	app.Get("/users", handlers.GetAllUsers)
	app.Get("/users/:id", handlers.GetUser)
	app.Patch("/users/:id", handlers.UpdateUser)
	app.Delete("/users/:id", handlers.DeleteUser)

	app.Get("/test", handlers.TestResponse)
	app.Get("/set/fields/:id", handlers.SetFields)
	app.Get("/set/field/:id", handlers.SetField)

	app.Get("/download/file", handlers.DownloadFile)
	app.Post("/upload/file", handlers.UploadFile)

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
