package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	return gormDB, nil
}

func InitAndCheckMigration(ctx context.Context, gormDB *gorm.DB) error {
	if gormDB == nil {
		return fmt.Errorf("gormDB is required")
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to resolve migrations directory")
	}
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "migrations")
	if !isDir(migrationsDir) {
		candidates := []string{
			filepath.Join("internal", "infra", "db", "migrations"),
			filepath.Join(string(filepath.Separator), "app", "internal", "infra", "db", "migrations"),
		}
		for _, candidate := range candidates {
			if isDir(candidate) {
				migrationsDir = candidate
				break
			}
		}
	}
	if !isDir(migrationsDir) {
		wd, _ := os.Getwd()
		return fmt.Errorf("migrations directory does not exist: %s (cwd=%s)", migrationsDir, wd)
	}

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("failed to apply migration: %v", err)
	}

	return nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
