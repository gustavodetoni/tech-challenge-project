package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	jwtAuth "github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAuthController_Register_BadRequest_InvalidBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_Conflict_EmailInUse(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(repository.ErrConflict).Once()

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"name":"Maria","email":"maria@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	userRepo.AssertExpectations(t)
}

func TestAuthController_Register_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil).Once()

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"name":"Maria","email":"maria@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	require.NotEmpty(t, out["access_token"])
	assert.Equal(t, "Bearer", out["token_type"])
	userRepo.AssertExpectations(t)
}

func TestAuthController_Register_BadRequest_InvalidInput(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"name":"Maria","email":"bad","password":"123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(assert.AnError).Once()

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"name":"Maria","email":"maria@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	userRepo.AssertExpectations(t)
}

func TestAuthController_Login_Unauthorized_InvalidCredentials(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	userRepo.On("FindByEmail", mock.Anything, "maria@example.com").Return((*user.User)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"maria@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	userRepo.AssertExpectations(t)
}

func TestAuthController_Login_BadRequest_InvalidInput(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"bad","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Login_InternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	userRepo.On("FindByEmail", mock.Anything, "maria@example.com").Return((*user.User)(nil), assert.AnError).Once()

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"maria@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	userRepo.AssertExpectations(t)
}

func TestAuthController_Login_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userRepo.On("FindByEmail", mock.Anything, "maria@example.com").Return(&user.User{
		ID:           "u1",
		Name:         "Maria",
		Email:        "maria@example.com",
		PasswordHash: string(hash),
		Role:         user.RoleViewer,
	}, nil).Once()

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"maria@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	require.NotEmpty(t, out["access_token"])
	userRepo.AssertExpectations(t)
}

func TestAuthController_Me_Unauthorized_NoClaims(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	r := gin.New()
	r.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_Me_Success(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	userRepo.On("FindByID", mock.Anything, "subject-1").Return(&user.User{
		ID:    "subject-1",
		Name:  "Maria",
		Email: "maria@example.com",
		Role:  user.RoleViewer,
	}, nil).Once()

	r := gin.New()
	r.Use(withClaims("VIEWER"))
	r.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"subject-1"`)
	userRepo.AssertExpectations(t)
}

func TestAuthController_Me_Unauthorized_WhenUserNotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	userRepo := new(repomocks.UserRepository)
	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authSvc := appAuth.NewAuthUseCase(userRepo, jwtManager)
	h := controllers.NewAuthController(authSvc, userRepo)

	userRepo.On("FindByID", mock.Anything, "subject-1").Return((*user.User)(nil), repository.ErrNotFound).Once()

	r := gin.New()
	r.Use(withClaims("VIEWER"))
	r.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	userRepo.AssertExpectations(t)
}
