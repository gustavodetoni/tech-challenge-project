package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
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

// AdminAverageExecutionTime godoc
// @Summary Average execution time (service orders)
// @Tags admin-metrics
// @Security BearerAuth
// @Param from query string false "From (RFC3339)"
// @Param to query string false "To (RFC3339)"
// @Success 200 {object} avgExecutionResponse
// @Router /admin/metrics/avg-execution-time [get]
func (h *AdminMetricsController) AverageExecutionTime(c *gin.Context) {
	var fromPtr *time.Time
	var toPtr *time.Time

	if v := c.Query("from"); v != "" {
		tm, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
			return
		}
		fromPtr = &tm
	}
	if v := c.Query("to"); v != "" {
		tm, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
			return
		}
		toPtr = &tm
	}

	avg, err := h.svc.AverageExecutionMinutes(c.Request.Context(), fromPtr, toPtr)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, avgExecutionResponse{AverageMinutes: avg})
}

