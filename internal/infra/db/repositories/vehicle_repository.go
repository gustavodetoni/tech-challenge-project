package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
)

type VehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) *VehicleRepository { return &VehicleRepository{db: db} }

type vehicleRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ClientID        string     `gorm:"column:client_id"`
	Plate           string     `gorm:"column:plate"`
	Brand           string     `gorm:"column:brand"`
	Model           string     `gorm:"column:model"`
	ManufactureYear *int       `gorm:"column:manufacture_year"`
	ModelYear       int        `gorm:"column:model_year"`
	Color           *string    `gorm:"column:color"`
	Mileage         *int       `gorm:"column:mileage"`
	Chassis         *string    `gorm:"column:chassis"`
	Notes           *string    `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (vehicleRow) TableName() string { return "vehicles" }

func (r *VehicleRepository) Create(ctx context.Context, v *vehicle.Vehicle) error {
	if v == nil {
		return errors.New("vehicle is required")
	}

	row := vehicleRow{
		ID:              v.ID,
		ClientID:        v.ClientID,
		Plate:           v.Plate,
		Brand:           v.Brand,
		Model:           v.Model,
		ManufactureYear: v.ManufactureYear,
		ModelYear:       v.ModelYear,
		Color:           v.Color,
		Mileage:         v.Mileage,
		Chassis:         v.Chassis,
		Notes:           v.Notes,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
		DeletedAt:       v.DeletedAt,
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

func (r *VehicleRepository) Update(ctx context.Context, v *vehicle.Vehicle) error {
	if v == nil {
		return errors.New("vehicle is required")
	}

	updates := map[string]any{
		"client_id":        v.ClientID,
		"plate":            v.Plate,
		"brand":            v.Brand,
		"model":            v.Model,
		"manufacture_year": v.ManufactureYear,
		"model_year":       v.ModelYear,
		"color":            v.Color,
		"mileage":          v.Mileage,
		"chassis":          v.Chassis,
		"notes":            v.Notes,
		"updated_at":       v.UpdatedAt,
	}
	tx := r.db.WithContext(ctx).
		Model(&vehicleRow{}).
		Where("id = ? AND deleted_at IS NULL", v.ID).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *VehicleRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&vehicleRow{}).
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

func (r *VehicleRepository) FindByID(ctx context.Context, id string) (*vehicle.Vehicle, error) {
	var row vehicleRow
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

	return mapVehicleRowToDomain(row), nil
}

func (r *VehicleRepository) FindByPlate(ctx context.Context, plate string) (*vehicle.Vehicle, error) {
	var row vehicleRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("plate = ?", plate).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return mapVehicleRowToDomain(row), nil
}

func (r *VehicleRepository) ListByClientID(ctx context.Context, clientID string, limit, offset int) ([]vehicle.Vehicle, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var rows []vehicleRow
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("client_id = ?", clientID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]vehicle.Vehicle, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapVehicleRowToDomain(row))
	}
	return out, nil
}

func mapVehicleRowToDomain(row vehicleRow) *vehicle.Vehicle {
	return &vehicle.Vehicle{
		ID:              row.ID,
		ClientID:        row.ClientID,
		Plate:           row.Plate,
		Brand:           row.Brand,
		Model:           row.Model,
		ManufactureYear: row.ManufactureYear,
		ModelYear:       row.ModelYear,
		Color:           row.Color,
		Mileage:         row.Mileage,
		Chassis:         row.Chassis,
		Notes:           row.Notes,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		DeletedAt:       row.DeletedAt,
	}
}
