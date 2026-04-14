package dto

import (
	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
)

type CreateServiceOrderRequest struct {
	ClientDocumentType   string  `json:"client_document_type" binding:"required"`
	ClientDocumentNumber string  `json:"client_document_number" binding:"required"`
	ClientName           string  `json:"client_name" binding:"required"`
	ClientEmail          *string `json:"client_email"`
	ClientPhone          *string `json:"client_phone"`

	VehiclePlate           string  `json:"vehicle_plate" binding:"required"`
	VehicleBrand           string  `json:"vehicle_brand" binding:"required"`
	VehicleModel           string  `json:"vehicle_model" binding:"required"`
	VehicleManufactureYear *int    `json:"vehicle_manufacture_year"`
	VehicleModelYear       int     `json:"vehicle_model_year" binding:"required"`
	VehicleColor           *string `json:"vehicle_color"`

	CustomerComplaint *string `json:"customer_complaint"`

	Services []serviceorder.ItemInput `json:"services"`
	Parts    []serviceorder.ItemInput `json:"parts"`
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
}

type ClientServiceOrderResponse struct {
	Code             string `json:"code"`
	Status           string `json:"status"`
	OpenedAt         string `json:"opened_at"`
	BudgetStatus     string `json:"budget_status"`
	BudgetTotalCents int64  `json:"budget_total_cents"`
}