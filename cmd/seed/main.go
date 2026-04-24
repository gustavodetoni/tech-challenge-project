package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/soat-architecture/tech-challenge-project/internal/infra/db"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/db/seed"
)

var (
	connectDB                     = db.Connect
	initAndCheck                  = db.InitAndCheckMigration
	runSeed                       = seed.Run
	nowUTC       func() time.Time = func() time.Time { return time.Now().UTC() }
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	opts := seed.Options{
		Force:      envBool("SEED_FORCE", false),
		RandomSeed: envInt64("SEED_RANDOM_SEED", 123),

		Users:    envInt("SEED_USERS", 8),
		Clients:  envInt("SEED_CLIENTS", 12),
		Services: envInt("SEED_SERVICES", 8),
		Parts:    envInt("SEED_PARTS", 24),
		Orders:   envInt("SEED_ORDERS", 12),
		Now:      nowUTC(),
	}

	gormDB, err := connectDB(ctx, databaseURL)
	if err != nil {
		return err
	}
	log.Println("Connected to database")

	if err := initAndCheck(ctx, gormDB); err != nil {
		return err
	}

	if err := runSeed(ctx, gormDB, opts); err != nil {
		return err
	}

	return nil
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}
