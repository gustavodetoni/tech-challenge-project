package order

import "time"

type StatusHistoryEntry struct {
	FromStatus      *Status
	ToStatus        Status
	ChangedAt       time.Time
	ChangedByUserID *string
	Reason          *string
}

type ServiceOrderDetail struct {
	ServiceOrder   ServiceOrder
	LatestBudget   *Budget
	BudgetServices []BudgetServiceItem
	BudgetParts    []BudgetPartItem
	StatusHistory  []StatusHistoryEntry
}
