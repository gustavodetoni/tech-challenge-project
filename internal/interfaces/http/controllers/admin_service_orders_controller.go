package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
)

type AdminServiceOrdersController struct {
	svc *admin.ServiceOrderService
}

func NewAdminServiceOrdersController(svc *admin.ServiceOrderService) *AdminServiceOrdersController {
	return &AdminServiceOrdersController{svc: svc}
}

type serviceOrderSummaryResponse struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Status     string `json:"status"`
	OpenedAt   string `json:"opened_at"`
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	VehicleID  string `json:"vehicle_id"`
	Plate      string `json:"plate"`
}

type serviceOrderDetailResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Status    string `json:"status"`
	ClientID  string `json:"client_id"`
	VehicleID string `json:"vehicle_id"`
	OpenedAt  string `json:"opened_at"`
}

// AdminListServiceOrders godoc
// @Summary List service orders
// @Tags admin-service-orders
// @Security BearerAuth
// @Param status query string false "Status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} serviceOrderSummaryResponse
// @Router /admin/service-orders [get]
func (h *AdminServiceOrdersController) List(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	var status *order.Status
	if v := c.Query("status"); v != "" {
		s := order.Status(v)
		status = &s
	}

	items, err := h.svc.List(c.Request.Context(), limit, offset, status)
	if err != nil {
		writeRepoError(c, err)
		return
	}

	out := make([]serviceOrderSummaryResponse, 0, len(items))
	for _, it := range items {
		out = append(out, serviceOrderSummaryResponse{
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

// AdminGetServiceOrder godoc
// @Summary Get service order
// @Tags admin-service-orders
// @Security BearerAuth
// @Param id path string true "Service Order ID"
// @Success 200 {object} serviceOrderDetailResponse
// @Failure 404 {object} map[string]string
// @Router /admin/service-orders/{id} [get]
func (h *AdminServiceOrdersController) Get(c *gin.Context) {
	id := c.Param("id")
	it, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, serviceOrderDetailResponse{
		ID:        it.ID,
		Code:      it.Code,
		Status:    string(it.Status),
		ClientID:  it.ClientID,
		VehicleID: it.VehicleID,
		OpenedAt:  it.OpenedAt.Format(time.RFC3339),
	})
}
