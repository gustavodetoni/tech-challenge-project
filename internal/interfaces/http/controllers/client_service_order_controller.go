package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type ClientServiceOrderController struct {
	svc *serviceorder.Service
}

func NewClientServiceOrderController(svc *serviceorder.Service) *ClientServiceOrderController {
	return &ClientServiceOrderController{svc: svc}
}

type clientServiceOrderResponse struct {
	Code             string `json:"code"`
	Status           string `json:"status"`
	OpenedAt         string `json:"opened_at"`
	BudgetStatus     string `json:"budget_status"`
	BudgetTotalCents int64  `json:"budget_total_cents"`
}

// ClientGetServiceOrder godoc
// @Summary Get service order progress (client)
// @Tags client-service-orders
// @Param document_number query string true "CPF/CNPJ"
// @Param code path string true "Service Order Code"
// @Success 200 {object} clientServiceOrderResponse
// @Failure 404 {object} map[string]string
// @Router /client/service-orders/{code} [get]
func (h *ClientServiceOrderController) Get(c *gin.Context) {
	code := c.Param("code")
	doc := c.Query("document_number")
	if doc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document_number is required"})
		return
	}

	view, err := h.svc.ClientGetByCode(c.Request.Context(), code, doc)
	if err != nil {
		writeRepoError(c, err)
		return
	}

	c.JSON(http.StatusOK, clientServiceOrderResponse{
		Code:             view.Code,
		Status:           string(view.Status),
		OpenedAt:         view.OpenedAt,
		BudgetStatus:     string(view.BudgetStatus),
		BudgetTotalCents: view.BudgetTotalCents,
	})
}

type approveBudgetRequest struct {
	ApprovedByName *string `json:"approved_by_name"`
}

// ClientApproveBudget godoc
// @Summary Approve latest budget
// @Tags client-service-orders
// @Accept json
// @Produce json
// @Param document_number query string true "CPF/CNPJ"
// @Param code path string true "Service Order Code"
// @Param request body approveBudgetRequest false "Approval info"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /client/service-orders/{code}/budget/approve [post]
func (h *ClientServiceOrderController) ApproveBudget(c *gin.Context) {
	code := c.Param("code")
	doc := c.Query("document_number")
	if doc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document_number is required"})
		return
	}

	var req approveBudgetRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.ClientApproveBudget(c.Request.Context(), code, doc, req.ApprovedByName); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeRepoError(c, err)
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

// ClientRejectBudget godoc
// @Summary Reject latest budget
// @Tags client-service-orders
// @Accept json
// @Produce json
// @Param document_number query string true "CPF/CNPJ"
// @Param code path string true "Service Order Code"
// @Param request body rejectBudgetRequest true "Rejection reason"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /client/service-orders/{code}/budget/reject [post]
func (h *ClientServiceOrderController) RejectBudget(c *gin.Context) {
	code := c.Param("code")
	doc := c.Query("document_number")
	if doc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document_number is required"})
		return
	}

	var req rejectBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.svc.ClientRejectBudget(c.Request.Context(), code, doc, req.Reason); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeRepoError(c, err)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
