package part

import "time"

type Part struct {
	ID             string
	SKU            string
	Name           string
	Description    *string
	UnitPriceCents int64
	StockQuantity  int
	Active         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type StockMovementType string

const (
	StockMovementIn         StockMovementType = "IN"
	StockMovementOut        StockMovementType = "OUT"
	StockMovementAdjustment StockMovementType = "ADJUSTMENT"
)
