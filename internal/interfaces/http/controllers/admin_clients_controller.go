package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
)

type AdminClientsController struct {
	svc *admin.ClientService
}

func NewAdminClientsController(svc *admin.ClientService) *AdminClientsController {
	return &AdminClientsController{svc: svc}
}

type createClientRequest struct {
	DocumentType   string  `json:"document_type" binding:"required"`
	DocumentNumber string  `json:"document_number" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
}

type clientResponse struct {
	ID             string  `json:"id"`
	DocumentType   string  `json:"document_type"`
	DocumentNumber string  `json:"document_number"`
	Name           string  `json:"name"`
	Email          *string `json:"email,omitempty"`
	Phone          *string `json:"phone,omitempty"`
}

// AdminListClients godoc
// @Summary List clients
// @Tags admin-clients
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} clientResponse
// @Router /admin/clients [get]
func (h *AdminClientsController) List(c *gin.Context) {
	limit, offset := parseLimitOffset(c)
	clients, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		writeRepoError(c, err)
		return
	}

	out := make([]clientResponse, 0, len(clients))
	for _, cl := range clients {
		out = append(out, mapClientResponse(cl))
	}
	c.JSON(http.StatusOK, out)
}

// AdminCreateClient godoc
// @Summary Create client
// @Tags admin-clients
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createClientRequest true "Client"
// @Success 201 {object} clientResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /admin/clients [post]
func (h *AdminClientsController) Create(c *gin.Context) {
	var req createClientRequest
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

	c.JSON(http.StatusCreated, mapClientResponse(*cl))
}

// AdminGetClient godoc
// @Summary Get client
// @Tags admin-clients
// @Security BearerAuth
// @Param id path string true "Client ID"
// @Success 200 {object} clientResponse
// @Failure 404 {object} map[string]string
// @Router /admin/clients/{id} [get]
func (h *AdminClientsController) Get(c *gin.Context) {
	id := c.Param("id")
	cl, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		writeRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapClientResponse(*cl))
}

// AdminUpdateClient godoc
// @Summary Update client
// @Tags admin-clients
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body createClientRequest true "Client"
// @Success 200 {object} clientResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/clients/{id} [put]
func (h *AdminClientsController) Update(c *gin.Context) {
	id := c.Param("id")
	var req createClientRequest
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
		writeRepoError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapClientResponse(*cl))
}

// AdminDeleteClient godoc
// @Summary Delete client
// @Tags admin-clients
// @Security BearerAuth
// @Param id path string true "Client ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /admin/clients/{id} [delete]
func (h *AdminClientsController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func mapClientResponse(cl client.Client) clientResponse {
	return clientResponse{
		ID:             cl.ID,
		DocumentType:   string(cl.DocumentType),
		DocumentNumber: cl.DocumentNumber,
		Name:           cl.Name,
		Email:          cl.Email,
		Phone:          cl.Phone,
	}
}

