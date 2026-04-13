package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type ServiceOrderRepository struct {
	db *gorm.DB
}

func NewServiceOrderRepository(db *gorm.DB) *ServiceOrderRepository { return &ServiceOrderRepository{db: db} }

type serviceOrderRow struct {
	ID                string     `gorm:"column:id;type:uuid;primaryKey"`
	Code              string     `gorm:"column:code"`
	ClientID          string     `gorm:"column:client_id"`
	VehicleID         string     `gorm:"column:vehicle_id"`
	AssignedUserID    *string    `gorm:"column:assigned_user_id"`
	Status            string     `gorm:"column:status"`
	OpenedAt          time.Time  `gorm:"column:opened_at"`
	ExecutionStartedAt *time.Time `gorm:"column:execution_started_at"`
	FinishedAt        *time.Time `gorm:"column:finished_at"`
	DeliveredAt       *time.Time `gorm:"column:delivered_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderRow) TableName() string { return "service_orders" }

func (r *ServiceOrderRepository) List(ctx context.Context, limit, offset int, status *order.Status) ([]order.ServiceOrderSummary, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	type row struct {
		ID         string
		Code       string
		Status     string
		OpenedAt   time.Time
		ClientID   string
		ClientName string
		VehicleID  string
		Plate      string
	}

	q := r.db.WithContext(ctx).
		Table("service_orders so").
		Select("so.id, so.code, so.status, so.opened_at, c.id as client_id, c.name as client_name, v.id as vehicle_id, v.plate").
		Joins("join clients c on c.id = so.client_id and c.deleted_at is null").
		Joins("join vehicles v on v.id = so.vehicle_id and v.deleted_at is null").
		Where("so.deleted_at is null").
		Order("so.opened_at desc").
		Limit(limit).
		Offset(offset)
	if status != nil && *status != "" {
		q = q.Where("so.status = ?", string(*status))
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]order.ServiceOrderSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, order.ServiceOrderSummary{
			ID:         r.ID,
			Code:       r.Code,
			Status:     order.Status(r.Status),
			OpenedAt:   r.OpenedAt,
			ClientID:   r.ClientID,
			ClientName: r.ClientName,
			VehicleID:  r.VehicleID,
			Plate:      r.Plate,
		})
	}
	return out, nil
}

func (r *ServiceOrderRepository) FindByID(ctx context.Context, id string) (*order.ServiceOrder, error) {
	var row serviceOrderRow
	err := r.db.WithContext(ctx).
		Where("deleted_at is null").
		Where("id = ?", id).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &order.ServiceOrder{
		ID:             row.ID,
		Code:           row.Code,
		ClientID:       row.ClientID,
		VehicleID:      row.VehicleID,
		AssignedUserID: row.AssignedUserID,
		Status:         order.Status(row.Status),
		OpenedAt:       row.OpenedAt,
		ExecutionStart: row.ExecutionStartedAt,
		FinishedAt:     row.FinishedAt,
		DeliveredAt:    row.DeliveredAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		DeletedAt:      row.DeletedAt,
	}, nil
}

func (r *ServiceOrderRepository) AverageExecutionMinutes(ctx context.Context, from, to *time.Time) (float64, error) {
	q := r.db.WithContext(ctx).
		Table("service_orders so").
		Select("coalesce(avg(extract(epoch from (so.finished_at - so.execution_started_at)))/60.0, 0) as avg_minutes").
		Where("so.deleted_at is null").
		Where("so.execution_started_at is not null").
		Where("so.finished_at is not null")
	if from != nil {
		q = q.Where("so.finished_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("so.finished_at <= ?", *to)
	}

	var avg float64
	if err := q.Scan(&avg).Error; err != nil {
		return 0, err
	}
	return avg, nil
}

