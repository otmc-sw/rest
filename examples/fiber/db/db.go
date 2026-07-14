/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/otmc-sw/logger"
	db "github.com/otmc-sw/rest/examples/fiber/db/sqlc"

	_ "modernc.org/sqlite"
)

type DataBase struct {
	*sql.DB
	*db.Queries
}

func New() (*DataBase, error) {
	dataDir := filepath.Join(mustGetwd(), "data", "db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, logErr("create data directory", err)
	}

	dbPath := filepath.Join(dataDir, "main.db")

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, logErr("open database", err)
	}

	if err := enableForeignKeys(sqlDB); err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	database := &DataBase{
		DB:      sqlDB,
		Queries: db.New(sqlDB),
	}

	if err := database.MigrateSchemas(); err != nil {
		return nil, err
	}

	return database, nil
}

func enableForeignKeys(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return logErr("enable foreign keys", err)
	}
	return nil
}

func (db *DataBase) MigrateSchemas() error {
	if err := db.execSQLDir(MigrationFS, "migration/pre"); err != nil {
		return err
	}

	if err := db.execSQLDir(SchemaFS, "schemas"); err != nil {
		return err
	}

	logger.Info("✅ Database creation completed.")

	if err := db.execSQLDir(MigrationFS, "migration/post"); err != nil {
		return err
	}

	if err := db.execSQLDir(MigrationFS, "migration/samples"); err != nil {
		return err
	}

	return nil
}

func (db *DataBase) execSQLDir(fsys fs.FS, root string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".sql" {
			return err
		}

		return db.execSQLFile(fsys, path)
	})
}

func (db *DataBase) execSQLFile(fsys fs.FS, path string) error {
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read SQL file %s: %w", path, err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("execute SQL file %s: %w", path, err)
	}

	logger.Info("📝 Executed SQL file: %s", path)
	return nil
}

func (db *DataBase) Close() error {
	logger.Info("✅ Database connection closed.")
	return db.DB.Close()
}

func OpenDatabase(dbPath string) (*DataBase, error) {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := enableForeignKeys(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	database := &DataBase{
		DB:      sqlDB,
		Queries: db.New(sqlDB),
	}

	return database, nil
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func logErr(action string, err error) error {
	logger.Error("❌ Failed to %s: %v", action, err)
	return fmt.Errorf("failed to %s: %w", action, err)
}
