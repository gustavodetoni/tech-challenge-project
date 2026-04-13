package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type AdminPartsController struct {
	svc *admin.PartService
}

func NewAdminPartsController(svc *admin.PartService) *AdminPartsController { return &AdminPartsController{svc: svc} }

type createPartRequest struct {
	SKU            string  `json:"sku" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Description    *string `json:"description"`
	UnitPriceCents int64   `json:"unit_price_cents"`
	StockQuantity  int     `json:"stock_quantity"`
	Active         bool    `json:"active"`
}

type partResponse struct {
	ID            string  `json:"id"`
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	StockQuantity int     `json:"stock_quantity"`
	Active        bool    `json:"active"`
}

type adjustStockRequest struct {
	MovementType string  `json:"movement_type" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required"`
	Notes        *string `json:"notes"`
}

// AdminListParts godoc
// @Summary List parts and supplies
// @Tags admin-parts
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} partResponse
// @Router /admin/parts [get]
func (h *AdminPartsController) List(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	parts, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	out := make([]partResponse, 0, len(parts))
	for _, p := range parts {
		out = append(out, mapPartResponse(p))
	}
	c.JSON(http.StatusOK, out)
}

// AdminCreatePart godoc
// @Summary Create part
// @Tags admin-parts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createPartRequest true "Part"
// @Success 201 {object} partResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /admin/parts [post]
func (h *AdminPartsController) Create(c *gin.Context) {
	var req createPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	p, err := h.svc.Create(c.Request.Context(), part.Part{
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
	c.JSON(http.StatusCreated, mapPartResponse(*p))
}

// AdminGetPart godoc
// @Summary Get part
// @Tags admin-parts
// @Security BearerAuth
// @Param id path string true "Part ID"
// @Success 200 {object} partResponse
// @Failure 404 {object} map[string]string
// @Router /admin/parts/{id} [get]
func (h *AdminPartsController) Get(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapPartResponse(*p))
}

// AdminUpdatePart godoc
// @Summary Update part
// @Tags admin-parts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Part ID"
// @Param request body createPartRequest true "Part"
// @Success 200 {object} partResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/parts/{id} [put]
func (h *AdminPartsController) Update(c *gin.Context) {
	id := c.Param("id")
	var req createPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	p, err := h.svc.Update(c.Request.Context(), id, part.Part{
		SKU:            req.SKU,
		Name:           req.Name,
		Description:    req.Description,
		UnitPriceCents: req.UnitPriceCents,
		Active:         req.Active,
	})
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapPartResponse(*p))
}

// AdminDeletePart godoc
// @Summary Delete part
// @Tags admin-parts
// @Security BearerAuth
// @Param id path string true "Part ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /admin/parts/{id} [delete]
func (h *AdminPartsController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AdminAdjustStock godoc
// @Summary Adjust part stock (creates stock movement)
// @Tags admin-parts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Part ID"
// @Param request body adjustStockRequest true "Stock movement"
// @Success 200 {object} partResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/parts/{id}/stock-movements [post]
func (h *AdminPartsController) AdjustStock(c *gin.Context) {
	id := c.Param("id")
	var req adjustStockRequest
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
			writeRepoError(c, err)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapPartResponse(*p))
}

func mapPartResponse(p part.Part) partResponse {
	return partResponse{
		ID:             p.ID,
		SKU:            p.SKU,
		Name:           p.Name,
		Description:    p.Description,
		UnitPriceCents: p.UnitPriceCents,
		StockQuantity:  p.StockQuantity,
		Active:         p.Active,
	}
}
