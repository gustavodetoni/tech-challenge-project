package routes_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	adminApp "github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
	jwtAuth "github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/routes"
	memrepo "github.com/soat-architecture/tech-challenge-project/internal/tests/infra/memory/repositories"
)

func TestAuthAndAdminRolePromotion_Flow(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	store := memrepo.NewStore()
	userRepo := memrepo.NewUserRepository(store)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authMW := middlewares.NewAuthMiddleware(jwtManager)

	authService := appAuth.NewAuthUseCase(userRepo, jwtManager)
	userSvc := adminApp.NewUserAdminUseCase(userRepo)

	router := gin.New()
	router.Use(gin.Recovery())
	routes.Register(router, routes.Deps{
		Health:     controllers.NewHealthController(),
		Auth:       controllers.NewAuthController(authService, userRepo),
		AuthMW:     authMW,
		AdminUsers: controllers.NewAdminUsersController(userSvc),
	})

	email := "john@example.com"
	password := "Senha@123"

	// Register
	var reg map[string]any
	doJSON(t, router, http.MethodPost, "/auth/register", "", map[string]any{
		"name":     "John",
		"email":    email,
		"password": password,
	}, http.StatusCreated, &reg)
	require.Equal(t, "VIEWER", reg["role"])

	// Login
	var login map[string]any
	doJSON(t, router, http.MethodPost, "/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	}, http.StatusOK, &login)
	accessToken := login["access_token"].(string)

	// Me
	var me map[string]any
	doJSON(t, router, http.MethodGet, "/me", accessToken, nil, http.StatusOK, &me)
	userID := me["id"].(string)
	require.Equal(t, "VIEWER", me["role"])

	// Promote via admin endpoint (using an admin token)
	adminToken, _, err := jwtManager.NewToken("admin-1", "ADMIN")
	require.NoError(t, err)
	doJSON(t, router, http.MethodPut, "/admin/users/"+userID+"/role", adminToken, map[string]any{
		"role": "MANAGER",
	}, http.StatusNoContent, nil)

	// Login again -> token role updated
	var login2 map[string]any
	doJSON(t, router, http.MethodPost, "/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	}, http.StatusOK, &login2)
	require.Equal(t, "MANAGER", login2["role"])
}
