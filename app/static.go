/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package app

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/otmc-sw/logger"
)

// registerStatic serves the built frontend (SPA) and its static assets.
// It runs after API routes and custom handlers so those always win; the SPA
// fallback only catches requests nothing else matched.
//
// Directory resolution order:
//  1. Config.StaticDir, when set
//  2. ./frontend/dist           when Config.Production is true
//  3. ../frontend/dist          otherwise (dev monorepo layout)
//
// When no directory exists the registration is skipped, so API-only apps are
// unaffected.
func (a *App) registerStatic() {
	dist := a.cfg.StaticDir
	if dist == "" {
		dist = filepath.Join("frontend", "dist")
		if !a.cfg.Production {
			dist = filepath.Join("..", "frontend", "dist")
		}
	}

	if st, err := os.Stat(dist); err != nil || !st.IsDir() {
		if a.cfg.StaticDir != "" || a.cfg.Production {
			logger.Warn("⚠️  Static directory not found, skipping frontend serving: %s", dist)
		} else {
			logger.Debug("🐛 No frontend dist found at %s, static serving disabled.", dist)
		}
		return
	}

	a.fiber.Static("/assets", filepath.Join(dist, "assets"))
	a.fiber.Get("/logo.svg", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(dist, "logo.svg"))
	})
	a.fiber.Get("/app.ico", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(dist, "app.ico"))
	})
	a.fiber.Use(func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(dist, "index.html"))
	})
	logger.Info("🖼  Serving frontend static files from %s", dist)
}
