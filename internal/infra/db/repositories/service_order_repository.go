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

func NewServiceOrderRepository(db *gorm.DB) *ServiceOrderRepository {
	return &ServiceOrderRepository{db: db}
}

type serviceOrderRow struct {
	ID                 string     `gorm:"column:id;type:uuid;primaryKey"`
	Code               string     `gorm:"column:code"`
	ClientID           string     `gorm:"column:client_id"`
	VehicleID          string     `gorm:"column:vehicle_id"`
	AssignedUserID     *string    `gorm:"column:assigned_user_id"`
	Status             string     `gorm:"column:status"`
	OpenedAt           time.Time  `gorm:"column:opened_at"`
	ExecutionStartedAt *time.Time `gorm:"column:execution_started_at"`
	FinishedAt         *time.Time `gorm:"column:finished_at"`
	DeliveredAt        *time.Time `gorm:"column:delivered_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
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

type budgetDetailRow struct {
	ID               string     `gorm:"column:id"`
	ServiceOrderID   string     `gorm:"column:service_order_id"`
	Version          int        `gorm:"column:version"`
	Status           string     `gorm:"column:status"`
	TotalAmountCents int64      `gorm:"column:total_amount_cents"`
	SentAt           *time.Time `gorm:"column:sent_at"`
	ApprovedAt       *time.Time `gorm:"column:approved_at"`
	RejectedAt       *time.Time `gorm:"column:rejected_at"`
	ApprovedByName   *string    `gorm:"column:approved_by_name"`
	RejectionReason  *string    `gorm:"column:rejection_reason"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (budgetDetailRow) TableName() string { return "budgets" }

type budgetServiceDetailRow struct {
	ServiceID       *string    `gorm:"column:service_id"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (budgetServiceDetailRow) TableName() string { return "budget_services" }

type budgetPartDetailRow struct {
	PartID          *string    `gorm:"column:part_id"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (budgetPartDetailRow) TableName() string { return "budget_parts" }

type statusHistoryDetailRow struct {
	FromStatus      *string    `gorm:"column:from_status"`
	ToStatus        string     `gorm:"column:to_status"`
	ChangedByUserID *string    `gorm:"column:changed_by_user_id"`
	Reason          *string    `gorm:"column:reason"`
	ChangedAt       time.Time  `gorm:"column:changed_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (statusHistoryDetailRow) TableName() string { return "service_order_status_history" }

func (r *ServiceOrderRepository) GetDetailByID(ctx context.Context, id string) (*order.ServiceOrderDetail, error) {
	so, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var b budgetDetailRow
	budgetErr := r.db.WithContext(ctx).
		Where("deleted_at is null").
		Where("service_order_id = ?", so.ID).
		Order("version desc").
		First(&b).Error
	if budgetErr != nil && !errors.Is(budgetErr, gorm.ErrRecordNotFound) {
		return nil, budgetErr
	}

	var latestBudget *order.Budget
	if budgetErr == nil {
		latestBudget = &order.Budget{
			ID:               b.ID,
			ServiceOrderID:   b.ServiceOrderID,
			Version:          b.Version,
			Status:           order.BudgetStatus(b.Status),
			TotalAmountCents: b.TotalAmountCents,
			SentAt:           b.SentAt,
			ApprovedAt:       b.ApprovedAt,
			RejectedAt:       b.RejectedAt,
			ApprovedByName:   b.ApprovedByName,
			RejectionReason:  b.RejectionReason,
			CreatedAt:        b.CreatedAt,
			UpdatedAt:        b.UpdatedAt,
			DeletedAt:        b.DeletedAt,
		}
	}

	budgetServices := []order.BudgetServiceItem{}
	budgetParts := []order.BudgetPartItem{}
	if latestBudget != nil {
		var bs []budgetServiceDetailRow
		if err := r.db.WithContext(ctx).
			Where("deleted_at is null").
			Where("budget_id = ?", latestBudget.ID).
			Order("created_at asc").
			Find(&bs).Error; err != nil {
			return nil, err
		}
		budgetServices = make([]order.BudgetServiceItem, 0, len(bs))
		for _, it := range bs {
			serviceID := ""
			if it.ServiceID != nil {
				serviceID = *it.ServiceID
			}
			budgetServices = append(budgetServices, order.BudgetServiceItem{
				ServiceID:       serviceID,
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
			})
		}

		var bp []budgetPartDetailRow
		if err := r.db.WithContext(ctx).
			Where("deleted_at is null").
			Where("budget_id = ?", latestBudget.ID).
			Order("created_at asc").
			Find(&bp).Error; err != nil {
			return nil, err
		}
		budgetParts = make([]order.BudgetPartItem, 0, len(bp))
		for _, it := range bp {
			partID := ""
			if it.PartID != nil {
				partID = *it.PartID
			}
			budgetParts = append(budgetParts, order.BudgetPartItem{
				PartID:          partID,
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
			})
		}
	}

	var histRows []statusHistoryDetailRow
	if err := r.db.WithContext(ctx).
		Where("deleted_at is null").
		Where("service_order_id = ?", so.ID).
		Order("changed_at asc").
		Find(&histRows).Error; err != nil {
		return nil, err
	}
	hist := make([]order.StatusHistoryEntry, 0, len(histRows))
	for _, it := range histRows {
		var from *order.Status
		if it.FromStatus != nil {
			s := order.Status(*it.FromStatus)
			from = &s
		}
		hist = append(hist, order.StatusHistoryEntry{
			FromStatus:      from,
			ToStatus:        order.Status(it.ToStatus),
			ChangedAt:       it.ChangedAt,
			ChangedByUserID: it.ChangedByUserID,
			Reason:          it.Reason,
		})
	}

	return &order.ServiceOrderDetail{
		ServiceOrder:   *so,
		LatestBudget:   latestBudget,
		BudgetServices: budgetServices,
		BudgetParts:    budgetParts,
		StatusHistory:  hist,
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

func (r *ServiceOrderRepository) AverageServiceExecutionMinutes(ctx context.Context, serviceID *string, from, to *time.Time) ([]repository.ServiceExecutionAverage, error) {
	type row struct {
		ServiceID   *string `gorm:"column:service_id"`
		Description string  `gorm:"column:description"`
		AvgMinutes  float64 `gorm:"column:avg_minutes"`
		SampleCount int64   `gorm:"column:sample_count"`
	}

	q := r.db.WithContext(ctx).
		Table("service_order_services sos").
		Joins("join service_orders so on so.id = sos.service_order_id and so.deleted_at is null").
		Select("sos.service_id, sos.description, coalesce(avg(extract(epoch from (sos.completed_at - sos.started_at)))/60.0, 0) as avg_minutes, count(*) as sample_count").
		Where("sos.deleted_at is null").
		Where("sos.started_at is not null").
		Where("sos.completed_at is not null")
	if serviceID != nil && *serviceID != "" {
		q = q.Where("sos.service_id = ?", *serviceID)
	}
	if from != nil {
		q = q.Where("sos.completed_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("sos.completed_at <= ?", *to)
	}
	q = q.Group("sos.service_id, sos.description").Order("avg_minutes desc")

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]repository.ServiceExecutionAverage, 0, len(rows))
	for _, it := range rows {
		out = append(out, repository.ServiceExecutionAverage{
			ServiceID:      it.ServiceID,
			Description:    it.Description,
			AverageMinutes: it.AvgMinutes,
			SampleCount:    it.SampleCount,
		})
	}
	return out, nil
}
