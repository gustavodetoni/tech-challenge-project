package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type ServiceOrderFlowRepository struct {
	db *gorm.DB
}

func NewServiceOrderFlowRepository(db *gorm.DB) *ServiceOrderFlowRepository {
	return &ServiceOrderFlowRepository{db: db}
}

type budgetRow struct {
	ID               string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID   string     `gorm:"column:service_order_id"`
	Version          int        `gorm:"column:version"`
	Status           string     `gorm:"column:status"`
	TotalAmountCents int64      `gorm:"column:total_amount_cents"`
	Notes            *string    `gorm:"column:notes"`
	SentAt           *time.Time `gorm:"column:sent_at"`
	DecidedAt        *time.Time `gorm:"column:decided_at"`
	ApprovedAt       *time.Time `gorm:"column:approved_at"`
	RejectedAt       *time.Time `gorm:"column:rejected_at"`
	ApprovedByName   *string    `gorm:"column:approved_by_name"`
	RejectionReason  *string    `gorm:"column:rejection_reason"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (budgetRow) TableName() string { return "budgets" }

type budgetServiceRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	BudgetID        string     `gorm:"column:budget_id"`
	ServiceID       *string    `gorm:"column:service_id"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (budgetServiceRow) TableName() string { return "budget_services" }

type budgetPartRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	BudgetID        string     `gorm:"column:budget_id"`
	PartID          *string    `gorm:"column:part_id"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (budgetPartRow) TableName() string { return "budget_parts" }

type serviceOrderStatusHistoryRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID  string     `gorm:"column:service_order_id"`
	FromStatus      *string    `gorm:"column:from_status"`
	ToStatus        string     `gorm:"column:to_status"`
	ChangedByUserID *string    `gorm:"column:changed_by_user_id"`
	Reason          *string    `gorm:"column:reason"`
	ChangedAt       time.Time  `gorm:"column:changed_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderStatusHistoryRow) TableName() string { return "service_order_status_history" }

type serviceOrderServiceRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID  string     `gorm:"column:service_order_id"`
	ServiceID       *string    `gorm:"column:service_id"`
	BudgetServiceID *string    `gorm:"column:budget_service_id"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	Status          string     `gorm:"column:status"`
	AssignedUserID  *string    `gorm:"column:assigned_user_id"`
	StartedAt       *time.Time `gorm:"column:started_at"`
	CompletedAt     *time.Time `gorm:"column:completed_at"`
	Notes           *string    `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderServiceRow) TableName() string { return "service_order_services" }

type serviceOrderPartRow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	ServiceOrderID  string     `gorm:"column:service_order_id"`
	PartID          *string    `gorm:"column:part_id"`
	BudgetPartID    *string    `gorm:"column:budget_part_id"`
	Description     string     `gorm:"column:description"`
	Quantity        int        `gorm:"column:quantity"`
	UnitPriceCents  int64      `gorm:"column:unit_price_cents"`
	TotalPriceCents int64      `gorm:"column:total_price_cents"`
	Status          string     `gorm:"column:status"`
	Notes           *string    `gorm:"column:notes"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (serviceOrderPartRow) TableName() string { return "service_order_parts" }

type partRowForUpdate struct {
	ID            string     `gorm:"column:id;type:uuid;primaryKey"`
	StockQuantity int        `gorm:"column:stock_quantity"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

func (partRowForUpdate) TableName() string { return "parts" }

