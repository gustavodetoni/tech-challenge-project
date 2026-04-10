package db

import (
	"context"
	"fmt"

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

