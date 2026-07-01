package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	domainUser "github.com/soat-architecture/tech-challenge-project/internal/domain/user"
	jwtAuth "github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestAuthService_Register_Success(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *domainUser.User) bool {
		return u != nil &&
			u.Role == domainUser.RoleViewer &&
			u.Email == "john.doe@example.com" &&
			u.Name == "John Doe" &&
			u.PasswordHash != "" &&
			u.ID != ""
	})).Return(nil).Once()

	out, err := svc.Register(context.Background(), appAuth.RegisterInput{
		Name:     "  John Doe ",
		Email:    "  JOHN.DOE@EXAMPLE.COM ",
		Password: "Senha@123",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotEmpty(t, out.AccessToken)
	require.NotEmpty(t, out.UserID)
	assert.Equal(t, domainUser.RoleViewer, out.Role)

	claims, err := jwtManager.ParseAndValidate(out.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, out.UserID, claims.Subject)
	assert.Equal(t, string(domainUser.RoleViewer), claims.Role)
	assert.Equal(t, "issuer", claims.Issuer)
	require.NotNil(t, claims.ExpiresAt)
	assert.True(t, claims.ExpiresAt.After(time.Now().UTC()))

	repo.AssertExpectations(t)
}

func TestAuthService_Register_EmailInUse(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	repo.On("Create", mock.Anything, mock.Anything).Return(repository.ErrConflict).Once()

	_, err := svc.Register(context.Background(), appAuth.RegisterInput{
		Name:     "John",
		Email:    "john@example.com",
		Password: "Senha@123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, appAuth.ErrEmailInUse)
	repo.AssertExpectations(t)
}

func TestAuthService_Register_InvalidInput(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	_, err := svc.Register(context.Background(), appAuth.RegisterInput{
		Name:     "  ",
		Email:    "not-an-email",
		Password: "123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, appAuth.ErrInvalidInput)
}

func TestAuthService_Register_InvalidInput_EmptyEmail(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	_, err := svc.Register(context.Background(), appAuth.RegisterInput{
		Name:     "John",
		Email:    "   ",
		Password: "Senha@123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, appAuth.ErrInvalidInput)
}

func TestAuthService_Register_TokenIssueError(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

	_, err := svc.Register(context.Background(), appAuth.RegisterInput{
		Name:     "John",
		Email:    "john@example.com",
		Password: "Senha@123",
	})
	require.Error(t, err)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	pwHash, err := bcrypt.GenerateFromPassword([]byte("Senha@123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	u := &domainUser.User{
		ID:           "user-1",
		Name:         "John",
		Email:        "john@example.com",
		PasswordHash: string(pwHash),
		Role:         domainUser.RoleManager,
	}

	repo.On("FindByEmail", mock.Anything, "john@example.com").Return(u, nil).Once()

	out, err := svc.Login(context.Background(), appAuth.LoginInput{
		Email:    " JOHN@EXAMPLE.COM ",
		Password: "Senha@123",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "user-1", out.UserID)
	assert.Equal(t, domainUser.RoleManager, out.Role)

	claims, err := jwtManager.ParseAndValidate(out.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.Subject)
	assert.Equal(t, "MANAGER", claims.Role)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials_NotFound(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	repo.On("FindByEmail", mock.Anything, "john@example.com").Return((*domainUser.User)(nil), repository.ErrNotFound).Once()

	_, err := svc.Login(context.Background(), appAuth.LoginInput{
		Email:    "john@example.com",
		Password: "Senha@123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, appAuth.ErrInvalidCredentials)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials_WrongPassword(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	pwHash, err := bcrypt.GenerateFromPassword([]byte("Senha@123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	u := &domainUser.User{
		ID:           "user-1",
		Name:         "John",
		Email:        "john@example.com",
		PasswordHash: string(pwHash),
		Role:         domainUser.RoleViewer,
	}

	repo.On("FindByEmail", mock.Anything, "john@example.com").Return(u, nil).Once()

	_, err = svc.Login(context.Background(), appAuth.LoginInput{
		Email:    "john@example.com",
		Password: "wrong",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, appAuth.ErrInvalidCredentials)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_TokenIssueError(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("", "issuer", "", 30*time.Minute)
	repo := new(repomocks.UserRepository)
	svc := appAuth.NewAuthUseCase(repo, jwtManager)

	pwHash, err := bcrypt.GenerateFromPassword([]byte("Senha@123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	u := &domainUser.User{
		ID:           "user-1",
		Name:         "John",
		Email:        "john@example.com",
		PasswordHash: string(pwHash),
		Role:         domainUser.RoleManager,
	}
	repo.On("FindByEmail", mock.Anything, "john@example.com").Return(u, nil).Once()

	_, err = svc.Login(context.Background(), appAuth.LoginInput{
		Email:    "john@example.com",
		Password: "Senha@123",
	})
	require.Error(t, err)
	repo.AssertExpectations(t)
}

func TestJWTManager_ParseAndValidate_RejectsWrongSecret(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	token, _, err := jwtManager.NewToken("sub", "ADMIN")
	require.NoError(t, err)

	other := jwtAuth.NewManager("other-secret", "issuer", "", 30*time.Minute)
	_, err = other.ParseAndValidate(token)
	require.Error(t, err)
}

func TestJWTClaims_AudienceOptional(t *testing.T) {
	t.Parallel()

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 30*time.Minute)
	token, _, err := jwtManager.NewToken("sub", "ADMIN")
	require.NoError(t, err)

	claims, err := jwtManager.ParseAndValidate(token)
	require.NoError(t, err)
	// When audience is empty, it should not fail validation.
	assert.Len(t, claims.Audience, 0)
}
