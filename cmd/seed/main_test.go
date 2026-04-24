package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/soat-architecture/tech-challenge-project/internal/infra/db/seed"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEnvBool(t *testing.T) {
	t.Run("empty_uses_fallback", func(t *testing.T) {
		require.True(t, envBool("SEED_BOOL_X", true))
	})

	t.Run("valid_true", func(t *testing.T) {
		t.Setenv("SEED_BOOL_X", "true")
		require.True(t, envBool("SEED_BOOL_X", false))
	})

	t.Run("invalid_uses_fallback", func(t *testing.T) {
		t.Setenv("SEED_BOOL_X", "not-a-bool")
		require.False(t, envBool("SEED_BOOL_X", false))
		require.True(t, envBool("SEED_BOOL_X", true))
	})
}

func TestEnvInt(t *testing.T) {
	t.Run("empty_uses_fallback", func(t *testing.T) {
		require.Equal(t, 10, envInt("SEED_INT_X", 10))
	})

	t.Run("valid_int", func(t *testing.T) {
		t.Setenv("SEED_INT_X", "42")
		require.Equal(t, 42, envInt("SEED_INT_X", 10))
	})

	t.Run("invalid_uses_fallback", func(t *testing.T) {
		t.Setenv("SEED_INT_X", "nope")
		require.Equal(t, 10, envInt("SEED_INT_X", 10))
	})
}

func TestEnvInt64(t *testing.T) {
	t.Run("empty_uses_fallback", func(t *testing.T) {
		require.Equal(t, int64(123), envInt64("SEED_INT64_X", 123))
	})

	t.Run("valid_int64", func(t *testing.T) {
		t.Setenv("SEED_INT64_X", "922337203685477580")
		require.Equal(t, int64(922337203685477580), envInt64("SEED_INT64_X", 1))
	})

	t.Run("invalid_uses_fallback", func(t *testing.T) {
		t.Setenv("SEED_INT64_X", "nope")
		require.Equal(t, int64(1), envInt64("SEED_INT64_X", 1))
	})
}

func TestRun_MissingDatabaseURL(t *testing.T) {
	// Not parallel: touches injected package-level vars in other tests.
	require.ErrorContains(t, run(context.Background()), "DATABASE_URL is required")
}

func TestRun_SuccessAndOptions(t *testing.T) {
	// Not parallel: overrides package-level vars.
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("SEED_FORCE", "true")
	t.Setenv("SEED_RANDOM_SEED", "777")
	t.Setenv("SEED_USERS", "2")
	t.Setenv("SEED_CLIENTS", "3")
	t.Setenv("SEED_SERVICES", "4")
	t.Setenv("SEED_PARTS", "5")
	t.Setenv("SEED_ORDERS", "6")

	fixedNow := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)

	origConnect := connectDB
	origInit := initAndCheck
	origRunSeed := runSeed
	origNow := nowUTC
	t.Cleanup(func() {
		connectDB = origConnect
		initAndCheck = origInit
		runSeed = origRunSeed
		nowUTC = origNow
	})

	var gotURL string
	var gotOpts seed.Options
	connectDB = func(ctx context.Context, databaseURL string) (*gorm.DB, error) {
		gotURL = databaseURL
		return &gorm.DB{}, nil
	}
	initAndCheck = func(ctx context.Context, gdb *gorm.DB) error { return nil }
	runSeed = func(ctx context.Context, gdb *gorm.DB, opts seed.Options) error {
		gotOpts = opts
		return nil
	}
	nowUTC = func() time.Time { return fixedNow }

	require.NoError(t, run(context.Background()))
	require.Equal(t, "postgres://example", gotURL)
	require.Equal(t, seed.Options{
		Force:      true,
		RandomSeed: 777,
		Users:      2,
		Clients:    3,
		Services:   4,
		Parts:      5,
		Orders:     6,
		Now:        fixedNow,
	}, gotOpts)
}

func TestRun_PropagatesErrors(t *testing.T) {
	// Not parallel: overrides package-level vars.
	t.Setenv("DATABASE_URL", "postgres://example")

	origConnect := connectDB
	origInit := initAndCheck
	origRunSeed := runSeed
	t.Cleanup(func() {
		connectDB = origConnect
		initAndCheck = origInit
		runSeed = origRunSeed
	})

	connectDB = func(ctx context.Context, databaseURL string) (*gorm.DB, error) {
		return nil, errors.New("connect-failed")
	}
	require.ErrorContains(t, run(context.Background()), "connect-failed")

	connectDB = func(ctx context.Context, databaseURL string) (*gorm.DB, error) { return &gorm.DB{}, nil }
	initAndCheck = func(ctx context.Context, gdb *gorm.DB) error { return errors.New("migration-failed") }
	require.ErrorContains(t, run(context.Background()), "migration-failed")

	initAndCheck = func(ctx context.Context, gdb *gorm.DB) error { return nil }
	runSeed = func(ctx context.Context, gdb *gorm.DB, opts seed.Options) error { return errors.New("seed-failed") }
	require.ErrorContains(t, run(context.Background()), "seed-failed")
}
