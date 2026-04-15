package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type PartRepository struct {
	db *gorm.DB
}

func NewPartRepository(db *gorm.DB) *PartRepository { return &PartRepository{db: db} }

type partRow struct {
	ID             string     `gorm:"column:id;type:uuid;primaryKey"`
	SKU            string     `gorm:"column:sku"`
	Name           string     `gorm:"column:name"`
	Description    *string    `gorm:"column:description"`
	UnitPriceCents int64      `gorm:"column:unit_price_cents"`
	StockQuantity  int        `gorm:"column:stock_quantity"`
	Active         bool       `gorm:"column:active"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (partRow) TableName() string { return "parts" }

type stockMovementRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	PartID          string     `gorm:"column:part_id"`
	MovementType    string     `gorm:"column:movement_type"`
	Quantity        int        `gorm:"column:quantity"`
	UnitCostCents   *int64     `gorm:"column:unit_cost_cents"`
	ReferenceType   *string    `gorm:"column:reference_type"`
	ReferenceID     *string    `gorm:"column:reference_id"`
	Notes           *string    `gorm:"column:notes"`
	CreatedByUserID *string    `gorm:"column:created_by_user_id"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (stockMovementRow) TableName() string { return "stock_movements" }

func (r *PartRepository) Create(ctx context.Context, p *part.Part) error {
	if p == nil {
		return errors.New("part is required")
	}

	row := partRow{
		ID:             p.ID,
		SKU:            p.SKU,
		Name:           p.Name,
		Description:    p.Description,
		UnitPriceCents: p.UnitPriceCents,
		StockQuantity:  p.StockQuantity,
		Active:         p.Active,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
		DeletedAt:      p.DeletedAt,
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

func (r *PartRepository) Update(ctx context.Context, p *part.Part) error {
	if p == nil {
		return errors.New("part is required")
	}

	updates := map[string]any{
		"sku":              p.SKU,
		"name":             p.Name,
		"description":      p.Description,
		"unit_price_cents": p.UnitPriceCents,
		"active":           p.Active,
		"updated_at":       p.UpdatedAt,
	}
	tx := r.db.WithContext(ctx).
		Model(&partRow{}).
		Where("id = ? AND deleted_at IS NULL", p.ID).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *PartRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&partRow{}).
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

func (r *PartRepository) FindByID(ctx context.Context, id string) (*part.Part, error) {
	var row partRow
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

	return mapPart(row), nil
}

func (r *PartRepository) FindByIDs(ctx context.Context, ids []string) ([]part.Part, error) {
	if len(ids) == 0 {
		return []part.Part{}, nil
	}

	var rows []partRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id IN ?", ids).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]part.Part, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapPart(row))
	}
	return out, nil
}

func (r *PartRepository) List(ctx context.Context, limit, offset int) ([]part.Part, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var rows []partRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]part.Part, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapPart(row))
	}
	return out, nil
}

func (r *PartRepository) AdjustStock(ctx context.Context, partID string, movementType part.StockMovementType, quantity int, notes *string, createdByUserID *string) (*part.Part, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be > 0")
	}

	var updated *part.Part
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p partRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("deleted_at IS NULL").
			Where("id = ?", partID).
			First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrNotFound
			}
			return err
		}

		var newQty int
		switch movementType {
		case part.StockMovementIn:
			newQty = p.StockQuantity + quantity
		case part.StockMovementOut:
			if p.StockQuantity < quantity {
				return errors.New("insufficient stock")
			}
			newQty = p.StockQuantity - quantity
		case part.StockMovementAdjustment:
			newQty = quantity
		default:
			return errors.New("invalid movement type")
		}

		now := time.Now().UTC()
		if err := tx.Model(&partRow{}).
			Where("id = ? AND deleted_at IS NULL", partID).
			Updates(map[string]any{"stock_quantity": newQty, "updated_at": now}).Error; err != nil {
			return err
		}

		m := stockMovementRow{
			ID:              uuid.NewString(),
			PartID:          partID,
			MovementType:    string(movementType),
			Quantity:        quantity,
			Notes:           notes,
			CreatedByUserID: createdByUserID,
			CreatedAt:       now,
		}
		if err := tx.Create(&m).Error; err != nil {
			return err
		}

		p.StockQuantity = newQty
		p.UpdatedAt = now
		updated = mapPart(p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func mapPart(row partRow) *part.Part {
	return &part.Part{
		ID:             row.ID,
		SKU:            row.SKU,
		Name:           row.Name,
		Description:    row.Description,
		UnitPriceCents: row.UnitPriceCents,
		StockQuantity:  row.StockQuantity,
		Active:         row.Active,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		DeletedAt:      row.DeletedAt,
	}
}
