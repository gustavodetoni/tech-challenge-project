package dto

type CreateServiceRequest struct {
	Name             string  `json:"name" binding:"required"`
	Description      *string `json:"description"`
	BasePriceCents   int64   `json:"base_price_cents"`
	EstimatedMinutes int     `json:"estimated_minutes"`
	Active           bool    `json:"active"`
}

type ServiceResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	BasePriceCents   int64   `json:"base_price_cents"`
	EstimatedMinutes int     `json:"estimated_minutes"`
	Active           bool    `json:"active"`
}
