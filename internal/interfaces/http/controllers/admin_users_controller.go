package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

type AdminUsersController struct {
	svc *admin.UserService
}

func NewAdminUsersController(svc *admin.UserService) *AdminUsersController {
	return &AdminUsersController{svc: svc}
}

// @Summary Update user role
// @Tags admin-users
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRoleRequest true "Role update"
// @Success 204
// @Router /admin/users/{id}/role [put]
func (h *AdminUsersController) UpdateRole(c *gin.Context) {
	claims, ok := middlewares.GetClaims(c)
	if !ok || claims == nil || claims.Role != string(user.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	id := c.Param("id")
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.svc.UpdateRole(c.Request.Context(), id, user.Role(req.Role)); err != nil {
		if errors.Is(err, admin.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		shared.WriteRepoError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
