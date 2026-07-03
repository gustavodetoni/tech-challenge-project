package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appVersion "github.com/soat-architecture/tech-challenge-project"
)

type HealthController struct{}

func NewHealthController() *HealthController { return &HealthController{} }

// @Summary Health check
// @Tags system
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": appVersion.Version,
	})
}
