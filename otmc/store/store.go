/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package store

import (
	"fmt"
	"reflect"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"

	"github.com/otmc-sw/rest/otmc/schema"
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
}

// Open creates a GORM connection for the given driver.
func Open(cfg Config) (*Manager, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "", "sqlite":
		dsn := cfg.DSN
		if dsn == "" {
			dsn = "otmc.db"
		}
		dialector = sqlite.Open(dsn)
	case "postgres", "postgresql":
		dialector = postgres.Open(cfg.DSN)
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("otmc: unsupported database driver %q", cfg.Driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("otmc: open database: %w", err)
	}
	return &Manager{DB: db, byModel: map[reflect.Type]*Repository{}}, nil
}

// Register parses schema metadata, runs AutoMigrate and creates repositories.
func (m *Manager) Register(models ...*schema.Model) error {
	for _, model := range models {
		repo := NewRepository(m.DB, model)
		if err := m.DB.AutoMigrate(reflect.New(model.GoType).Interface()); err != nil {
			return fmt.Errorf("otmc: automigrate %s: %w", model.Name, err)
		}
		m.repos = append(m.repos, repo)
		m.byModel[model.GoType] = repo
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

// Close closes the underlying database connection.
func (m *Manager) Close() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
