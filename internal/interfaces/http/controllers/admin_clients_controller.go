package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/mapper"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type AdminClientsController struct {
	svc *admin.ClientService
}

func NewAdminClientsController(svc *admin.ClientService) *AdminClientsController {
	return &AdminClientsController{svc: svc}
}

// @Summary List clients
// @Tags admin-clients
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.ClientResponse
// @Router /admin/clients [get]
func (h *AdminClientsController) List(c *gin.Context) {
	limit, offset := shared.ParseLimitOffset(c)
	clients, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}

	out := make([]dto.ClientResponse, 0, len(clients))
	for _, cl := range clients {
		out = append(out, mapper.ClientResponseFromDomain(cl))
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Create client
// @Tags admin-clients
// @Param request body dto.CreateClientRequest true "Client"
// @Success 201 {object} dto.ClientResponse
// @Router /admin/clients [post]
func (h *AdminClientsController) Create(c *gin.Context) {
	var req dto.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	cl, err := h.svc.Create(c.Request.Context(), client.Client{
		DocumentType:   client.DocumentType(req.DocumentType),
		DocumentNumber: req.DocumentNumber,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapper.ClientResponseFromDomain(*cl))
}

// @Summary Get client
// @Tags admin-clients
// @Param id path string true "Client ID"
// @Success 200 {object} dto.ClientResponse
// @Router /admin/clients/{id} [get]
func (h *AdminClientsController) Get(c *gin.Context) {
	id := c.Param("id")
	cl, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapper.ClientResponseFromDomain(*cl))
}

// @Summary Update client
// @Tags admin-clients
// @Param id path string true "Client ID"
// @Param request body dto.CreateClientRequest true "Client"
// @Success 200 {object} dto.ClientResponse
// @Router /admin/clients/{id} [put]
func (h *AdminClientsController) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	cl, err := h.svc.Update(c.Request.Context(), id, client.Client{
		DocumentType:   client.DocumentType(req.DocumentType),
		DocumentNumber: req.DocumentNumber,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
	})
	if err != nil {
		shared.WriteRepoError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.ClientResponseFromDomain(*cl))
}

// @Summary Delete client
// @Tags admin-clients
// @Param id path string true "Client ID"
// @Success 204
// @Router /admin/clients/{id} [delete]
func (h *AdminClientsController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		shared.WriteRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
