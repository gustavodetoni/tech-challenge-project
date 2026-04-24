package seed_test

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

	seed "github.com/soat-architecture/tech-challenge-project/internal/infra/db/seed"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type recordingDriver struct {
	mu sync.Mutex

	execQueries  []string
	queryQueries []string

	seedMarkerCount int64
}

func (d *recordingDriver) Open(name string) (driver.Conn, error) {
	return &recordingConn{d: d}, nil
}

func (d *recordingDriver) recordExec(q string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.execQueries = append(d.execQueries, q)
}

func (d *recordingDriver) recordQuery(q string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.queryQueries = append(d.queryQueries, q)
}

func (d *recordingDriver) ExecQueries() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]string, 0, len(d.execQueries))
	out = append(out, d.execQueries...)
	return out
}

func (d *recordingDriver) QueryQueries() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]string, 0, len(d.queryQueries))
	out = append(out, d.queryQueries...)
	return out
}

type recordingConn struct {
	d *recordingDriver
}

func (c *recordingConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c *recordingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return &recordingStmt{c: c, query: query}, nil
}

func (c *recordingConn) Close() error { return nil }

func (c *recordingConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *recordingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return &recordingTx{}, nil
}

func (c *recordingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.d.recordExec(query)
	return recordingResult{rowsAffected: 1}, nil
}

func (c *recordingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.d.recordQuery(query)

	if strings.Contains(strings.ToLower(query), "count(") {
		return &recordingRows{
			columns: []string{"count"},
			rows:    [][]any{{c.d.seedMarkerCount}},
		}, nil
	}

	return &recordingRows{columns: []string{}, rows: nil}, nil
}

func (c *recordingConn) Ping(ctx context.Context) error { return nil }

type recordingStmt struct {
	c     *recordingConn
	query string
}

func (s *recordingStmt) Close() error  { return nil }
func (s *recordingStmt) NumInput() int { return -1 }
func (s *recordingStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), nil)
}
func (s *recordingStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.QueryContext(context.Background(), nil)
}
func (s *recordingStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.c.ExecContext(ctx, s.query, args)
}
func (s *recordingStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.c.QueryContext(ctx, s.query, args)
}

type recordingTx struct{}

func (t *recordingTx) Commit() error   { return nil }
func (t *recordingTx) Rollback() error { return nil }

type recordingResult struct {
	rowsAffected int64
}

func (r recordingResult) LastInsertId() (int64, error) { return 0, nil }
func (r recordingResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type recordingRows struct {
	columns []string
	rows    [][]any
	idx     int
}

func (r *recordingRows) Columns() []string { return r.columns }
func (r *recordingRows) Close() error      { return nil }
func (r *recordingRows) Next(dest []driver.Value) error {
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

func newTestGormDB(t *testing.T, drv *recordingDriver) *gorm.DB {
	t.Helper()

	driverName := fmt.Sprintf("recorddb_%d", time.Now().UnixNano())
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

func containsFold(s string, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func TestRun_NilDB(t *testing.T) {
	err := seed.Run(context.Background(), nil, seed.Options{Users: 1, Clients: 1, Services: 1, Parts: 1, Orders: 1, Now: time.Now().UTC()})
	require.ErrorContains(t, err, "gormDB is required")
}

func TestRun_InvalidOptions(t *testing.T) {
	drv := &recordingDriver{}
	gdb := newTestGormDB(t, drv)

	err := seed.Run(context.Background(), gdb, seed.Options{
		Users:    -1,
		Clients:  1,
		Services: 1,
		Parts:    1,
		Orders:   1,
		Now:      time.Now().UTC(),
	})
	require.ErrorContains(t, err, "SEED_USERS must be >= 1")
}

func TestRun_ForceFalse_AlreadySeeded_Skips(t *testing.T) {
	drv := &recordingDriver{seedMarkerCount: 1}
	gdb := newTestGormDB(t, drv)

	opts := seed.Options{
		Force:      false,
		RandomSeed: 1,
		Users:      1,
		Clients:    1,
		Services:   1,
		Parts:      1,
		Orders:     1,
		Now:        time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
	}

	require.NoError(t, seed.Run(context.Background(), gdb, opts))

	queries := drv.QueryQueries()
	require.NotEmpty(t, queries)
	require.True(t, containsFold(strings.Join(queries, "\n"), "users"))

	execs := drv.ExecQueries()
	require.False(t, containsFold(strings.Join(execs, "\n"), "truncate"))
}

func TestRun_ForceTrue_FullSeed_SucceedsAndTruncates(t *testing.T) {
	drv := &recordingDriver{seedMarkerCount: 0}
	gdb := newTestGormDB(t, drv)

	opts := seed.Options{
		Force:      true,
		RandomSeed: 123,
		Users:      2,
		Clients:    2,
		Services:   2,
		Parts:      2,
		Orders:     2,
		Now:        time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
	}

	require.NoError(t, seed.Run(context.Background(), gdb, opts))

	execs := strings.Join(drv.ExecQueries(), "\n")
	require.True(t, containsFold(execs, "truncate table"))
	require.True(t, containsFold(execs, "users"))
	require.True(t, containsFold(execs, "insert"))
	require.True(t, containsFold(execs, "clients"))
	require.True(t, containsFold(execs, "vehicles"))
	require.True(t, containsFold(execs, "services"))
	require.True(t, containsFold(execs, "parts"))
	require.True(t, containsFold(execs, "service_orders"))
	require.True(t, containsFold(execs, "service_order_status_history"))
	require.True(t, containsFold(execs, "budgets"))
	require.True(t, containsFold(execs, "budget_services"))
	// With Orders=2, i=1 produces parts in budget.
	require.True(t, containsFold(execs, "budget_parts"))
	require.True(t, containsFold(execs, "service_order_services"))
	require.True(t, containsFold(execs, "service_order_parts"))
}

func TestRun_ForceFalse_NotSeeded_Seeds(t *testing.T) {
	drv := &recordingDriver{seedMarkerCount: 0}
	gdb := newTestGormDB(t, drv)

	opts := seed.Options{
		Force:      false,
		RandomSeed: 123,
		Users:      2,
		Clients:    2,
		Services:   2,
		Parts:      2,
		Orders:     2,
		Now:        time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
	}

	require.NoError(t, seed.Run(context.Background(), gdb, opts))

	execs := strings.Join(drv.ExecQueries(), "\n")
	require.False(t, containsFold(execs, "truncate table"))
	require.True(t, containsFold(execs, "users"))
	require.True(t, containsFold(execs, "service_orders"))
}
