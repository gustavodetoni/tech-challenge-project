package order

import "time"

type BudgetStatus string

const (
	BudgetStatusDraft    BudgetStatus = "DRAFT"
	BudgetStatusSent     BudgetStatus = "SENT"
	BudgetStatusApproved BudgetStatus = "APPROVED"
	BudgetStatusRejected BudgetStatus = "REJECTED"
)

type Budget struct {
	ID               string
	ServiceOrderID   string
	Version          int
	Status           BudgetStatus
	TotalAmountCents int64
	SentAt           *time.Time
	ApprovedAt       *time.Time
	RejectedAt       *time.Time
	ApprovedByName   *string
	RejectionReason  *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type BudgetServiceItem struct {
	ServiceID       string
	Description     string
	Quantity        int
	UnitPriceCents  int64
	TotalPriceCents int64
}

type BudgetPartItem struct {
	PartID          string
	Description     string
	Quantity        int
	UnitPriceCents  int64
	TotalPriceCents int64
}
