package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type AdminServiceOrderFlowController struct {
	svc *serviceorder.Service
}

func NewAdminServiceOrderFlowController(svc *serviceorder.Service) *AdminServiceOrderFlowController {
	return &AdminServiceOrderFlowController{svc: svc}
}

// @Summary Create service order (draft budget)
// @Tags admin-service-orders
// @Param request body dto.CreateServiceOrderRequest true "Service order"
// @Success 201 {object} dto.CreateServiceOrderResponse
// @Router /admin/service-orders [post]
func (h *AdminServiceOrderFlowController) CreateDraft(c *gin.Context) {
	var req dto.CreateServiceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	services := make([]serviceorder.ItemInput, 0, len(req.Services))
	for _, it := range req.Services {
		services = append(services, serviceorder.ItemInput{ID: it.ID, Quantity: it.Quantity})
	}
	parts := make([]serviceorder.ItemInput, 0, len(req.Parts))
	for _, it := range req.Parts {
		parts = append(parts, serviceorder.ItemInput{ID: it.ID, Quantity: it.Quantity})
	}

	out, err := h.svc.CreateDraft(c.Request.Context(), serviceorder.CreateDraftInput{
		ClientDocumentType:     client.DocumentType(req.ClientDocumentType),
		ClientDocumentNumber:   req.ClientDocumentNumber,
		ClientName:             req.ClientName,
		ClientEmail:            req.ClientEmail,
		ClientPhone:            req.ClientPhone,
		VehiclePlate:           req.VehiclePlate,
		VehicleBrand:           req.VehicleBrand,
		VehicleModel:           req.VehicleModel,
		VehicleManufactureYear: req.VehicleManufactureYear,
		VehicleModelYear:       req.VehicleModelYear,
		VehicleColor:           req.VehicleColor,
		CustomerComplaint:      req.CustomerComplaint,
		Services:               services,
		Parts:                  parts,
	})
	if err != nil {
		switch {
		case errors.Is(err, serviceorder.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "service/part not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.CreateServiceOrderResponse{
		ServiceOrderID: out.ServiceOrderID,
		Code:           out.Code,
		BudgetID:       out.BudgetID,
		BudgetStatus:   string(out.BudgetStatus),
		TotalCents:     out.TotalCents,
	})
}

// @Summary Revise budget (new version, draft)
// @Tags admin-service-orders
// @Param id path string true "Service Order ID"
// @Param request body dto.ReviseBudgetRequest true "Budget revision"
// @Success 201 {object} dto.ReviseBudgetResponse
// @Router /admin/service-orders/{id}/budget/revise [post]
func (h *AdminServiceOrderFlowController) ReviseBudget(c *gin.Context) {
	id := c.Param("id")

	var req dto.ReviseBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	services := make([]serviceorder.ItemInput, 0, len(req.Services))
	for _, it := range req.Services {
		services = append(services, serviceorder.ItemInput{ID: it.ID, Quantity: it.Quantity})
	}
	parts := make([]serviceorder.ItemInput, 0, len(req.Parts))
	for _, it := range req.Parts {
		parts = append(parts, serviceorder.ItemInput{ID: it.ID, Quantity: it.Quantity})
	}

	claims, _ := middlewares.GetClaims(c)
	var userID *string
	if claims != nil && claims.Subject != "" {
		userID = &claims.Subject
	}

	out, err := h.svc.ReviseBudget(c.Request.Context(), id, serviceorder.ReviseBudgetInput{
		Services: services,
		Parts:    parts,
	}, userID)
	if err != nil {
		switch {
		case errors.Is(err, serviceorder.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrNotFound):
			shared.WriteRepoError(c, err)
		case errors.Is(err, repository.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ReviseBudgetResponse{
		BudgetID:     out.BudgetID,
		BudgetStatus: string(out.BudgetStatus),
		Version:      out.Version,
		TotalCents:   out.TotalCents,
	})
}

// @Summary Start diagnosis
// @Tags admin-service-orders
// @Param id path string true "Service Order ID"
// @Success 204
// @Router /admin/service-orders/{id}/diagnosis/start [post]
func (h *AdminServiceOrderFlowController) StartDiagnosis(c *gin.Context) {
	id := c.Param("id")
	claims, _ := middlewares.GetClaims(c)
	var userID *string
	if claims != nil && claims.Subject != "" {
		userID = &claims.Subject
	}
	if err := h.svc.StartDiagnosis(c.Request.Context(), id, userID); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Send budget for approval
// @Tags admin-service-orders
// @Param id path string true "Service Order ID"
// @Success 204
// @Router /admin/service-orders/{id}/budget/send [post]
func (h *AdminServiceOrderFlowController) SendBudget(c *gin.Context) {
	id := c.Param("id")
	claims, _ := middlewares.GetClaims(c)
	var userID *string
	if claims != nil && claims.Subject != "" {
		userID = &claims.Subject
	}
	if err := h.svc.SendBudget(c.Request.Context(), id, userID); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Finish service order
// @Tags admin-service-orders
// @Param id path string true "Service Order ID"
// @Success 204
// @Router /admin/service-orders/{id}/finish [post]
func (h *AdminServiceOrderFlowController) Finish(c *gin.Context) {
	id := c.Param("id")
	claims, _ := middlewares.GetClaims(c)
	var userID *string
	if claims != nil && claims.Subject != "" {
		userID = &claims.Subject
	}
	if err := h.svc.Finish(c.Request.Context(), id, userID); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Deliver service order
// @Tags admin-service-orders
// @Param id path string true "Service Order ID"
// @Success 204
// @Router /admin/service-orders/{id}/deliver [post]
func (h *AdminServiceOrderFlowController) Deliver(c *gin.Context) {
	id := c.Param("id")
	claims, _ := middlewares.GetClaims(c)
	var userID *string
	if claims != nil && claims.Subject != "" {
		userID = &claims.Subject
	}
	if err := h.svc.Deliver(c.Request.Context(), id, userID); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
