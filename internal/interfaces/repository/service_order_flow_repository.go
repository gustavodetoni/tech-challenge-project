package repository

import (
	"context"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
)

type CreateServiceOrderDraftParams struct {
	ServiceOrder   order.ServiceOrder
	Budget         order.Budget
	BudgetServices []order.BudgetServiceItem
	BudgetParts    []order.BudgetPartItem
}

type ClientServiceOrderView struct {
	Code             string
	Status           order.Status
	ClientID         string
	VehicleID        string
	OpenedAt         string
	BudgetStatus     order.BudgetStatus
	BudgetTotalCents int64
}

type ServiceOrderFlowRepository interface {
	CreateDraft(ctx context.Context, p CreateServiceOrderDraftParams) (*order.ServiceOrder, *order.Budget, error)
	StartDiagnosis(ctx context.Context, serviceOrderID string, changedByUserID *string) error
	SendLatestBudget(ctx context.Context, serviceOrderID string, changedByUserID *string) error
	Finish(ctx context.Context, serviceOrderID string, changedByUserID *string) error
	Deliver(ctx context.Context, serviceOrderID string, changedByUserID *string) error

	GetClientViewByCode(ctx context.Context, code string, documentNumber string) (*ClientServiceOrderView, error)
	ApproveLatestBudgetByCode(ctx context.Context, code string, documentNumber string, approvedByName *string) error
	RejectLatestBudgetByCode(ctx context.Context, code string, documentNumber string, rejectionReason string) error
}
