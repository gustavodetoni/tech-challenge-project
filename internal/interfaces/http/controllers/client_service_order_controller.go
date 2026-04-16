package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type ClientServiceOrderController struct {
	svc *serviceorder.Service
}

func NewClientServiceOrderController(svc *serviceorder.Service) *ClientServiceOrderController {
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

	c.JSON(http.StatusOK, dto.ClientServiceOrderResponse{
		Code:             view.Code,
		Status:           string(view.Status),
		OpenedAt:         view.OpenedAt,
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
