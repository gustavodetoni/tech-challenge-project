package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type AdminServiceOrdersController struct {
	svc *admin.ServiceOrderService
}

func NewAdminServiceOrdersController(svc *admin.ServiceOrderService) *AdminServiceOrdersController {
	return &AdminServiceOrdersController{svc: svc}
}

// @Summary List service orders
// @Tags admin-service-orders
// @Param status query string false "Status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} serviceOrderSummaryResponse
// @Router /admin/service-orders [get]
func (h *AdminServiceOrdersController) List(c *gin.Context) {
	limit, offset := shared.ParseLimitOffset(c)
	var status *order.Status
	if v := c.Query("status"); v != "" {
		s := order.Status(v)
		status = &s
	}

	items, err := h.svc.List(c.Request.Context(), limit, offset, status)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}

	out := make([]dto.ServiceOrderSummaryResponse, 0, len(items))
	for _, it := range items {
		out = append(out, dto.ServiceOrderSummaryResponse{
			ID:         it.ID,
			Code:       it.Code,
			Status:     string(it.Status),
			OpenedAt:   it.OpenedAt.Format(time.RFC3339),
			ClientID:   it.ClientID,
			ClientName: it.ClientName,
			VehicleID:  it.VehicleID,
			Plate:      it.Plate,
		})
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Get service order
// @Tags admin-service-orders
// @Param id path string true "Service Order ID"
// @Success 200 {object} serviceOrderDetailResponse
// @Router /admin/service-orders/{id} [get]
func (h *AdminServiceOrdersController) Get(c *gin.Context) {
	id := c.Param("id")
	it, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ServiceOrderDetailResponse{
		ID:        it.ID,
		Code:      it.Code,
		Status:    string(it.Status),
		ClientID:  it.ClientID,
		VehicleID: it.VehicleID,
		OpenedAt:  it.OpenedAt.Format(time.RFC3339),
	})
}
