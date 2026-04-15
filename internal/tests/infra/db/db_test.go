package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	db2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestConnect_EmptyDatabaseURL(t *testing.T) {
	_, err := db2.Connect(context.Background(), "")
	require.Error(t, err)
}

func TestConnect_InvalidDatabaseURL(t *testing.T) {
	_, err := db2.Connect(context.Background(), "postgres://%")
	require.Error(t, err)
}

func TestInitAndCheckMigration_NilDB(t *testing.T) {
	err := db2.InitAndCheckMigration(context.Background(), nil)
	require.Error(t, err)
}

func TestInitAndCheckMigration_ApplyMigrationError(t *testing.T) {
	driverName := "migration_err_driver_" + time.Now().UTC().Format("150405.000000000")
	sql.Register(driverName, &errDriver{})

	sqlDB, err := sql.Open(driverName, "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn:             sqlDB,
		WithoutReturning: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		DisableAutomaticPing:   true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db2.InitAndCheckMigration(context.Background(), gdb)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to apply migration")
}

type errDriver struct{}

func (d *errDriver) Open(name string) (driver.Conn, error) { return &errConn{}, nil }

type errConn struct{}

func (c *errConn) Prepare(query string) (driver.Stmt, error) { return &errStmt{}, nil }
func (c *errConn) Close() error                              { return nil }
func (c *errConn) Begin() (driver.Tx, error)                 { return &errTx{}, nil }
func (c *errConn) Ping(ctx context.Context) error            { return nil }

func (c *errConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return &errTx{}, nil
}

func (c *errConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("db boom")
}

func (c *errConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("db boom")
}

type errStmt struct{}

func (s *errStmt) Close() error  { return nil }
func (s *errStmt) NumInput() int { return -1 }
func (s *errStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("db boom")
}
func (s *errStmt) Query(args []driver.Value) (driver.Rows, error) { return nil, errors.New("db boom") }

type errTx struct{}

func (t *errTx) Commit() error {
	return nil
}
func (t *errTx) Rollback() error {
	return nil
}

var _ gorm.ConnPool = (*sql.DB)(nil)
