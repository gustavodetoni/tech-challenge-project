package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type ClientServiceOrderController struct {
	svc *serviceorder.FlowUseCase
}

func NewClientServiceOrderController(svc *serviceorder.FlowUseCase) *ClientServiceOrderController {
	return &ClientServiceOrderController{svc: svc}
}

// @Summary Get service order progress (client)
// @Tags client-service-orders
// @Param document_number query string true "CPF/CNPJ"
// @Param code path string true "Service Order Code"
// @Success 200 {object} dto.ClientServiceOrderResponse
// @Router /client/service-orders/{code} [get]
func (h *ClientServiceOrderController) Get(c *gin.Context) {
	code, doc, ok := getCodeAndDocumentNumber(c)
	if !ok {
		return
	}

	view, err := h.svc.ClientGetByCode(c.Request.Context(), code, doc)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}

	statusHistory := make([]dto.ServiceOrderStatusHistory, 0, len(view.StatusHistory))
	for _, it := range view.StatusHistory {
		var from *string
		if it.FromStatus != nil {
			s := string(*it.FromStatus)
			from = &s
		}
		statusHistory = append(statusHistory, dto.ServiceOrderStatusHistory{
			FromStatus:      from,
			ToStatus:        string(it.ToStatus),
			ChangedAt:       it.ChangedAt.UTC().Format(time.RFC3339),
			ChangedByUserID: it.ChangedByUserID,
			Reason:          it.Reason,
		})
	}

	budgetServices := make([]dto.ServiceOrderLineResponse, 0, len(view.BudgetServices))
	for _, it := range view.BudgetServices {
		var refID *string
		if it.ServiceID != "" {
			s := it.ServiceID
			refID = &s
		}
		budgetServices = append(budgetServices, dto.ServiceOrderLineResponse{
			RefID:           refID,
			Description:     it.Description,
			Quantity:        it.Quantity,
			UnitPriceCents:  it.UnitPriceCents,
			TotalPriceCents: it.TotalPriceCents,
		})
	}

	budgetParts := make([]dto.ServiceOrderLineResponse, 0, len(view.BudgetParts))
	for _, it := range view.BudgetParts {
		var refID *string
		if it.PartID != "" {
			s := it.PartID
			refID = &s
		}
		budgetParts = append(budgetParts, dto.ServiceOrderLineResponse{
			RefID:           refID,
			Description:     it.Description,
			Quantity:        it.Quantity,
			UnitPriceCents:  it.UnitPriceCents,
			TotalPriceCents: it.TotalPriceCents,
		})
	}

	latestBudget := dto.ServiceOrderBudgetResponse{
		ID:         view.BudgetID,
		Version:    view.BudgetVersion,
		Status:     string(view.BudgetStatus),
		TotalCents: view.BudgetTotalCents,
		SentAt:     view.BudgetSentAt,
		ApprovedAt: view.BudgetApprovedAt,
		RejectedAt: view.BudgetRejectedAt,

		ApprovedByName:  view.BudgetApprovedByName,
		RejectionReason: view.BudgetRejectionReason,
	}

	c.JSON(http.StatusOK, dto.ClientServiceOrderResponse{
		Code:              view.Code,
		Status:            string(view.Status),
		OpenedAt:          view.OpenedAt,
		CustomerComplaint: view.CustomerComplaint,
		Vehicle: dto.ClientServiceOrderVehicleResponse{
			Plate:           view.VehiclePlate,
			Brand:           view.VehicleBrand,
			Model:           view.VehicleModel,
			ManufactureYear: view.VehicleManufactureYear,
			ModelYear:       view.VehicleModelYear,
			Color:           view.VehicleColor,
		},
		LatestBudget:     latestBudget,
		BudgetServices:   budgetServices,
		BudgetParts:      budgetParts,
		StatusHistory:    statusHistory,
		BudgetStatus:     string(view.BudgetStatus),
		BudgetTotalCents: view.BudgetTotalCents,
	})
}

// @Summary Approve latest budget
// @Tags client-service-orders
// @Param document_number query string true "CPF/CNPJ"
// @Param code path string true "Service Order Code"
// @Success 204
// @Router /client/service-orders/{code}/budget/approve [post]
func (h *ClientServiceOrderController) ApproveBudget(c *gin.Context) {
	code, doc, ok := getCodeAndDocumentNumber(c)
	if !ok {
		return
	}

	if err := h.svc.ClientApproveBudget(c.Request.Context(), code, doc); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			shared.WriteRepoError(c, err)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type rejectBudgetRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// @Summary Reject latest budget
// @Tags client-service-orders
// @Param document_number query string true "CPF/CNPJ"
// @Param code path string true "Service Order Code"
// @Param request body rejectBudgetRequest true "Rejection reason"
// @Success 204
// @Router /client/service-orders/{code}/budget/reject [post]
func (h *ClientServiceOrderController) RejectBudget(c *gin.Context) {
	code, doc, ok := getCodeAndDocumentNumber(c)
	if !ok {
		return
	}

	var req rejectBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.svc.ClientRejectBudget(c.Request.Context(), code, doc, req.Reason); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			shared.WriteRepoError(c, err)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func getCodeAndDocumentNumber(c *gin.Context) (string, string, bool) {
	code := c.Param("code")
	doc := c.Query("document_number")

	if doc == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "document_number is required",
		})
		return "", "", false
	}

	return code, doc, true
}
