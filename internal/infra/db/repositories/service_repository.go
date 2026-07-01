package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository { return &ServiceRepository{db: db} }

type serviceRow struct {
	ID               string     `gorm:"column:id;type:uuid;primaryKey"`
	Name             string     `gorm:"column:name"`
	Description      *string    `gorm:"column:description"`
	BasePriceCents   int64      `gorm:"column:base_price_cents"`
	EstimatedMinutes int        `gorm:"column:estimated_minutes"`
	Active           bool       `gorm:"column:active"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (serviceRow) TableName() string { return "services" }

func (r *ServiceRepository) Create(ctx context.Context, s *service.Service) error {
	if s == nil {
		return errors.New("service is required")
	}
	row := serviceRow{
		ID:               s.ID,
		Name:             s.Name,
		Description:      s.Description,
		BasePriceCents:   s.BasePriceCents,
		EstimatedMinutes: s.EstimatedMinutes,
		Active:           s.Active,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		DeletedAt:        s.DeletedAt,
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

func (r *ServiceRepository) Update(ctx context.Context, s *service.Service) error {
	if s == nil {
		return errors.New("service is required")
	}

	updates := map[string]any{
		"name":              s.Name,
		"description":       s.Description,
		"base_price_cents":  s.BasePriceCents,
		"estimated_minutes": s.EstimatedMinutes,
		"active":            s.Active,
		"updated_at":        s.UpdatedAt,
	}
	tx := r.db.WithContext(ctx).
		Model(&serviceRow{}).
		Where("id = ? AND deleted_at IS NULL", s.ID).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *ServiceRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&serviceRow{}).
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

func (r *ServiceRepository) FindByID(ctx context.Context, id string) (*service.Service, error) {
	var row serviceRow
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
	return mapServiceRowToDomain(row), nil
}

func (r *ServiceRepository) FindByIDs(ctx context.Context, ids []string) ([]service.Service, error) {
	if len(ids) == 0 {
		return []service.Service{}, nil
	}

	var rows []serviceRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id IN ?", ids).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]service.Service, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapServiceRowToDomain(row))
	}
	return out, nil
}

func (r *ServiceRepository) List(ctx context.Context, limit, offset int) ([]service.Service, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var rows []serviceRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]service.Service, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapServiceRowToDomain(row))
	}
	return out, nil
}

func mapServiceRowToDomain(row serviceRow) *service.Service {
	return &service.Service{
		ID:               row.ID,
		Name:             row.Name,
		Description:      row.Description,
		BasePriceCents:   row.BasePriceCents,
		EstimatedMinutes: row.EstimatedMinutes,
		Active:           row.Active,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		DeletedAt:        row.DeletedAt,
	}
}
