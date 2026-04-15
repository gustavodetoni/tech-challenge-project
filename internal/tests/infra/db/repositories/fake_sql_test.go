package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dbOpKind string

const (
	dbOpExec  dbOpKind = "exec"
	dbOpQuery dbOpKind = "query"
)

type dbOp struct {
	kind         dbOpKind
	wantContains []string

	columns []string
	rows    [][]any

	rowsAffected int64
	err          error
}

type fakeDriver struct {
	tb  testing.TB
	mu  sync.Mutex
	ops []dbOp
	idx int
}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{d: d}, nil
}

type fakeConn struct {
	d *fakeDriver
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *fakeConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return &fakeStmt{c: c, query: query}, nil
}

func (c *fakeConn) Close() error { return nil }

func (c *fakeConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *fakeConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return &fakeTx{}, nil
}

func (c *fakeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	op, err := c.d.nextOp(dbOpExec, query)
	if err != nil {
		return nil, err
	}
	if op.err != nil {
		return nil, op.err
	}
	return fakeResult{rowsAffected: op.rowsAffected}, nil
}

func (c *fakeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	op, err := c.d.nextOp(dbOpQuery, query)
	if err != nil {
		return nil, err
	}
	if op.err != nil {
		return nil, op.err
	}
	return &fakeRows{columns: op.columns, rows: op.rows}, nil
}

func (c *fakeConn) Ping(ctx context.Context) error { return nil }

type fakeStmt struct {
	c     *fakeConn
	query string
}

func (s *fakeStmt) Close() error  { return nil }
func (s *fakeStmt) NumInput() int { return -1 }
func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), nil)
}
func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.QueryContext(context.Background(), nil)
}
func (s *fakeStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.c.ExecContext(ctx, s.query, args)
}
func (s *fakeStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.c.QueryContext(ctx, s.query, args)
}

type fakeTx struct{}

func (t *fakeTx) Commit() error   { return nil }
func (t *fakeTx) Rollback() error { return nil }

type fakeResult struct {
	rowsAffected int64
}

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type fakeRows struct {
	columns []string
	rows    [][]any
	idx     int
}

func (r *fakeRows) Columns() []string { return r.columns }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	for i := range dest {
		if i >= len(r.rows[r.idx]) {
			dest[i] = nil
			continue
		}
		dest[i] = r.rows[r.idx][i]
	}
	r.idx++
	return nil
}

func (d *fakeDriver) nextOp(kind dbOpKind, query string) (dbOp, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.idx >= len(d.ops) {
		return dbOp{}, fmt.Errorf("unexpected %s: %s", kind, query)
	}
	op := d.ops[d.idx]
	d.idx++
	if op.kind != kind {
		return dbOp{}, fmt.Errorf("expected op kind %s, got %s (query: %s)", op.kind, kind, query)
	}
	for _, s := range op.wantContains {
		if !strings.Contains(query, s) {
			return dbOp{}, fmt.Errorf("query does not contain %q: %s", s, query)
		}
	}
	return op, nil
}

func newTestGormDB(t *testing.T, ops []dbOp) *gorm.DB {
	t.Helper()

	driverName := fmt.Sprintf("fakedb_%d", time.Now().UnixNano())
	drv := &fakeDriver{tb: t, ops: ops}
	sql.Register(driverName, drv)

	sqlDB, err := sql.Open(driverName, "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn:             sqlDB,
		WithoutReturning: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return gdb
}
