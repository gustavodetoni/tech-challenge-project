package dto

type ServiceOrderItemRequest struct {
	ID       string `json:"id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

type CreateServiceOrderRequest struct {
	ClientDocumentType   string  `json:"client_document_type" binding:"required"`
	ClientDocumentNumber string  `json:"client_document_number" binding:"required"`
	ClientEmail          *string `json:"client_email"`
	ClientPhone          *string `json:"client_phone"`

	VehiclePlate           string  `json:"vehicle_plate" binding:"required"`
	VehicleManufactureYear *int    `json:"vehicle_manufacture_year"`
	VehicleColor           *string `json:"vehicle_color"`

	CustomerComplaint *string `json:"customer_complaint"`

	Services []ServiceOrderItemRequest `json:"services"`
	Parts    []ServiceOrderItemRequest `json:"parts"`
}

type CreateServiceOrderResponse struct {
	ServiceOrderID string `json:"service_order_id"`
	Code           string `json:"code"`
	BudgetID       string `json:"budget_id"`
	BudgetStatus   string `json:"budget_status"`
	TotalCents     int64  `json:"total_cents"`
}

type ServiceOrderSummaryResponse struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Status     string `json:"status"`
	OpenedAt   string `json:"opened_at"`
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	VehicleID  string `json:"vehicle_id"`
	Plate      string `json:"plate"`
}

type ServiceOrderDetailResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Status    string `json:"status"`
	ClientID  string `json:"client_id"`
	VehicleID string `json:"vehicle_id"`
	OpenedAt  string `json:"opened_at"`

	LatestBudget   *ServiceOrderBudgetResponse `json:"latest_budget,omitempty"`
	BudgetServices []ServiceOrderLineResponse  `json:"budget_services,omitempty"`
	BudgetParts    []ServiceOrderLineResponse  `json:"budget_parts,omitempty"`
	StatusHistory  []ServiceOrderStatusHistory `json:"status_history,omitempty"`
}

type ServiceOrderBudgetResponse struct {
	ID         string  `json:"id"`
	Version    int     `json:"version"`
	Status     string  `json:"status"`
	TotalCents int64   `json:"total_cents"`
	SentAt     *string `json:"sent_at,omitempty"`
	ApprovedAt *string `json:"approved_at,omitempty"`
	RejectedAt *string `json:"rejected_at,omitempty"`

	ApprovedByName  *string `json:"approved_by_name,omitempty"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

type ServiceOrderLineResponse struct {
	RefID           *string `json:"ref_id,omitempty"`
	Description     string  `json:"description"`
	Quantity        int     `json:"quantity"`
	UnitPriceCents  int64   `json:"unit_price_cents"`
	TotalPriceCents int64   `json:"total_price_cents"`
}

type ServiceOrderStatusHistory struct {
	FromStatus      *string `json:"from_status,omitempty"`
	ToStatus        string  `json:"to_status"`
	ChangedAt       string  `json:"changed_at"`
	ChangedByUserID *string `json:"changed_by_user_id,omitempty"`
	Reason          *string `json:"reason,omitempty"`
}

type ClientServiceOrderResponse struct {
	Code              string  `json:"code"`
	Status            string  `json:"status"`
	OpenedAt          string  `json:"opened_at"`
	CustomerComplaint *string `json:"customer_complaint,omitempty"`

	Vehicle ClientServiceOrderVehicleResponse `json:"vehicle"`

	LatestBudget ServiceOrderBudgetResponse `json:"latest_budget"`

	BudgetServices []ServiceOrderLineResponse  `json:"budget_services"`
	BudgetParts    []ServiceOrderLineResponse  `json:"budget_parts"`
	StatusHistory  []ServiceOrderStatusHistory `json:"status_history"`

	BudgetStatus     string `json:"budget_status"`
	BudgetTotalCents int64  `json:"budget_total_cents"`
}

type ClientServiceOrderVehicleResponse struct {
	Plate           string  `json:"plate"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	ManufactureYear *int    `json:"manufacture_year,omitempty"`
	ModelYear       int     `json:"model_year"`
	Color           *string `json:"color,omitempty"`
}

type ClientServiceOrderStatusResponse struct {
	Code   string `json:"code"`
	Status string `json:"status"`
}

type ExternalBudgetDecisionRequest struct {
	DocumentNumber string  `json:"document_number" binding:"required"`
	Decision       string  `json:"decision" binding:"required"`
	Reason         *string `json:"reason"`
}

type ReviseBudgetRequest struct {
	Services []ServiceOrderItemRequest `json:"services"`
	Parts    []ServiceOrderItemRequest `json:"parts"`
}

type ReviseBudgetResponse struct {
	BudgetID     string `json:"budget_id"`
	BudgetStatus string `json:"budget_status"`
	Version      int    `json:"version"`
	TotalCents   int64  `json:"total_cents"`
}
