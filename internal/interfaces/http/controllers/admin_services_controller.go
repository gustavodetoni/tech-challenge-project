package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/mapper"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type AdminServicesController struct {
	svc *admin.ServiceCatalogUseCase
}

func NewAdminServicesController(svc *admin.ServiceCatalogUseCase) *AdminServicesController {
	return &AdminServicesController{svc: svc}
}

// @Summary List services
// @Tags admin-services
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.ServiceResponse
// @Router /admin/services [get]
func (h *AdminServicesController) List(c *gin.Context) {
	limit, offset := shared.ParseLimitOffset(c)
	services, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	out := make([]dto.ServiceResponse, 0, len(services))
	for _, s := range services {
		out = append(out, mapper.ServiceResponseFromDomain(s))
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Create service
// @Tags admin-services
// @Param request body dto.CreateServiceRequest true "Service"
// @Success 201 {object} dto.ServiceResponse
// @Router /admin/services [post]
func (h *AdminServicesController) Create(c *gin.Context) {
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	svc, err := h.svc.CreateFromInput(c.Request.Context(), admin.CreateServiceInput{
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
	c.JSON(http.StatusCreated, mapper.ServiceResponseFromDomain(*svc))
}

// @Summary Get service
// @Tags admin-services
// @Param id path string true "Service ID"
// @Success 200 {object} dto.ServiceResponse
// @Router /admin/services/{id} [get]
func (h *AdminServicesController) Get(c *gin.Context) {
	id := c.Param("id")
	svc, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapper.ServiceResponseFromDomain(*svc))
}

// @Summary Update service
// @Tags admin-services
// @Param id path string true "Service ID"
// @Param request body dto.CreateServiceRequest true "Service"
// @Success 200 {object} dto.ServiceResponse
// @Router /admin/services/{id} [put]
func (h *AdminServicesController) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	svc, err := h.svc.UpdateFromInput(c.Request.Context(), id, admin.UpdateServiceInput{
		Name:             req.Name,
		Description:      req.Description,
		BasePriceCents:   req.BasePriceCents,
		EstimatedMinutes: req.EstimatedMinutes,
		Active:           req.Active,
	})
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapper.ServiceResponseFromDomain(*svc))
}

// @Summary Delete service
// @Tags admin-services
// @Param id path string true "Service ID"
// @Success 204
// @Router /admin/services/{id} [delete]
func (h *AdminServicesController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
