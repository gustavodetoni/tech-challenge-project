package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/dto"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
)

type AuthController struct {
	auth  *appAuth.Service
	users repository.UserRepository
}

func NewAuthController(auth *appAuth.Service, users repository.UserRepository) *AuthController {
	return &AuthController{auth: auth, users: users}
}

// @Summary Register a new user
// @Tags auth
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} dto.AuthResponse
// @Router /auth/register [post]
func (a *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	out, err := a.auth.Register(c.Request.Context(), appAuth.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, appAuth.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, appAuth.ErrEmailInUse):
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken: out.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   out.ExpiresAt,
		UserID:      out.UserID,
		Role:        string(out.Role),
	})
}

// @Summary Login with email/password
// @Tags auth
// @Success 200 {object} dto.AuthResponse
// @Router /auth/login [post]
func (a *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	out, err := a.auth.Login(c.Request.Context(), appAuth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, appAuth.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, appAuth.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken: out.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   out.ExpiresAt,
		UserID:      out.UserID,
		Role:        string(out.Role),
	})
}

// @Summary Current identity
// @Tags auth
// @Success 200 {object} dto.MeResponse
// @Router /me [get]
func (a *AuthController) Me(c *gin.Context) {
	claims, ok := middlewares.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	u, err := a.users.FindByID(c.Request.Context(), claims.Subject)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	resp := dto.MeResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  string(u.Role),
	}
	if claims.IssuedAt != nil {
		resp.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		resp.ExpiresAt = claims.ExpiresAt.Time
	}
	c.JSON(http.StatusOK, resp)
}