type stockMovementRowForFlow struct {
	ID              string     `gorm:"column:id;type:uuid;primaryKey"`
	PartID          string     `gorm:"column:part_id"`
	MovementType    string     `gorm:"column:movement_type"`
	Quantity        int        `gorm:"column:quantity"`
	ReferenceType   *string    `gorm:"column:reference_type"`
	ReferenceID     *string    `gorm:"column:reference_id"`
	Notes           *string    `gorm:"column:notes"`
	CreatedByUserID *string    `gorm:"column:created_by_user_id"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (stockMovementRowForFlow) TableName() string { return "stock_movements" }

func (r *ServiceOrderFlowRepository) CreateDraft(ctx context.Context, p repository.CreateServiceOrderDraftParams) (*order.ServiceOrder, *order.Budget, error) {
	now := time.Now().UTC()
	so := p.ServiceOrder
	b := p.Budget

	so.CreatedAt = now
	so.UpdatedAt = now
	b.CreatedAt = now
	b.UpdatedAt = now

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("service_orders").Create(map[string]any{
			"id":                 so.ID,
			"code":               so.Code,
			"client_id":          so.ClientID,
			"vehicle_id":         so.VehicleID,
			"assigned_user_id":   so.AssignedUserID,
			"status":             string(so.Status),
			"customer_complaint": so.CustomerComplaint,
			"opened_at":          so.OpenedAt,
			"created_at":         so.CreatedAt,
			"updated_at":         so.UpdatedAt,
		}).Error; err != nil {
			return err
		}

		h := serviceOrderStatusHistoryRow{
			ID:              uuid.NewString(),
			ServiceOrderID:  so.ID,
			FromStatus:      nil,
			ToStatus:        string(so.Status),
			ChangedByUserID: nil,
			ChangedAt:       now,
			CreatedAt:       now,
		}
		if err := tx.Create(&h).Error; err != nil {
			return err
		}

		if err := tx.Create(&budgetRow{
			ID:               b.ID,
			ServiceOrderID:   so.ID,
			Version:          b.Version,
			Status:           string(b.Status),
			TotalAmountCents: b.TotalAmountCents,
			CreatedAt:        now,
			UpdatedAt:        now,
		}).Error; err != nil {
			return err
		}

		for _, it := range p.BudgetServices {
			serviceID := it.ServiceID
			row := budgetServiceRow{
				ID:              uuid.NewString(),
				BudgetID:        b.ID,
				ServiceID:       &serviceID,
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}

		for _, it := range p.BudgetParts {
			partID := it.PartID
			row := budgetPartRow{
				ID:              uuid.NewString(),
				BudgetID:        b.ID,
				PartID:          &partID,
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, repository.ErrConflict
		}
		return nil, nil, err
	}

	createdBudget := b
	createdBudget.ServiceOrderID = so.ID
	return &so, &createdBudget, nil
}

func (r *ServiceOrderFlowRepository) CreateBudgetRevision(ctx context.Context, p repository.CreateBudgetRevisionParams) (*order.Budget, error) {
	now := time.Now().UTC()
	b := p.Budget
	b.ServiceOrderID = p.ServiceOrderID
	b.CreatedAt = now
	b.UpdatedAt = now

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var so struct {
			ID     string
			Status string
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Table("service_orders").
			Select("id, status").
			Where("id = ? AND deleted_at is null", p.ServiceOrderID).
			Limit(1).
			Scan(&so).Error; err != nil {
			return err
		}
		if so.ID == "" {
			return repository.ErrNotFound
		}
		if order.Status(so.Status) != order.StatusInDiagnosis {
			return repository.ErrConflict
		}

		latest, err := latestBudgetForUpdate(tx, p.ServiceOrderID)
		if err != nil {
			return err
		}

		b.Version = latest.Version + 1
		if b.Version < 2 {
			b.Version = 2
		}

		if err := tx.Create(&budgetRow{
			ID:               b.ID,
			ServiceOrderID:   b.ServiceOrderID,
			Version:          b.Version,
			Status:           string(b.Status),
			TotalAmountCents: b.TotalAmountCents,
			CreatedAt:        now,
			UpdatedAt:        now,
		}).Error; err != nil {
			return err
		}

		for _, it := range p.BudgetServices {
			serviceID := it.ServiceID
			row := budgetServiceRow{
				ID:              uuid.NewString(),
				BudgetID:        b.ID,
				ServiceID:       &serviceID,
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}

		for _, it := range p.BudgetParts {
			partID := it.PartID
			row := budgetPartRow{
				ID:              uuid.NewString(),
				BudgetID:        b.ID,
				PartID:          &partID,
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repository.ErrConflict
		}
		return nil, err
	}

	return &b, nil
}

func (r *ServiceOrderFlowRepository) StartDiagnosis(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	return r.transitionStatus(ctx, serviceOrderID, order.StatusInDiagnosis, changedByUserID, nil, map[string]any{
		"diagnosis_started_at": time.Now().UTC(),
	})
}

func (r *ServiceOrderFlowRepository) SendLatestBudget(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		b, err := latestBudgetForUpdate(tx, serviceOrderID)
		if err != nil {
			return err
		}
		if b.Status != string(order.BudgetStatusDraft) {
			return fmt.Errorf("budget not in DRAFT")
		}
		if err := tx.Model(&budgetRow{}).
			Where("id = ? AND deleted_at is null", b.ID).
			Updates(map[string]any{"status": string(order.BudgetStatusSent), "sent_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		return r.transitionStatusTx(tx, serviceOrderID, order.StatusWaitingApproval, changedByUserID, nil, map[string]any{
			"waiting_approval_at": now,
		})
	})
}

func (r *ServiceOrderFlowRepository) Finish(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	now := time.Now().UTC()
	return r.transitionStatus(ctx, serviceOrderID, order.StatusFinished, changedByUserID, nil, map[string]any{
		"finished_at": now,
	})
}

func (r *ServiceOrderFlowRepository) Deliver(ctx context.Context, serviceOrderID string, changedByUserID *string) error {
	now := time.Now().UTC()
	return r.transitionStatus(ctx, serviceOrderID, order.StatusDelivered, changedByUserID, nil, map[string]any{
		"delivered_at": now,
	})
}

func (r *ServiceOrderFlowRepository) GetClientViewByCode(ctx context.Context, code string, documentNumber string) (*repository.ClientServiceOrderView, error) {
	timePtrRFC3339 := func(t *time.Time) *string {
		if t == nil {
			return nil
		}
		s := t.UTC().Format(time.RFC3339)
		return &s
	}

	type soRow struct {
		ID                string
		Code              string
		Status            string
		ClientID          string
		VehicleID         string
		OpenedAt          time.Time
		CustomerComplaint *string

		Plate           string
		Brand           string
		Model           string
		ManufactureYear *int
		ModelYear       int
		Color           *string
	}

	var so soRow
	err := r.db.WithContext(ctx).
		Table("service_orders so").
		Select("so.id, so.code, so.status, so.client_id, so.vehicle_id, so.opened_at, so.customer_complaint, v.plate, v.brand, v.model, v.manufacture_year, v.model_year, v.color").
		Joins("join clients c on c.id = so.client_id and c.deleted_at is null").
		Joins("join vehicles v on v.id = so.vehicle_id and v.deleted_at is null").
		Where("so.deleted_at is null").
		Where("so.code = ?", code).
		Where("c.document_number = ?", documentNumber).
		Limit(1).
		Scan(&so).Error
	if err != nil {
		return nil, err
	}
	if so.ID == "" {
		return nil, repository.ErrNotFound
	}

	var b budgetRow
	budgetErr := r.db.WithContext(ctx).
		Table("budgets").
		Select("id, service_order_id, version, status, total_amount_cents, sent_at, approved_at, rejected_at, approved_by_name, rejection_reason").
		Where("deleted_at is null").
		Where("service_order_id = ?", so.ID).
		Order("version desc").
		Limit(1).
		Scan(&b).Error
	if budgetErr != nil {
		return nil, budgetErr
	}
	if b.ID == "" {
		return nil, repository.ErrNotFound
	}

	type budgetServiceLineRow struct {
		ServiceID       *string
		Description     string
		Quantity        int
		UnitPriceCents  int64
		TotalPriceCents int64
	}
	var bs []budgetServiceLineRow
	if err := r.db.WithContext(ctx).
		Table("budget_services").
		Select("service_id, description, quantity, unit_price_cents, total_price_cents").
		Where("deleted_at is null").
		Where("budget_id = ?", b.ID).
		Order("created_at asc").
		Scan(&bs).Error; err != nil {
		return nil, err
	}
	budgetServices := make([]order.BudgetServiceItem, 0, len(bs))
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

	type budgetPartLineRow struct {
		PartID          *string
		Description     string
		Quantity        int
		UnitPriceCents  int64
		TotalPriceCents int64
	}
	var bp []budgetPartLineRow
	if err := r.db.WithContext(ctx).
		Table("budget_parts").
		Select("part_id, description, quantity, unit_price_cents, total_price_cents").
		Where("deleted_at is null").
		Where("budget_id = ?", b.ID).
		Order("created_at asc").
		Scan(&bp).Error; err != nil {
		return nil, err
	}
	budgetParts := make([]order.BudgetPartItem, 0, len(bp))
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

	type histRow struct {
		FromStatus      *string
		ToStatus        string
		ChangedAt       time.Time
		ChangedByUserID *string
		Reason          *string
	}
	var histRows []histRow
	if err := r.db.WithContext(ctx).
		Table("service_order_status_history").
		Select("from_status, to_status, changed_at, changed_by_user_id, reason").
		Where("deleted_at is null").
		Where("service_order_id = ?", so.ID).
		Order("changed_at asc").
		Scan(&histRows).Error; err != nil {
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

	return &repository.ClientServiceOrderView{
		Code:              so.Code,
		Status:            order.Status(so.Status),
		ClientID:          so.ClientID,
		VehicleID:         so.VehicleID,
		OpenedAt:          so.OpenedAt.UTC().Format(time.RFC3339),
		CustomerComplaint: so.CustomerComplaint,

		VehiclePlate:           so.Plate,
		VehicleBrand:           so.Brand,
		VehicleModel:           so.Model,
		VehicleManufactureYear: so.ManufactureYear,
		VehicleModelYear:       so.ModelYear,
		VehicleColor:           so.Color,

		BudgetID:              b.ID,
		BudgetVersion:         b.Version,
		BudgetStatus:          order.BudgetStatus(b.Status),
		BudgetTotalCents:      b.TotalAmountCents,
		BudgetSentAt:          timePtrRFC3339(b.SentAt),
		BudgetApprovedAt:      timePtrRFC3339(b.ApprovedAt),
		BudgetRejectedAt:      timePtrRFC3339(b.RejectedAt),
		BudgetApprovedByName:  b.ApprovedByName,
		BudgetRejectionReason: b.RejectionReason,

		BudgetServices: budgetServices,
		BudgetParts:    budgetParts,
		StatusHistory:  hist,
	}, nil
}

func (r *ServiceOrderFlowRepository) ApproveLatestBudgetByCode(ctx context.Context, code string, documentNumber string, approvedByName *string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		soID, err := serviceOrderIDByCodeAndDocumentForUpdate(tx, code, documentNumber)
		if err != nil {
			return err
		}

		b, err := latestBudgetForUpdate(tx, soID)
		if err != nil {
			return err
		}
		if b.Status != string(order.BudgetStatusSent) {
			return fmt.Errorf("budget not in SENT")
		}

		if err := tx.Model(&budgetRow{}).
			Where("id = ? AND deleted_at is null", b.ID).
			Updates(map[string]any{
				"status":           string(order.BudgetStatusApproved),
				"decided_at":       now,
				"approved_at":      now,
				"approved_by_name": approvedByName,
				"updated_at":       now,
			}).Error; err != nil {
			return err
		}

		var bs []budgetServiceRow
		if err := tx.Where("budget_id = ? AND deleted_at is null", b.ID).Find(&bs).Error; err != nil {
			return err
		}
		for _, it := range bs {
			row := serviceOrderServiceRow{
				ID:              uuid.NewString(),
				ServiceOrderID:  soID,
				ServiceID:       it.ServiceID,
				BudgetServiceID: PtrString(it.ID),
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
				Status:          "APPROVED",
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}

		var bp []budgetPartRow
		if err := tx.Where("budget_id = ? AND deleted_at is null", b.ID).Find(&bp).Error; err != nil {
			return err
		}

		refType := "SERVICE_ORDER"
		for _, it := range bp {
			if it.PartID == nil {
				continue
			}

			var p partRowForUpdate
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ? AND deleted_at is null", *it.PartID).
				First(&p).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return repository.ErrNotFound
				}
				return err
			}
			if p.StockQuantity < it.Quantity {
				return fmt.Errorf("insufficient stock for part %s", *it.PartID)
			}

			newQty := p.StockQuantity - it.Quantity
			if err := tx.Model(&partRowForUpdate{}).
				Where("id = ? AND deleted_at is null", p.ID).
				Updates(map[string]any{"stock_quantity": newQty, "updated_at": now}).Error; err != nil {
				return err
			}

			refID := soID
			note := "Consumed by service order approval"
			if err := tx.Create(&stockMovementRowForFlow{
				ID:              uuid.NewString(),
				PartID:          p.ID,
				MovementType:    "OUT",
				Quantity:        it.Quantity,
				ReferenceType:   &refType,
				ReferenceID:     &refID,
				Notes:           &note,
				CreatedByUserID: nil,
				CreatedAt:       now,
			}).Error; err != nil {
				return err
			}

			row := serviceOrderPartRow{
				ID:              uuid.NewString(),
				ServiceOrderID:  soID,
				PartID:          it.PartID,
				BudgetPartID:    PtrString(it.ID),
				Description:     it.Description,
				Quantity:        it.Quantity,
				UnitPriceCents:  it.UnitPriceCents,
				TotalPriceCents: it.TotalPriceCents,
				Status:          "APPROVED",
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}

		return r.transitionStatusTx(tx, soID, order.StatusInProgress, nil, nil, map[string]any{
			"execution_started_at": now,
		})
	})
}

func (r *ServiceOrderFlowRepository) RejectLatestBudgetByCode(ctx context.Context, code string, documentNumber string, rejectionReason string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		soID, err := serviceOrderIDByCodeAndDocumentForUpdate(tx, code, documentNumber)
		if err != nil {
			return err
		}

		b, err := latestBudgetForUpdate(tx, soID)
		if err != nil {
			return err
		}
		if b.Status != string(order.BudgetStatusSent) {
			return fmt.Errorf("budget not in SENT")
		}

		if err := tx.Model(&budgetRow{}).
			Where("id = ? AND deleted_at is null", b.ID).
			Updates(map[string]any{
				"status":           string(order.BudgetStatusRejected),
				"decided_at":       now,
				"rejected_at":      now,
				"rejection_reason": rejectionReason,
				"updated_at":       now,
			}).Error; err != nil {
			return err
		}

		reason := rejectionReason
		return r.transitionStatusTx(tx, soID, order.StatusInDiagnosis, nil, &reason, map[string]any{
			"diagnosis_started_at": now,
		})
	})
}

func (r *ServiceOrderFlowRepository) transitionStatus(ctx context.Context, serviceOrderID string, to order.Status, changedByUserID *string, reason *string, updates map[string]any) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.transitionStatusTx(tx, serviceOrderID, to, changedByUserID, reason, updates)
	})
}

func (r *ServiceOrderFlowRepository) transitionStatusTx(tx *gorm.DB, serviceOrderID string, to order.Status, changedByUserID *string, reason *string, updates map[string]any) error {
	var current struct {
		Status string
	}
	if err := tx.Table("service_orders").
		Select("status").
		Where("id = ? AND deleted_at is null", serviceOrderID).
		Scan(&current).Error; err != nil {
		return err
	}
	if current.Status == "" {
		return repository.ErrNotFound
	}
	if current.Status == string(to) {
		return nil
	}
	if !IsAllowedTransition(order.Status(current.Status), to) {
		return fmt.Errorf("invalid status transition %s -> %s", current.Status, to)
	}

	now := time.Now().UTC()
	updates = cloneUpdates(updates)
	updates["status"] = string(to)
	updates["updated_at"] = now
	if err := tx.Table("service_orders").
		Where("id = ? AND deleted_at is null", serviceOrderID).
		Updates(updates).Error; err != nil {
		return err
	}

	from := current.Status
	h := serviceOrderStatusHistoryRow{
		ID:              uuid.NewString(),
		ServiceOrderID:  serviceOrderID,
		FromStatus:      &from,
		ToStatus:        string(to),
		ChangedByUserID: changedByUserID,
		Reason:          reason,
		ChangedAt:       now,
		CreatedAt:       now,
	}
	return tx.Create(&h).Error
}

func IsAllowedTransition(from, to order.Status) bool {
	switch from {
	case order.StatusReceived:
		return to == order.StatusInDiagnosis || to == order.StatusWaitingApproval
	case order.StatusInDiagnosis:
		return to == order.StatusWaitingApproval
	case order.StatusWaitingApproval:
		return to == order.StatusInProgress || to == order.StatusInDiagnosis
	case order.StatusInProgress:
		return to == order.StatusFinished
	case order.StatusFinished:
		return to == order.StatusDelivered
	default:
		return false
	}
}

func cloneUpdates(m map[string]any) map[string]any {
	out := make(map[string]any, len(m)+2)
	for k, v := range m {
		out[k] = v
	}
	return out
}

func latestBudgetForUpdate(tx *gorm.DB, serviceOrderID string) (*budgetRow, error) {
	var b budgetRow
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("service_order_id = ? AND deleted_at is null", serviceOrderID).
		Order("version desc").
		First(&b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func serviceOrderIDByCodeAndDocumentForUpdate(tx *gorm.DB, code string, documentNumber string) (string, error) {
	type row struct {
		ID string
	}
	var r0 row
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Table("service_orders so").
		Select("so.id").
		Joins("join clients c on c.id = so.client_id and c.deleted_at is null").
		Where("so.deleted_at is null").
		Where("so.code = ?", code).
		Where("c.document_number = ?", documentNumber).
		Limit(1).
		Scan(&r0).Error
	if err != nil {
		return "", err
	}
	if r0.ID == "" {
		return "", repository.ErrNotFound
	}
	return r0.ID, nil
}

func PtrString(v string) *string { return &v }
