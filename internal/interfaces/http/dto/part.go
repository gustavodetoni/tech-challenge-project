package dto

type CreatePartRequest struct {
	SKU            string  `json:"sku" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Description    *string `json:"description"`
	UnitPriceCents int64   `json:"unit_price_cents"`
	StockQuantity  int     `json:"stock_quantity"`
	Active         bool    `json:"active"`
}

type PartResponse struct {
	ID             string  `json:"id"`
	SKU            string  `json:"sku"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	UnitPriceCents int64   `json:"unit_price_cents"`
	StockQuantity  int     `json:"stock_quantity"`
	Active         bool    `json:"active"`
}

type AdjustStockRequest struct {
	MovementType string  `json:"movement_type" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required"`
	Notes        *string `json:"notes"`
}
