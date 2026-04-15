package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtAuth "github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
)

func TestAuthMiddleware_RequireAuth_UnauthorizedMissingToken(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", time.Minute)
	mw := middlewares.NewAuthMiddleware(jwtManager)

	r := gin.New()
	r.GET("/me", mw.RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RequireRoles_ForbiddenWhenRoleNotAllowed(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", time.Minute)
	mw := middlewares.NewAuthMiddleware(jwtManager)

	token, _, err := jwtManager.NewToken("sub", "VIEWER")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/admin", mw.RequireRoles("ADMIN"), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthMiddleware_RequireRoles_Allows(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", time.Minute)
	mw := middlewares.NewAuthMiddleware(jwtManager)

	token, _, err := jwtManager.NewToken("sub", "ADMIN")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/admin", mw.RequireRoles("ADMIN", "MANAGER"), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RequireAuth_AllowsRawTokenHeader(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", time.Minute)
	mw := middlewares.NewAuthMiddleware(jwtManager)

	token, _, err := jwtManager.NewToken("sub", "VIEWER")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/me", mw.RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", token) // no Bearer prefix
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_RequireRoles_EmptyRolesAllowsAnyRole(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", time.Minute)
	mw := middlewares.NewAuthMiddleware(jwtManager)

	token, _, err := jwtManager.NewToken("sub", "VIEWER")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/x", mw.RequireRoles(""), func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
