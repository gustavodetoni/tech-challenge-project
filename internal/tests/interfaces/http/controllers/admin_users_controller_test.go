package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	jwtAuth "github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAdminUsersController_UpdateRole_Forbidden_NoClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)
	h := controllers.NewAdminUsersController(svc)

	r := gin.New()
	r.PUT("/admin/users/:id/role", h.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/u1/role", bytes.NewBufferString(`{"role":"MANAGER"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminUsersController_UpdateRole_Forbidden_Manager(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)
	h := controllers.NewAdminUsersController(svc)

	r := gin.New()
	r.Use(withClaims("MANAGER"))
	r.PUT("/admin/users/:id/role", h.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/u1/role", bytes.NewBufferString(`{"role":"ADMIN"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminUsersController_UpdateRole_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)
	h := controllers.NewAdminUsersController(svc)

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.PUT("/admin/users/:id/role", h.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/u1/role", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUsersController_UpdateRole_BadRequest_InvalidRole(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)
	h := controllers.NewAdminUsersController(svc)

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.PUT("/admin/users/:id/role", h.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/u1/role", bytes.NewBufferString(`{"role":"NOPE"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUsersController_UpdateRole_NotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)
	h := controllers.NewAdminUsersController(svc)

	repo.On("UpdateRole", mock.Anything, "u1", user.RoleManager).Return(repository.ErrNotFound).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.PUT("/admin/users/:id/role", h.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/u1/role", bytes.NewBufferString(`{"role":"MANAGER"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

func TestAdminUsersController_UpdateRole_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	repo := new(repomocks.UserRepository)
	svc := admin.NewUserService(repo)
	h := controllers.NewAdminUsersController(svc)

	repo.On("UpdateRole", mock.Anything, "u1", user.RoleManager).Return(nil).Once()

	r := gin.New()
	r.Use(withClaims("ADMIN"))
	r.PUT("/admin/users/:id/role", h.UpdateRole)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/u1/role", bytes.NewBufferString(`{"role":"MANAGER"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

func withClaims(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middlewares.ContextClaimsKey, &jwtAuth.Claims{
			Role: role,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "subject-1",
			},
		})
		c.Next()
	}
}
