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
	svc *admin.ServiceOrderAdminUseCase
}

func NewAdminServiceOrdersController(svc *admin.ServiceOrderAdminUseCase) *AdminServiceOrdersController {
	return &AdminServiceOrdersController{svc: svc}
}

// @Summary List service orders
// @Tags admin-service-orders
// @Param status query string false "Status"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.ServiceOrderSummaryResponse
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
		shared.WriteError(c, err)
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
// @Success 200 {object} dto.ServiceOrderDetailResponse
// @Router /admin/service-orders/{id} [get]
func (h *AdminServiceOrdersController) Get(c *gin.Context) {
	id := c.Param("id")
	detail, err := h.svc.GetDetailByID(c.Request.Context(), id)
	if err != nil {
		shared.WriteError(c, err)
		return
	}

	var latestBudget *dto.ServiceOrderBudgetResponse
	if detail.LatestBudget != nil {
		latestBudget = &dto.ServiceOrderBudgetResponse{
			ID:              detail.LatestBudget.ID,
			Version:         detail.LatestBudget.Version,
			Status:          string(detail.LatestBudget.Status),
			TotalCents:      detail.LatestBudget.TotalAmountCents,
			SentAt:          timePtrToRFC3339Ptr(detail.LatestBudget.SentAt),
			ApprovedAt:      timePtrToRFC3339Ptr(detail.LatestBudget.ApprovedAt),
			RejectedAt:      timePtrToRFC3339Ptr(detail.LatestBudget.RejectedAt),
			ApprovedByName:  detail.LatestBudget.ApprovedByName,
			RejectionReason: detail.LatestBudget.RejectionReason,
		}
	}

	bs := make([]dto.ServiceOrderLineResponse, 0, len(detail.BudgetServices))
	for _, it := range detail.BudgetServices {
		ref := it.ServiceID
		bs = append(bs, dto.ServiceOrderLineResponse{
			RefID:           &ref,
			Description:     it.Description,
			Quantity:        it.Quantity,
			UnitPriceCents:  it.UnitPriceCents,
			TotalPriceCents: it.TotalPriceCents,
		})
	}

	bp := make([]dto.ServiceOrderLineResponse, 0, len(detail.BudgetParts))
	for _, it := range detail.BudgetParts {
		ref := it.PartID
		bp = append(bp, dto.ServiceOrderLineResponse{
			RefID:           &ref,
			Description:     it.Description,
			Quantity:        it.Quantity,
			UnitPriceCents:  it.UnitPriceCents,
			TotalPriceCents: it.TotalPriceCents,
		})
	}

	hist := make([]dto.ServiceOrderStatusHistory, 0, len(detail.StatusHistory))
	for _, it := range detail.StatusHistory {
		var from *string
		if it.FromStatus != nil {
			s := string(*it.FromStatus)
			from = &s
		}
		hist = append(hist, dto.ServiceOrderStatusHistory{
			FromStatus:      from,
			ToStatus:        string(it.ToStatus),
			ChangedAt:       it.ChangedAt.Format(time.RFC3339),
			ChangedByUserID: it.ChangedByUserID,
			Reason:          it.Reason,
		})
	}

	c.JSON(http.StatusOK, dto.ServiceOrderDetailResponse{
		ID:             detail.ServiceOrder.ID,
		Code:           detail.ServiceOrder.Code,
		Status:         string(detail.ServiceOrder.Status),
		ClientID:       detail.ServiceOrder.ClientID,
		VehicleID:      detail.ServiceOrder.VehicleID,
		OpenedAt:       detail.ServiceOrder.OpenedAt.Format(time.RFC3339),
		LatestBudget:   latestBudget,
		BudgetServices: bs,
		BudgetParts:    bp,
		StatusHistory:  hist,
	})
}

func timePtrToRFC3339Ptr(v *time.Time) *string {
	if v == nil {
		return nil
	}
	s := v.UTC().Format(time.RFC3339)
	return &s
}
