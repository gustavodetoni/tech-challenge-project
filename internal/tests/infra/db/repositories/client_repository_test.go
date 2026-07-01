package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	repositories2 "github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/stretchr/testify/require"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
)

func TestClientRepository_Create_NilClient(t *testing.T) {
	repo := repositories2.NewClientRepository(nil)
	err := repo.Create(context.Background(), nil)
	require.Error(t, err)
}

func TestClientRepository_Create_Conflict(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `clients`},
			err:          &pgconn.PgError{Code: "23505"},
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	now := time.Now().UTC()
	err := repo.Create(context.Background(), &client.Client{
		ID:             "c1",
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria",
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	require.ErrorIs(t, err, repository.ErrConflict)
}

func TestClientRepository_Create_Success(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`INSERT`, `clients`},
			rowsAffected: 1,
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	now := time.Now().UTC()
	err := repo.Create(context.Background(), &client.Client{
		ID:             "c1",
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria",
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	require.NoError(t, err)
}

func TestClientRepository_Update_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpExec,
			wantContains: []string{`UPDATE`, `clients`},
			rowsAffected: 0,
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	now := time.Now().UTC()
	err := repo.Update(context.Background(), &client.Client{
		ID:             "c1",
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria",
		UpdatedAt:      now,
	})
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestClientRepository_FindByID_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "clients"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	_, err := repo.FindByID(context.Background(), "c1")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestClientRepository_FindByID_Success_UppercasesDocumentType(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "clients"`},
			columns: []string{
				"id", "document_type", "document_number", "name", "email", "phone", "created_at", "updated_at", "deleted_at",
			},
			rows: [][]any{
				{"c1", "cpf", "46420082412", "Maria", nil, nil, now, now, nil},
			},
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	out, err := repo.FindByID(context.Background(), "c1")
	require.NoError(t, err)
	require.Equal(t, client.DocumentTypeCPF, out.DocumentType)
}

func TestClientRepository_List_NormalizesLimitOffset(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "clients"`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	_, err := repo.List(context.Background(), 0, -10)
	require.NoError(t, err)
}

func TestClientRepository_Delete_Update_FindByDocument(t *testing.T) {
	now := time.Now().UTC()
	gdb := newTestGormDB(t, []dbOp{
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `clients`}, rowsAffected: 0}, // Delete -> not found
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `clients`}, rowsAffected: 1}, // Delete -> ok
		{kind: dbOpExec, wantContains: []string{`UPDATE`, `clients`}, rowsAffected: 1}, // Update -> ok
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "clients"`, `document_number`},
			columns:      []string{"id", "document_type", "document_number", "name", "created_at", "updated_at"},
			rows:         [][]any{{"c1", "CPF", "46420082412", "Maria", now, now}},
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	err := repo.Delete(context.Background(), "c1")
	require.ErrorIs(t, err, repository.ErrNotFound)

	err = repo.Delete(context.Background(), "c1")
	require.NoError(t, err)

	err = repo.Update(context.Background(), &client.Client{
		ID:             "c1",
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria",
		UpdatedAt:      now,
	})
	require.NoError(t, err)

	out, err := repo.FindByDocument(context.Background(), "46420082412")
	require.NoError(t, err)
	require.Equal(t, "c1", out.ID)
}

func TestClientRepository_FindByDocument_NotFound(t *testing.T) {
	gdb := newTestGormDB(t, []dbOp{
		{
			kind:         dbOpQuery,
			wantContains: []string{`FROM "clients"`, `document_number`},
			columns:      []string{"id"},
			rows:         [][]any{},
		},
	})
	repo := repositories2.NewClientRepository(gdb)

	_, err := repo.FindByDocument(context.Background(), "x")
	require.ErrorIs(t, err, repository.ErrNotFound)
}
