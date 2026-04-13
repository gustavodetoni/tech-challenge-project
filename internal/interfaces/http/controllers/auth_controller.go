package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
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

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
}

type meResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"issued_at,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Register request"
// @Success 201 {object} authResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (a *AuthController) Register(c *gin.Context) {
	var req registerRequest
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

	c.JSON(http.StatusCreated, authResponse{
		AccessToken: out.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   out.ExpiresAt,
		UserID:      out.UserID,
		Role:        string(out.Role),
	})
}

// Login godoc
// @Summary Login with email/password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login request"
// @Success 200 {object} authResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (a *AuthController) Login(c *gin.Context) {
	var req loginRequest
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

	c.JSON(http.StatusOK, authResponse{
		AccessToken: out.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   out.ExpiresAt,
		UserID:      out.UserID,
		Role:        string(out.Role),
	})
}

// Me godoc
// @Summary Current identity
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} meResponse
// @Failure 401 {object} map[string]string
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

	resp := meResponse{
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
