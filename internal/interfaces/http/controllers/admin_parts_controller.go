package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/mapper"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type AdminPartsController struct {
	svc *admin.PartService
}

func NewAdminPartsController(svc *admin.PartService) *AdminPartsController {
	return &AdminPartsController{svc: svc}
}

// @Summary List parts and supplies
// @Tags admin-parts
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.PartResponse
// @Router /admin/parts [get]
func (h *AdminPartsController) List(c *gin.Context) {
	limit, offset := shared.ParseLimitOffset(c)
	parts, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	out := make([]dto.PartResponse, 0, len(parts))
	for _, p := range parts {
		out = append(out, mapper.PartResponseFromDomain(p))
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Create part
// @Tags admin-parts
// @Param request body dto.CreatePartRequest true "Part"
// @Success 201 {object} dto.PartResponse
// @Router /admin/parts [post]
func (h *AdminPartsController) Create(c *gin.Context) {
	var req dto.CreatePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	p, err := h.svc.CreateFromInput(c.Request.Context(), admin.CreatePartInput{
		SKU:            req.SKU,
		Name:           req.Name,
		Description:    req.Description,
		UnitPriceCents: req.UnitPriceCents,
		StockQuantity:  req.StockQuantity,
		Active:         true,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, mapper.PartResponseFromDomain(*p))
}

// @Summary Get part
// @Tags admin-parts
// @Param id path string true "Part ID"
// @Success 200 {object} dto.PartResponse
// @Router /admin/parts/{id} [get]
func (h *AdminPartsController) Get(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapper.PartResponseFromDomain(*p))
}

// @Summary Update part
// @Tags admin-parts
// @Param id path string true "Part ID"
// @Param request body dto.CreatePartRequest true "Part"
// @Success 200 {object} dto.PartResponse
// @Router /admin/parts/{id} [put]
func (h *AdminPartsController) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.CreatePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	p, err := h.svc.UpdateFromInput(c.Request.Context(), id, admin.UpdatePartInput{
		SKU:            req.SKU,
		Name:           req.Name,
		Description:    req.Description,
		UnitPriceCents: req.UnitPriceCents,
		Active:         req.Active,
	})
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapper.PartResponseFromDomain(*p))
}

// @Summary Delete part
// @Tags admin-parts
// @Param id path string true "Part ID"
// @Success 204
// @Router /admin/parts/{id} [delete]
func (h *AdminPartsController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Adjust part stock (creates stock movement)
// @Tags admin-parts
// @Param id path string true "Part ID"
// @Param request body dto.AdjustStockRequest true "Stock movement"
// @Success 200 {object} dto.PartResponse
// @Router /admin/parts/{id}/stock-movements [post]
func (h *AdminPartsController) AdjustStock(c *gin.Context) {
	id := c.Param("id")
	var req dto.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	claims, _ := middlewares.GetClaims(c)
	var createdBy *string
	if claims != nil && claims.Subject != "" {
		createdBy = &claims.Subject
	}

	p, err := h.svc.AdjustStock(c.Request.Context(), id, part.StockMovementType(req.MovementType), req.Quantity, req.Notes, createdBy)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, repository.ErrConflict) {
			shared.WriteRepoError(c, err)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapper.PartResponseFromDomain(*p))
}
