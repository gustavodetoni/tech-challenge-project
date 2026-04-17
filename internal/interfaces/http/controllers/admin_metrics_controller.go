package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type AdminMetricsController struct {
	svc *admin.ServiceOrderService
}

func NewAdminMetricsController(svc *admin.ServiceOrderService) *AdminMetricsController {
	return &AdminMetricsController{svc: svc}
}

type avgExecutionResponse struct {
	AverageMinutes float64 `json:"average_minutes"`
}

type avgServiceExecutionItem struct {
	ServiceID      *string `json:"service_id,omitempty"`
	Description    string  `json:"description"`
	AverageMinutes float64 `json:"average_minutes"`
	SampleCount    int64   `json:"sample_count"`
}

type avgServiceExecutionResponse struct {
	Items []avgServiceExecutionItem `json:"items"`
}

// @Summary Average execution time (service orders)
// @Tags admin-metrics
// @Param from query string false "From (RFC3339)"
// @Param to query string false "To (RFC3339)"
// @Success 200 {object} avgExecutionResponse
// @Router /admin/metrics/avg-execution-time [get]
func (h *AdminMetricsController) AverageExecutionTime(c *gin.Context) {
	fromPtr, toPtr, ok := parseFromTo(c)
	if !ok {
		return
	}

	avg, err := h.svc.AverageExecutionMinutes(c.Request.Context(), fromPtr, toPtr)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, avgExecutionResponse{AverageMinutes: avg})
}

// @Summary Average execution time by service
// @Tags admin-metrics
// @Param service_id query string false "Service ID"
// @Param from query string false "From (RFC3339)"
// @Param to query string false "To (RFC3339)"
// @Success 200 {object} avgServiceExecutionResponse
// @Router /admin/metrics/avg-service-execution-time [get]
func (h *AdminMetricsController) AverageServiceExecutionTime(c *gin.Context) {
	fromPtr, toPtr, ok := parseFromTo(c)
	if !ok {
		return
	}

	var serviceIDPtr *string
	if v := c.Query("service_id"); v != "" {
		serviceIDPtr = &v
	}

	rows, err := h.svc.AverageServiceExecutionMinutes(c.Request.Context(), serviceIDPtr, fromPtr, toPtr)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	items := make([]avgServiceExecutionItem, 0, len(rows))
	for _, it := range rows {
		items = append(items, avgServiceExecutionItem{
			ServiceID:      it.ServiceID,
			Description:    it.Description,
			AverageMinutes: it.AverageMinutes,
			SampleCount:    it.SampleCount,
		})
	}
	c.JSON(http.StatusOK, avgServiceExecutionResponse{Items: items})
}

func parseFromTo(c *gin.Context) (*time.Time, *time.Time, bool) {
	var fromPtr *time.Time
	var toPtr *time.Time

	if v := c.Query("from"); v != "" {
		tm, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
			return nil, nil, false
		}
		fromPtr = &tm
	}
	if v := c.Query("to"); v != "" {
		tm, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
			return nil, nil, false
		}
		toPtr = &tm
	}
	return fromPtr, toPtr, true
}
