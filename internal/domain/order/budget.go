package order

import (
	"fmt"
	"time"
)

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

func (b *Budget) Send(now time.Time) error {
	if b == nil {
		return nil
	}
	if b.Status != BudgetStatusDraft {
		return fmt.Errorf("budget not in DRAFT")
	}
	b.Status = BudgetStatusSent
	b.SentAt = &now
	b.UpdatedAt = now
	return nil
}

func (b *Budget) Approve(now time.Time, approvedByName *string) error {
	if b == nil {
		return nil
	}
	if b.Status != BudgetStatusSent {
		return fmt.Errorf("budget not in SENT")
	}
	b.Status = BudgetStatusApproved
	b.ApprovedAt = &now
	b.ApprovedByName = approvedByName
	b.UpdatedAt = now
	return nil
}

func (b *Budget) Reject(now time.Time, reason string) error {
	if b == nil {
		return nil
	}
	if b.Status != BudgetStatusSent {
		return fmt.Errorf("budget not in SENT")
	}
	b.Status = BudgetStatusRejected
	b.RejectedAt = &now
	b.RejectionReason = &reason
	b.UpdatedAt = now
	return nil
}
