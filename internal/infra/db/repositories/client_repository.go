package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository { return &ClientRepository{db: db} }

type clientRow struct {
	ID             string     `gorm:"column:id;type:uuid;primaryKey"`
	DocumentType   string     `gorm:"column:document_type"`
	DocumentNumber string     `gorm:"column:document_number"`
	Name           string     `gorm:"column:name"`
	Email          *string    `gorm:"column:email"`
	Phone          *string    `gorm:"column:phone"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (clientRow) TableName() string { return "clients" }

func (r *ClientRepository) Create(ctx context.Context, c *client.Client) error {
	if c == nil {
		return errors.New("client is required")
	}
	row := clientRow{
		ID:             c.ID,
		DocumentType:   string(c.DocumentType),
		DocumentNumber: c.DocumentNumber,
		Name:           c.Name,
		Email:          c.Email,
		Phone:          c.Phone,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		DeletedAt:      c.DeletedAt,
	}

	err := r.db.WithContext(ctx).Create(&row).Error
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return repository.ErrConflict
	}
	return err
}

func (r *ClientRepository) Update(ctx context.Context, c *client.Client) error {
	if c == nil {
		return errors.New("client is required")
	}

	updates := map[string]any{
		"document_type":   string(c.DocumentType),
		"document_number": c.DocumentNumber,
		"name":            c.Name,
		"email":           c.Email,
		"phone":           c.Phone,
		"updated_at":      c.UpdatedAt,
	}
	tx := r.db.WithContext(ctx).
		Model(&clientRow{}).
		Where("id = ? AND deleted_at IS NULL", c.ID).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *ClientRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&clientRow{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *ClientRepository) FindByID(ctx context.Context, id string) (*client.Client, error) {
	var row clientRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id = ?", id).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return mapClientRowToDomain(row), nil
}

func (r *ClientRepository) FindByDocument(ctx context.Context, documentNumber string) (*client.Client, error) {
	var row clientRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("document_number = ?", documentNumber).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return mapClientRowToDomain(row), nil
}

func (r *ClientRepository) List(ctx context.Context, limit, offset int) ([]client.Client, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var rows []clientRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]client.Client, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapClientRowToDomain(row))
	}
	return out, nil
}

func mapClientRowToDomain(row clientRow) *client.Client {
	return &client.Client{
		ID:             row.ID,
		DocumentType:   client.DocumentType(strings.ToUpper(row.DocumentType)),
		DocumentNumber: row.DocumentNumber,
		Name:           row.Name,
		Email:          row.Email,
		Phone:          row.Phone,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		DeletedAt:      row.DeletedAt,
	}
}
