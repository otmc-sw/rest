/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/otmc-sw/logger"
)

// StaticRegister serves the built frontend (SPA) and its static assets.
// It runs after API routes and custom handlers so those always win; the SPA
// fallback only catches requests nothing else matched.
//
// Directory resolution order:
//  1. staticDir, when set
//  2. ./frontend/dist           when production is true
//  3. ../frontend/dist          otherwise (dev monorepo layout)
//
// When no directory exists the registration is skipped, so API-only apps are
// unaffected.
func StaticRegister(app *fiber.App, staticDir string, production bool) {
	dist := staticDir
	if dist == "" {
		dist = filepath.Join("frontend", "dist")
		if !production {
			dist = filepath.Join("..", "frontend", "dist")
		}
	}

	if st, err := os.Stat(dist); err != nil || !st.IsDir() {
		if staticDir != "" || production {
			logger.Warn("⚠️  Static directory not found, skipping frontend serving: %s", dist)
		} else {
			logger.Debug("🐛 No frontend dist found at %s, static serving disabled.", dist)
		}
		return
	}

	app.Static("/assets", filepath.Join(dist, "assets"))
	app.Get("/logo.svg", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(dist, "logo.svg"))
	})
	app.Get("/app.ico", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(dist, "app.ico"))
	})
	app.Use(func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(dist, "index.html"))
	})
	logger.Info("🖼  Serving frontend static files from %s", dist)
}
