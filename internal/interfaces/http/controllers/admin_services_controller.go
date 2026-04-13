package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
)

type AdminServicesController struct {
	svc *admin.ServiceService
}

func NewAdminServicesController(svc *admin.ServiceService) *AdminServicesController {
	return &AdminServicesController{svc: svc}
}

type createServiceRequest struct {
	Name            string  `json:"name" binding:"required"`
	Description     *string `json:"description"`
	BasePriceCents  int64   `json:"base_price_cents"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	Active          bool    `json:"active"`
}

type serviceResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	BasePriceCents  int64   `json:"base_price_cents"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	Active          bool    `json:"active"`
}

// AdminListServices godoc
// @Summary List services
// @Tags admin-services
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} serviceResponse
// @Router /admin/services [get]
func (h *AdminServicesController) List(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	services, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	out := make([]serviceResponse, 0, len(services))
	for _, s := range services {
		out = append(out, mapServiceResponse(s))
	}
	c.JSON(http.StatusOK, out)
}

// AdminCreateService godoc
// @Summary Create service
// @Tags admin-services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createServiceRequest true "Service"
// @Success 201 {object} serviceResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /admin/services [post]
func (h *AdminServicesController) Create(c *gin.Context) {
	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	svc, err := h.svc.Create(c.Request.Context(), service.Service{
		Name:             req.Name,
		Description:      req.Description,
		BasePriceCents:   req.BasePriceCents,
		EstimatedMinutes: req.EstimatedMinutes,
		Active:           true,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, mapServiceResponse(*svc))
}

// AdminGetService godoc
// @Summary Get service
// @Tags admin-services
// @Security BearerAuth
// @Param id path string true "Service ID"
// @Success 200 {object} serviceResponse
// @Failure 404 {object} map[string]string
// @Router /admin/services/{id} [get]
func (h *AdminServicesController) Get(c *gin.Context) {
	id := c.Param("id")
	svc, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapServiceResponse(*svc))
}

// AdminUpdateService godoc
// @Summary Update service
// @Tags admin-services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param request body createServiceRequest true "Service"
// @Success 200 {object} serviceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/services/{id} [put]
func (h *AdminServicesController) Update(c *gin.Context) {
	id := c.Param("id")
	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	svc, err := h.svc.Update(c.Request.Context(), id, service.Service{
		Name:             req.Name,
		Description:      req.Description,
		BasePriceCents:   req.BasePriceCents,
		EstimatedMinutes: req.EstimatedMinutes,
		Active:           req.Active,
	})
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapServiceResponse(*svc))
}

// AdminDeleteService godoc
// @Summary Delete service
// @Tags admin-services
// @Security BearerAuth
// @Param id path string true "Service ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /admin/services/{id} [delete]
func (h *AdminServicesController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func mapServiceResponse(s service.Service) serviceResponse {
	return serviceResponse{
		ID:               s.ID,
		Name:             s.Name,
		Description:      s.Description,
		BasePriceCents:   s.BasePriceCents,
		EstimatedMinutes: s.EstimatedMinutes,
		Active:           s.Active,
	}
}

