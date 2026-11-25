/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"

	"github.com/glebarez/sqlite"

	"github.com/otmc-sw/rest/app/schema"
)

// Config describes how to connect to the database.
type Config struct {
	Driver string // "sqlite" | "postgres" | "mysql"
	DSN    string
}

// Manager owns the GORM database handle and all registered repositories.
type Manager struct {
	DB      *gorm.DB
	repos   []*Repository
	byModel map[reflect.Type]*Repository
	namer   *otmcNamer
}

// otmcNamer maps registered model type names to their OTMC table names so that
// AutoMigrate creates tables matching schema.Model.Table (e.g. "user", not
// GORM's default plural "users"). Unknown names fall back to GORM defaults.
type otmcNamer struct {
	gormschema.NamingStrategy
	tables map[string]string // Go type name -> table name
}

func (n *otmcNamer) TableName(str string) string {
	if t, ok := n.tables[str]; ok {
		return t
	}
	return n.NamingStrategy.TableName(str)
}

// Open creates a GORM connection for the given driver.
func Open(cfg Config) (*Manager, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "", "sqlite":
		dsn := cfg.DSN
		if dsn == "" {
			dsn = "./data/app.db"
		}
		if err := ensureSQLiteDir(dsn); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(dsn)
	case "postgres", "postgresql":
		dialector = postgres.Open(cfg.DSN)
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("otmc: unsupported database driver %q", cfg.Driver)
	}
	namer := &otmcNamer{
		NamingStrategy: gormschema.NamingStrategy{IdentifierMaxLength: 64},
		tables:         map[string]string{},
	}
	db, err := gorm.Open(dialector, &gorm.Config{NamingStrategy: namer})
	if err != nil {
		return nil, fmt.Errorf("otmc: open database: %w", err)
	}
	return &Manager{DB: db, byModel: map[reflect.Type]*Repository{}, namer: namer}, nil
}

// ensureSQLiteDir creates the parent directory for file-based sqlite DSNs.
// In-memory DSNs (":memory:", "file::memory:...") are skipped.
func ensureSQLiteDir(dsn string) error {
	path := strings.TrimPrefix(dsn, "file:")
	if i := strings.IndexAny(path, "?"); i >= 0 {
		path = path[:i]
	}
	if path == "" || path == ":memory:" || strings.HasPrefix(path, ":memory:") {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("otmc: create database directory %q: %w", dir, err)
	}
	return nil
}

// Register parses schema metadata, runs AutoMigrate and creates repositories.
// Table names are declared for ALL models before any AutoMigrate runs, so that
// schemas pulled in early via relationships resolve to the correct table.
func (m *Manager) Register(models ...*schema.Model) error {
	// Phase 1: declare all table names and create repositories.
	for _, model := range models {
		repo := NewRepository(m.DB, model)
		m.namer.tables[model.Name] = model.Table
		m.repos = append(m.repos, repo)
		m.byModel[model.GoType] = repo
	}
	// Phase 2: migrate.
	for _, model := range models {
		if err := m.DB.AutoMigrate(reflect.New(model.GoType).Interface()); err != nil {
			return fmt.Errorf("otmc: automigrate %s: %w", model.Name, err)
		}
	}
	return nil
}

// Repositories returns all registered repositories.
func (m *Manager) Repositories() []*Repository {
	return m.repos
}

// Repository resolves the repository for a registered model Go type.
func (m *Manager) Repository(t reflect.Type) *Repository {
	return m.byModel[t]
}

// GormDB returns the underlying GORM database instance.
func (m *Manager) GormDB() *gorm.DB {
	return m.DB
}

// Close closes the underlying database connection.
func (m *Manager) Close() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
